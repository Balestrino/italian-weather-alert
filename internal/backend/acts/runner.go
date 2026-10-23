package acts

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/Balestrino/italian-weather-alert/internal/backend/acquisition"
	"github.com/Balestrino/italian-weather-alert/internal/backend/classification"
	"github.com/Balestrino/italian-weather-alert/internal/backend/documents"
	"github.com/Balestrino/italian-weather-alert/internal/backend/inference"
	"github.com/Balestrino/italian-weather-alert/internal/backend/jobs"
	"github.com/Balestrino/italian-weather-alert/internal/backend/ocr"
	"github.com/Balestrino/italian-weather-alert/internal/backend/registry"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	KindResolve = "resolve_act_references"
	KindAcquire = "acquire_albo_act"
	KindClarify = "clarify_act_measures"
)

// Payload carries the citing notice version and its projected extraction run
// so the interpretation guard applies the notice's territory, archive and
// suspension controls to every referenced-act job.
type Payload struct {
	DocumentVersionID int64  `json:"document_version_id"`
	ExtractionRunID   int64  `json:"extraction_run_id"`
	Workload          string `json:"workload"`
	ReferenceID       int64  `json:"reference_id,omitempty"`
	ResolutionID      int64  `json:"resolution_id,omitempty"`
}

// AlboClient is the albo access an acquisition needs beyond resolution.
type AlboClient interface {
	Albo
	Download(ctx context.Context, raw string) ([]byte, string, error)
}

type documentStore interface {
	Version(context.Context, int64) (documents.Version, error)
	Read(context.Context, int64, string) ([]byte, error)
	Retain(context.Context, documents.Acquisition) (documents.Version, error)
}

type ocrStore interface {
	PagesForVersion(context.Context, int64) ([]ocr.PageResult, error)
	ResourcesForVersion(context.Context, int64) ([]ocr.ResourceResult, error)
}

type Runner struct {
	Pool      *pgxpool.Pool
	Store     *Store
	Documents documentStore
	OCR       ocrStore
	Queue     *jobs.Store
	Policy    inference.RetryPolicy
	NewAlbo   func(*registry.JCityGovAlboContract) AlboClient
	Now       func() time.Time
}

func (r *Runner) now() time.Time {
	if r.Now != nil {
		return r.Now().UTC()
	}
	return time.Now().UTC()
}

func failure(code string, temporary bool) error {
	return &jobs.HandlerError{Failure: jobs.Failure{Code: code, Detail: "referenced act processing failed", Temporary: temporary}}
}

// Enqueue schedules one kind of referenced-act job. Keys make every job
// idempotent per notice version, reference and resolution.
func Enqueue(ctx context.Context, queue *jobs.Store, policy inference.RetryPolicy, kind string, p Payload, at time.Time) (jobs.Job, error) {
	if queue == nil || p.DocumentVersionID < 1 || p.ExtractionRunID < 1 || at.IsZero() || (kind != KindResolve && kind != KindAcquire && kind != KindClarify) {
		return jobs.Job{}, ErrInvalid
	}
	if p.Workload == "" {
		p.Workload = "ordinary"
	}
	body, _ := json.Marshal(p)
	key := fmt.Sprintf("%s:%d:%d:%d:%d", kind, p.DocumentVersionID, p.ExtractionRunID, p.ReferenceID, p.ResolutionID)
	request, err := policy.JobRequest(kind, key, body, at)
	if err != nil {
		return jobs.Job{}, err
	}
	return queue.Enqueue(ctx, request)
}

// AfterProjection is the interpretation scheduler hook: a projection that
// published measures schedules reference resolution for its notice version.
func (r *Runner) AfterProjection(ctx context.Context, extractionRunID int64, projected int, workload string, at time.Time) error {
	if projected < 1 || workload == "evaluation" {
		return nil
	}
	var version int64
	if err := r.Pool.QueryRow(ctx, `SELECT document_version_id FROM extraction_results WHERE run_id=$1`, extractionRunID).Scan(&version); err != nil {
		return err
	}
	_, err := Enqueue(ctx, r.Queue, r.Policy, KindResolve, Payload{DocumentVersionID: version, ExtractionRunID: extractionRunID, Workload: workload}, at)
	return err
}

func (r *Runner) Handlers() map[string]jobs.Handler {
	handle := func(run func(context.Context, Payload) (jobs.Result, error)) jobs.Handler {
		return func(ctx context.Context, job jobs.Job) (jobs.Result, error) {
			var p Payload
			if json.Unmarshal(job.Payload, &p) != nil || p.DocumentVersionID < 1 || p.ExtractionRunID < 1 {
				return jobs.Result{}, failure("invalid_referenced_act_job", false)
			}
			return run(ctx, p)
		}
	}
	return map[string]jobs.Handler{KindResolve: handle(r.resolve), KindAcquire: handle(r.acquire), KindClarify: handle(r.clarify)}
}

// albo returns the enabled albo source serving the notice's source, if any.
func (r *Runner) albo(ctx context.Context, primary string) (string, int, *registry.JCityGovAlboContract, error) {
	rows, err := r.Pool.Query(ctx, `SELECT s.id,s.active_revision,c.body FROM registry_sources s
 JOIN registry_configurations c ON c.source_id=s.id AND c.revision=s.active_revision
 WHERE s.collection_enabled AND c.body->>'access_method'=$1 AND c.body->'jcitygov_albo'->>'serves_source'=$2 ORDER BY s.id`, registry.JCityGovAlboAccess, primary)
	if err != nil {
		return "", 0, nil, err
	}
	defer rows.Close()
	var found []string
	var revision int
	var contract *registry.JCityGovAlboContract
	for rows.Next() {
		var id string
		var body []byte
		var cfg registry.Configuration
		if err = rows.Scan(&id, &revision, &body); err != nil {
			return "", 0, nil, err
		}
		if err = json.Unmarshal(body, &cfg); err != nil || cfg.JCityGovAlbo == nil {
			return "", 0, nil, ErrInvalid
		}
		found, contract = append(found, id), cfg.JCityGovAlbo
	}
	if err = rows.Err(); err != nil {
		return "", 0, nil, err
	}
	switch len(found) {
	case 0:
		return "", 0, nil, nil
	case 1:
		return found[0], revision, contract, nil
	default:
		return "", 0, nil, fmt.Errorf("%w: several albo sources serve %s", ErrInvalid, primary)
	}
}

func (r *Runner) notice(ctx context.Context, p Payload) (documents.Version, Notice, error) {
	version, err := r.Documents.Version(ctx, p.DocumentVersionID)
	if err != nil {
		return version, Notice{}, err
	}
	content, err := classification.GatherContent(ctx, r.Documents, r.OCR, version)
	if err != nil {
		return version, Notice{}, err
	}
	n := Notice{Content: content}
	var metadata struct {
		Date string `json:"source_publication_date"`
	}
	if json.Unmarshal(version.Metadata, &metadata) == nil && metadata.Date != "" {
		if date, parseErr := time.Parse(time.DateOnly, metadata.Date); parseErr == nil {
			n.PublicationDate = date
		}
	}
	rows, err := r.Pool.Query(ctx, `SELECT m.id,m.kind,m.subject FROM domain_local_measures m JOIN domain_measure_bindings b ON b.local_measure_id=m.id
 WHERE b.extraction_run_id=$1 AND m.document_version_id=$2 ORDER BY b.extraction_ordinal`, p.ExtractionRunID, p.DocumentVersionID)
	if err != nil {
		return version, n, err
	}
	defer rows.Close()
	for rows.Next() {
		var m NoticeMeasure
		if err = rows.Scan(&m.ID, &m.Kind, &m.Subject); err != nil {
			return version, n, err
		}
		n.Measures = append(n.Measures, m)
	}
	return version, n, rows.Err()
}

func noticeSource(version documents.Version) string {
	for _, resource := range version.Resources {
		if resource.Role == "original" {
			return resource.SourceID
		}
	}
	return ""
}

// resolve detects the notice's references and resolves each one that has no
// final outcome yet. An unavailable albo leaves a recorded outcome and
// retries the job; other references keep their outcomes.
func (r *Runner) resolve(ctx context.Context, p Payload) (jobs.Result, error) {
	version, notice, err := r.notice(ctx, p)
	if err != nil {
		return jobs.Result{}, failure("notice_unavailable", true)
	}
	if len(notice.Measures) == 0 {
		return jobs.JSONResult(json.RawMessage(`{"status":"no_projected_measures"}`))
	}
	at := r.now()
	refs, err := r.Store.RecordReferences(ctx, version.ID, Detect(notice.Content), at)
	if err != nil {
		return jobs.Result{}, failure("act_references_unavailable", true)
	}
	alboSource, _, contract, err := r.albo(ctx, noticeSource(version))
	if err != nil {
		return jobs.Result{}, failure("albo_configuration_invalid", false)
	}
	var client AlboClient
	if contract != nil && r.NewAlbo != nil {
		client = r.NewAlbo(contract)
	}
	pending := false
	for _, ref := range refs {
		latest, found, err := r.Store.LatestResolution(ctx, ref.ID, at)
		if err != nil {
			return jobs.Result{}, failure("act_resolution_unavailable", true)
		}
		if found && latest.Outcome != "unavailable" {
			continue
		}
		var albo Albo
		if client != nil {
			albo = client
		}
		resolution := Resolve(ctx, ref.Reference, notice, contract, albo)
		id, err := r.Store.RecordResolution(ctx, ref.ID, alboSource, resolution, r.now())
		if err != nil {
			return jobs.Result{}, failure("act_resolution_unavailable", true)
		}
		next := Payload{DocumentVersionID: p.DocumentVersionID, ExtractionRunID: p.ExtractionRunID, Workload: p.Workload, ReferenceID: ref.ID, ResolutionID: id}
		switch resolution.Outcome {
		case "attached":
			_, err = Enqueue(ctx, r.Queue, r.Policy, KindClarify, next, r.now())
		case "numbered", "unique_candidate":
			_, err = Enqueue(ctx, r.Queue, r.Policy, KindAcquire, next, r.now())
		case "unavailable":
			pending = true
		}
		if err != nil {
			return jobs.Result{}, failure("act_job_unavailable", true)
		}
	}
	if pending {
		return jobs.Result{}, failure("albo_unavailable", true)
	}
	return jobs.JSONResult(json.RawMessage(fmt.Sprintf(`{"status":"resolved","references":%d}`, len(refs))))
}

// acquire retains the resolved albo record once: its detail page and the
// unsigned main document. Only OCR runs on the retained act.
func (r *Runner) acquire(ctx context.Context, p Payload) (jobs.Result, error) {
	var alboSource, record string
	if err := r.Pool.QueryRow(ctx, `SELECT albo_source_id,albo_record_id FROM act_resolutions WHERE id=$1 AND outcome IN ('numbered','unique_candidate')`, p.ResolutionID).Scan(&alboSource, &record); err != nil {
		return jobs.Result{}, failure("act_resolution_unavailable", false)
	}
	var actVersion int64
	err := r.Pool.QueryRow(ctx, `SELECT document_version_id FROM act_albo_records WHERE source_id=$1 AND record_id=$2`, alboSource, record).Scan(&actVersion)
	if errors.Is(err, pgx.ErrNoRows) {
		actVersion, err = r.retainAct(ctx, p, alboSource, record)
	}
	if err != nil {
		var handler *jobs.HandlerError
		if errors.As(err, &handler) {
			return jobs.Result{}, err
		}
		return jobs.Result{}, failure("albo_act_unavailable", true)
	}
	if err = r.enqueueOCR(ctx, actVersion, p.Workload); err != nil {
		return jobs.Result{}, failure("act_job_unavailable", true)
	}
	if _, err = Enqueue(ctx, r.Queue, r.Policy, KindClarify, p, r.now()); err != nil {
		return jobs.Result{}, failure("act_job_unavailable", true)
	}
	return jobs.JSONResult(json.RawMessage(fmt.Sprintf(`{"status":"acquired","act_document_version_id":%d}`, actVersion)))
}

func (r *Runner) retainAct(ctx context.Context, p Payload, alboSource, record string) (int64, error) {
	var revision int
	var body []byte
	if err := r.Pool.QueryRow(ctx, `SELECT s.active_revision,c.body FROM registry_sources s JOIN registry_configurations c ON c.source_id=s.id AND c.revision=s.active_revision
 WHERE s.id=$1 AND s.collection_enabled`, alboSource).Scan(&revision, &body); err != nil {
		return 0, failure("albo_source_disabled", false)
	}
	var cfg registry.Configuration
	if json.Unmarshal(body, &cfg) != nil || cfg.JCityGovAlbo == nil || r.NewAlbo == nil {
		return 0, failure("albo_configuration_invalid", false)
	}
	client := r.NewAlbo(cfg.JCityGovAlbo)
	detail, markup, err := client.Detail(ctx, record)
	if err != nil {
		return 0, failure("albo_detail_unavailable", true)
	}
	var main *acquisition.AlboAttachment
	for i, attachment := range detail.Attachments {
		if attachment.Description == "Documento principale" && attachment.UnsignedURL != "" {
			if main != nil {
				return 0, failure("albo_main_document_ambiguous", false)
			}
			main = &detail.Attachments[i]
		}
	}
	if main == nil {
		return 0, failure("albo_main_document_missing", false)
	}
	pdf, media, err := client.Download(ctx, main.UnsignedURL)
	if err != nil {
		return 0, failure("albo_document_unavailable", true)
	}
	if media != "application/pdf" || !strings.HasPrefix(string(pdf), "%PDF") {
		return 0, failure("albo_document_not_pdf", false)
	}
	fields := map[string]any{"albo_record_id": record, "category": detail.Category, "subcategory": detail.Subcategory, "act_number": detail.ActNumber, "act_year": detail.ActYear,
		"registry": fmt.Sprintf("%d/%d", detail.RegistryYear, detail.RegistryNumber), "object": detail.Object}
	if !detail.ActDate.IsZero() {
		fields["act_date"] = detail.ActDate.Format(time.DateOnly)
	}
	metadata, _ := json.Marshal(fields)
	version, err := r.Documents.Retain(ctx, documents.Acquisition{ID: "albo:" + alboSource + ":" + record, SourceID: alboSource, Configuration: revision, URL: detail.DetailURL, Metadata: metadata, Resources: []documents.Resource{
		{URL: detail.DetailURL, Role: "original", Required: true, SourceID: alboSource, Configuration: revision, MediaType: "text/html", Bytes: markup},
		{URL: main.UnsignedURL, Role: "attachment", Required: true, SourceID: alboSource, Configuration: revision, MediaType: "application/pdf", Bytes: pdf},
	}})
	if err != nil {
		return 0, failure("albo_act_retention_failed", true)
	}
	var actDate any
	if !detail.ActDate.IsZero() {
		actDate = detail.ActDate
	}
	var number, year any
	if detail.ActNumber > 0 {
		number, year = detail.ActNumber, detail.ActYear
	}
	if _, err = r.Pool.Exec(ctx, `INSERT INTO act_albo_records(source_id,record_id,category,subcategory,act_number,act_year,act_date,registry_year,registry_number,object,detail_url,document_version_id,recorded_at)
 VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13) ON CONFLICT(source_id,record_id) DO NOTHING`,
		alboSource, record, detail.Category, detail.Subcategory, number, year, actDate, detail.RegistryYear, detail.RegistryNumber, detail.Object, detail.DetailURL, version.ID, r.now()); err != nil {
		return 0, err
	}
	var stored int64
	err = r.Pool.QueryRow(ctx, `SELECT document_version_id FROM act_albo_records WHERE source_id=$1 AND record_id=$2`, alboSource, record).Scan(&stored)
	return stored, err
}

func (r *Runner) enqueueOCR(ctx context.Context, versionID int64, workload string) error {
	version, err := r.Documents.Version(ctx, versionID)
	if err != nil {
		return err
	}
	for _, resource := range version.Resources {
		if resource.Role == "attachment" && strings.HasPrefix(resource.MediaType, "application/pdf") && resource.Missing == "" {
			if _, err = ocr.Enqueue(ctx, r.Queue, r.Policy, ocr.Payload{DocumentVersionID: versionID, ResourceURL: resource.URL, Workload: workload}, r.now()); err != nil {
				return err
			}
		}
	}
	return nil
}

// clarify reads the resolved act's dispositive clause and appends clarified
// validity to the citing notice's measures whose own validity is unknown.
func (r *Runner) clarify(ctx context.Context, p Payload) (jobs.Result, error) {
	var outcome, attached string
	var actVersion int64
	err := r.Pool.QueryRow(ctx, `SELECT r.outcome,COALESCE(r.attached_resource_url,''),COALESCE(a.document_version_id,f.document_version_id)
 FROM act_resolutions r JOIN act_references f ON f.id=r.reference_id
 LEFT JOIN act_albo_records a ON a.source_id=r.albo_source_id AND a.record_id=r.albo_record_id
 WHERE r.id=$1 AND (r.outcome='attached' OR a.id IS NOT NULL)`, p.ResolutionID).Scan(&outcome, &attached, &actVersion)
	if err != nil {
		return jobs.Result{}, failure("act_resolution_unavailable", false)
	}
	version, err := r.Documents.Version(ctx, actVersion)
	if err != nil {
		return jobs.Result{}, failure("act_unavailable", true)
	}
	content, err := classification.GatherContent(ctx, r.Documents, r.OCR, version)
	if err != nil {
		return jobs.Result{}, failure("act_content_unavailable", true)
	}
	var sections []classification.ContentSection
	for _, section := range content.Sections {
		if section.Role == "attachment" && (attached == "" || section.ResourceURL == attached) {
			sections = append(sections, section)
		}
	}
	if !content.Complete || len(sections) == 0 {
		// OCR of a newly retained act completes asynchronously.
		if r.now().Sub(version.FirstAcquiredAt) < 24*time.Hour {
			return jobs.Result{}, &jobs.DeferredError{Until: r.now().Add(2 * time.Minute), Code: "act_content_pending"}
		}
		return r.recordUnavailable(ctx, p, actVersion, "act_content_unavailable")
	}
	clause, err := ParseClause(sections)
	if err != nil {
		return r.recordUnavailable(ctx, p, actVersion, "clause_not_recognized")
	}
	_, notice, err := r.notice(ctx, p)
	if err != nil {
		return jobs.Result{}, failure("notice_unavailable", true)
	}
	label, err := r.actLabel(ctx, p.ResolutionID, sections)
	if err != nil {
		return jobs.Result{}, failure("act_resolution_unavailable", true)
	}
	clarified := 0
	for _, m := range notice.Measures {
		done, err := r.Store.clarifyMeasure(ctx, p, m, actVersion, clause, label, r.now())
		if err != nil {
			return jobs.Result{}, failure("act_clarification_unavailable", true)
		}
		if done {
			clarified++
		}
	}
	return jobs.JSONResult(json.RawMessage(fmt.Sprintf(`{"status":"clarified","measures":%d,"act_document_version_id":%d}`, clarified, actVersion)))
}

func (r *Runner) recordUnavailable(ctx context.Context, p Payload, actVersion int64, reason string) (jobs.Result, error) {
	if _, err := r.Pool.Exec(ctx, `INSERT INTO act_clarifications(resolution_id,act_document_version_id,outcome,reason,parser_version,recorded_at)
 VALUES($1,$2,'clarification_unavailable',$3,$4,$5) ON CONFLICT DO NOTHING`, p.ResolutionID, actVersion, reason, ParserVersion, r.now()); err != nil {
		return jobs.Result{}, failure("act_clarification_unavailable", true)
	}
	return jobs.JSONResult(json.RawMessage(fmt.Sprintf(`{"status":"clarification_unavailable","reason":%q}`, reason)))
}

// actLabel names the act in assumptions and locators, e.g. "ordinanza 38/2026".
func (r *Runner) actLabel(ctx context.Context, resolutionID int64, sections []classification.ContentSection) (string, error) {
	var kind string
	var number, year int
	err := r.Pool.QueryRow(ctx, `SELECT f.kind,COALESCE(a.act_number,f.act_number,0),COALESCE(a.act_year,f.act_year,0) FROM act_resolutions r JOIN act_references f ON f.id=r.reference_id
 LEFT JOIN act_albo_records a ON a.source_id=r.albo_source_id AND a.record_id=r.albo_record_id WHERE r.id=$1`, resolutionID).Scan(&kind, &number, &year)
	if err != nil {
		return "", err
	}
	if number == 0 {
		for _, section := range sections {
			if m := attachedAct.FindStringSubmatch(section.Text); m != nil && canonicalKind(m[1]) == kind {
				return kind + " " + strings.TrimLeft(m[2], "0") + "/" + m[3], nil
			}
		}
		return kind, nil
	}
	return fmt.Sprintf("%s %d/%d", kind, number, year), nil
}
