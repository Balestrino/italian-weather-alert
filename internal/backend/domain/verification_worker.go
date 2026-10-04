package domain

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/Balestrino/italian-weather-alert/internal/backend/documents"
	"github.com/Balestrino/italian-weather-alert/internal/backend/extraction"
	"github.com/Balestrino/italian-weather-alert/internal/backend/registry"
	"github.com/jackc/pgx/v5"
)

// VerificationWorker revisits retained candidates independently of public queries
// and inference jobs. Source/territory switches remain the admission boundary.
// An unchanged evidence state has one immutable receipt; later primary evidence,
// interpretation, acquisition failures or recovery create a new receipt.
type VerificationWorker struct {
	Store     *Store
	Documents RegionalDocuments
	Interval  time.Duration
}

func (w *VerificationWorker) Run(ctx context.Context) error {
	if w.Store == nil || w.Documents == nil || w.Interval < time.Second {
		return ErrInvalid
	}
	for {
		workCtx, cancel := context.WithTimeout(ctx, 2*time.Minute)
		count, err := w.RunOnce(workCtx, time.Now().UTC())
		cancel()
		if ctx.Err() != nil {
			return nil
		}
		if err != nil {
			slog.Warn("automatic verification failed", "component", "verification", "error", err)
		}
		if count > 0 {
			slog.Info("automatic verification recorded", "component", "verification", "receipts", count)
		}
		timer := time.NewTimer(w.Interval)
		select {
		case <-ctx.Done():
			timer.Stop()
			return nil
		case <-timer.C:
		}
	}
}

type verificationSourceState struct {
	ID, Territory, Product, Platform string
	Revision                         int
	Configuration                    registry.Configuration
	Healthy                          bool
	Enabled                          bool
	CheckedAt                        time.Time
}

type verificationDocument struct {
	Version    documents.Version
	URL        string
	Source     verificationSourceState
	Texts      []verificationLiteral
	Extraction int64
}
type verificationLiteral struct {
	Text      string
	Selection EvidenceSelection
}

func (w *VerificationWorker) RunOnce(ctx context.Context, at time.Time) (int, error) {
	if w.Store == nil || w.Store.pool == nil || w.Documents == nil || at.IsZero() {
		return 0, ErrInvalid
	}
	connection, err := w.Store.pool.Acquire(ctx)
	if err != nil {
		return 0, err
	}
	defer connection.Release()
	var locked bool
	if err = connection.QueryRow(ctx, "SELECT pg_try_advisory_lock(730073,1)").Scan(&locked); err != nil || !locked {
		return 0, err
	}
	defer func() {
		unlockCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if _, err := connection.Exec(unlockCtx, "SELECT pg_advisory_unlock(730073,1)"); err != nil {
			_ = connection.Conn().Close(unlockCtx)
		}
	}()
	sources, err := w.sources(ctx, at)
	if err != nil {
		return 0, err
	}
	count := 0
	for _, platform := range sources {
		if platform.Configuration.CittadinoInformato == nil || !platform.Enabled {
			continue
		}
		candidates, err := w.documents(ctx, platform, at)
		if err != nil {
			return count, err
		}
		var municipal, regional []verificationDocument
		for _, primary := range sources {
			if primary.Configuration.CittadinoInformato != nil || !primary.Enabled {
				continue
			}
			if primary.Product == "municipal" && primary.Territory != platform.Territory {
				continue
			}
			if primary.Product != "municipal" && (primary.Product != ProductCriticality || !slices.Contains([]string{"09", "toscana", "regione toscana"}, strings.ToLower(primary.Territory))) {
				continue
			}
			docs, err := w.documents(ctx, primary, at)
			if err != nil {
				return count, err
			}
			if primary.Product == "municipal" {
				municipal = append(municipal, docs...)
			} else {
				regional = append(regional, docs...)
			}
		}
		for _, candidate := range candidates {
			var metadata struct {
				Kind string `json:"kind"`
			}
			if json.Unmarshal(candidate.Version.Metadata, &metadata) != nil {
				return count, ErrInvalid
			}
			if metadata.Kind == "listing" {
				continue
			}
			requests, err := w.requests(ctx, candidate, municipal, regional, sources, at)
			if err != nil {
				return count, err
			}
			for _, request := range requests {
				// Check times are evidence inspection times, not part of the stable
				// semantic state. Skip an already persisted state before calling the
				// strict request-id/input collision guard.
				identity := request
				identity.Checks = slices.Clone(request.Checks)
				for i := range identity.Checks {
					identity.Checks[i].CheckedAt = time.Time{}
				}
				prefix := "automatic-v1-" + verificationHash(identity) + ":"
				var previousID int64
				var previousRequest string
				err := w.Store.pool.QueryRow(ctx, "SELECT id,request_id FROM domain_verification_receipts WHERE candidate_source_id=$1 AND candidate_key=$2 ORDER BY id DESC LIMIT 1", request.Candidate.SourceID, request.CandidateKey).Scan(&previousID, &previousRequest)
				if err != nil && err != pgx.ErrNoRows {
					return count, err
				}
				if strings.HasPrefix(previousRequest, prefix) {
					continue
				}
				request.RequestID = prefix + strconv.FormatInt(previousID, 10)
				if _, err := w.Store.VerifyMultiSource(ctx, w.Documents, request, at); err != nil {
					return count, err
				}
				count++
			}
		}
	}
	return count, nil
}

func (w *VerificationWorker) sources(ctx context.Context, at time.Time) ([]verificationSourceState, error) {
	rows, err := w.Store.pool.Query(ctx, `SELECT s.id,s.territory,s.product_id,c.platform,COALESCE(s.active_revision,s.latest_revision),cfg.body,
 s.collection_enabled AND s.interpretation_suspended_at IS NULL AND territorial_source_allowed(s.id),
 COALESCE(a.last_complete_at>=$1::timestamptz-make_interval(secs=>a.delay_seconds) AND a.last_error_code IS NULL,false),COALESCE(a.last_complete_at,$1::timestamptz)
 FROM registry_sources s JOIN registry_channels c ON c.id=s.channel_id
 JOIN registry_configurations cfg ON cfg.source_id=s.id AND cfg.revision=COALESCE(s.active_revision,s.latest_revision)
 LEFT JOIN acquisition_source_status a ON a.source_id=s.id AND a.configuration=s.active_revision
 ORDER BY s.id`, at)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []verificationSourceState
	for rows.Next() {
		var source verificationSourceState
		var body []byte
		if err := rows.Scan(&source.ID, &source.Territory, &source.Product, &source.Platform, &source.Revision, &body, &source.Enabled, &source.Healthy, &source.CheckedAt); err != nil {
			return nil, err
		}
		if json.Unmarshal(body, &source.Configuration) != nil {
			return nil, ErrInvalid
		}
		out = append(out, source)
	}
	return out, rows.Err()
}

func (w *VerificationWorker) documents(ctx context.Context, source verificationSourceState, at time.Time) ([]verificationDocument, error) {
	rows, err := w.Store.pool.Query(ctx, `SELECT d.official_url,a.version_id,COALESCE((SELECT e.run_id FROM extraction_results e JOIN processing_runs p ON p.id=e.run_id WHERE e.document_version_id=a.version_id AND e.created_at<=$3 AND p.workload<>'evaluation' AND e.status='extracted' AND e.content_complete ORDER BY e.created_at DESC,e.run_id DESC LIMIT 1),0)
 FROM retained_documents d JOIN LATERAL (SELECT version_id FROM retained_acquisitions WHERE document_id=d.id AND configuration=$2 AND version_id IS NOT NULL AND acquired_at<=$3 ORDER BY acquired_at DESC,id DESC LIMIT 1) a ON true
 WHERE d.source_id=$1 AND (NOT $4 OR NOT EXISTS(SELECT 1 FROM interpretation_archives WHERE document_version_id=a.version_id))
 ORDER BY d.id LIMIT 10001`, source.ID, source.Revision, at, source.Configuration.CittadinoInformato != nil)
	if err != nil {
		return nil, err
	}
	var out []verificationDocument
	for rows.Next() {
		var d verificationDocument
		var id int64
		if err := rows.Scan(&d.URL, &id, &d.Extraction); err != nil {
			rows.Close()
			return nil, err
		}
		d.Source = source
		d.Version.ID = id
		out = append(out, d)
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return nil, err
	}
	if len(out) > 10000 {
		return nil, fmt.Errorf("automatic verification document limit reached")
	}
	for i := range out {
		v, err := w.Documents.Version(ctx, out[i].Version.ID)
		if err != nil {
			return nil, err
		}
		out[i].Version = v
		texts, err := w.literals(ctx, v, at)
		if err != nil {
			return nil, err
		}
		out[i].Texts = texts
	}
	return out, nil
}

func (w *VerificationWorker) literals(ctx context.Context, v documents.Version, at time.Time) ([]verificationLiteral, error) {
	var out []verificationLiteral
	for _, resource := range v.Resources {
		if resource.Missing != "" || !resource.InferenceEligible() {
			continue
		}
		selection := EvidenceSelection{ResourceURL: resource.URL, Locator: "automatic retained evidence"}
		if resource.MediaType == "application/pdf" || strings.HasPrefix(resource.MediaType, "image/") {
			rows, err := w.Store.pool.Query(ctx, `SELECT p.run_id,p.page_number,p.extracted_text FROM ocr_page_results p JOIN ocr_resource_results r USING(run_id)
 WHERE p.document_version_id=$1 AND p.resource_url=$2 AND p.status='complete' AND r.status='complete' AND p.created_at<=$3 AND r.created_at<=$3
 ORDER BY p.run_id DESC,p.page_number`, v.ID, resource.URL, at)
			if err != nil {
				return nil, err
			}
			seen := map[int]bool{}
			for rows.Next() {
				var run int64
				var page int
				var text string
				if err := rows.Scan(&run, &page, &text); err != nil {
					rows.Close()
					return nil, err
				}
				if seen[page] {
					continue
				}
				seen[page] = true
				s := selection
				s.OCRRunID = run
				s.Page = page
				out = append(out, verificationLiteral{cleanLiteral(text), s})
			}
			err = rows.Err()
			rows.Close()
			if err != nil {
				return nil, err
			}
			continue
		}
		body, err := w.Documents.Read(ctx, v.ID, resource.URL)
		if err != nil {
			return nil, err
		}
		if resource.MediaType == "application/json" {
			var node any
			if json.Unmarshal(body, &node) != nil {
				return nil, ErrInvalid
			}
			var walk func(any, string)
			walk = func(node any, pointer string) {
				switch n := node.(type) {
				case map[string]any:
					keys := make([]string, 0, len(n))
					for k := range n {
						keys = append(keys, k)
					}
					slices.Sort(keys)
					for _, k := range keys {
						walk(n[k], pointer+"/"+strings.ReplaceAll(strings.ReplaceAll(k, "~", "~0"), "/", "~1"))
					}
				case []any:
					for i, c := range n {
						walk(c, pointer+"/"+strconv.Itoa(i))
					}
				case string:
					if text, err := verificationText(resource.MediaType, body, pointer); err == nil && text != "" {
						s := selection
						s.JSONPointer = pointer
						out = append(out, verificationLiteral{text, s})
					}
				}
			}
			walk(node, "")
			continue
		}
		text, err := verificationText(resource.MediaType, body, "")
		if err != nil {
			continue
		}
		out = append(out, verificationLiteral{text, selection})
	}
	return out, nil
}

// A shared exact resource or an exact complete notice body identifies a
// counterpart to consult. It does not establish act edition or field agreement.
func counterpart(candidate, primary verificationDocument) bool {
	for _, a := range candidate.Version.Resources {
		for _, b := range primary.Version.Resources {
			if a.URL == b.URL && a.Missing == "" && b.Missing == "" {
				return true
			}
		}
	}
	for _, a := range candidate.Texts {
		if a.Selection.JSONPointer != "/contenuto" && a.Selection.JSONPointer != "/titolo" {
			continue
		}
		for _, b := range primary.Texts {
			if (a.Selection.JSONPointer == "/contenuto" && len(a.Text) > 80 || a.Selection.JSONPointer == "/titolo" && len(a.Text) > 20) && strings.Contains(b.Text, a.Text) {
				return true
			}
		}
	}
	return false
}

func (w *VerificationWorker) requests(ctx context.Context, candidate verificationDocument, municipal, regional []verificationDocument, sources []verificationSourceState, at time.Time) ([]VerificationRequest, error) {
	var metadata struct {
		Kind string `json:"kind"`
	}
	_ = json.Unmarshal(candidate.Version.Metadata, &metadata)
	base := VerificationRequest{ComparisonOnly: true, CandidateKey: "notice-" + verificationHash(candidate.URL), Kind: "local_measure", MunicipalityISTAT: candidate.Source.Territory, Candidate: VerificationEvidence{SourceID: candidate.Source.ID, VersionID: candidate.Version.ID, Fields: map[string]EvidenceSelection{}}}
	if metadata.Kind == "risk" || metadata.Kind == "risks" || metadata.Kind == "regional_republication" {
		var out []VerificationRequest
		for _, literal := range candidate.Texts {
			if !strings.HasPrefix(literal.Selection.JSONPointer, "/stati_allerta/") || !strings.HasSuffix(literal.Selection.JSONPointer, "/allerta") || !slices.Contains([]string{"green", "yellow", "orange", "red"}, normalizedField("level", literal.Text)) {
				continue
			}
			r := base
			r.Kind = "regional_record"
			r.CandidateKey = "risk-" + verificationHash(candidate.URL+literal.Selection.JSONPointer)
			r.Candidate.Fields = map[string]EvidenceSelection{"level": wholeLiteral(literal)}
			prefix := strings.Split(literal.Selection.JSONPointer, "/")
			riskPointer := "/stati_allerta/" + prefix[2] + "/tipologia"
			for _, risk := range candidate.Texts {
				if risk.Selection.JSONPointer == riskPointer && slices.Contains(Risks, normalizedField("risk", risk.Text)) {
					r.Candidate.Fields["risk"] = wholeLiteral(risk)
				}
			}
			r.Checks = []VerificationCheck{w.check("regional", candidate, regional, sources, at, true), {Role: "municipal", State: "not_applicable", Reason: "originating_regional_claim", CheckedAt: at}}
			out = append(out, r)
		}
		if len(out) > 0 {
			return out, nil
		}
		base.Kind = "regional_record"
		base.Checks = []VerificationCheck{w.check("regional", candidate, regional, sources, at, true), {Role: "municipal", State: "not_applicable", Reason: "originating_regional_claim", CheckedAt: at}}
		return []VerificationRequest{base}, nil
	}
	var measures []extraction.Measure
	if candidate.Extraction > 0 {
		r, found, err := extraction.NewStore(w.Store.pool).Get(ctx, candidate.Extraction)
		if err != nil {
			return nil, err
		}
		if found {
			measures = r.Measures
		}
	}
	if len(measures) == 0 {
		base.Checks = []VerificationCheck{w.check("municipal", candidate, municipal, sources, at, false), {Role: "regional", State: "not_applicable", Reason: "no_interpreted_regional_claim", CheckedAt: at}}
		return []VerificationRequest{base}, nil
	}
	var out []VerificationRequest
	for _, measure := range measures {
		r := base
		r.CandidateKey = fmt.Sprintf("%s-measure-%d", base.CandidateKey, measure.Ordinal)
		r.Candidate.Fields = measureSelectors(candidate, measure)
		r.Checks = []VerificationCheck{w.check("municipal", candidate, municipal, sources, at, false), {Role: "regional", State: "not_applicable", Reason: "local_measure_without_regional_claim", CheckedAt: at}}
		// Select a unique interpreted primary measure for the same subject/place;
		// a changed action must remain detectable as a conflict.
		check := &r.Checks[0]
		if check.State == "available" {
			for _, primary := range municipal {
				if primary.Version.ID != check.VersionID || primary.Extraction == 0 {
					continue
				}
				p, found, err := extraction.NewStore(w.Store.pool).Get(ctx, primary.Extraction)
				if err != nil {
					return nil, err
				}
				if !found {
					continue
				}
				var matched []extraction.Measure
				for _, m := range p.Measures {
					if cleanLiteral(m.Subject) == cleanLiteral(measure.Subject) && ((m.Place == nil && measure.Place == nil) || (m.Place != nil && measure.Place != nil && cleanLiteral(*m.Place) == cleanLiteral(*measure.Place))) {
						matched = append(matched, m)
					}
				}
				if len(matched) == 1 {
					check.Fields = measureSelectors(primary, matched[0])
				}
			}
		}
		out = append(out, r)
	}
	return out, nil
}

func (w *VerificationWorker) check(role string, candidate verificationDocument, documents []verificationDocument, sources []verificationSourceState, at time.Time, regional bool) VerificationCheck {
	check := VerificationCheck{Role: role, State: "unavailable", Reason: "primary_source_not_active", CheckedAt: at}
	var configured []verificationSourceState
	for _, s := range sources {
		if s.Configuration.CittadinoInformato != nil {
			continue
		}
		if !regional && s.Product == "municipal" && s.Territory == candidate.Source.Territory || regional && s.Product == ProductCriticality && slices.Contains([]string{"09", "toscana", "regione toscana"}, strings.ToLower(s.Territory)) {
			configured = append(configured, s)
		}
	}
	if len(configured) == 0 {
		return check
	}
	check.SourceID = configured[0].ID
	if len(configured) != 1 {
		check.Reason = "primary_source_ambiguous"
		return check
	}
	if !configured[0].Healthy || !configured[0].Enabled {
		check.Reason = "primary_check_incomplete_or_delayed"
		return check
	}
	check.State = "missing"
	check.Reason = "no_retained_counterpart"
	check.CheckedAt = at
	var matches []verificationDocument
	for _, primary := range documents {
		if primary.Source.ID == check.SourceID && (regional || counterpart(candidate, primary)) {
			matches = append(matches, primary)
		}
	}
	if regional && len(matches) > 0 {
		slices.SortFunc(matches, func(a, b verificationDocument) int {
			return b.Version.FirstAcquiredAt.Compare(a.Version.FirstAcquiredAt)
		})
		matches = matches[:1]
	}
	if len(matches) > 1 {
		check.State = "unavailable"
		check.Reason = "counterpart_ambiguous"
		return check
	}
	if len(matches) == 0 {
		return check
	}
	check.State = "available"
	check.Reason = "retained_counterpart_identity_not_inferred"
	check.VersionID = matches[0].Version.ID
	check.Fields = map[string]EvidenceSelection{}
	return check
}

func wholeLiteral(literal verificationLiteral) EvidenceSelection {
	s := literal.Selection
	s.StartByte = 0
	s.EndByte = len(literal.Text)
	return s
}

var automaticAct = regexp.MustCompile(`(?i)\b(?:ordinanza|ordinance)\s+(?:n\.?\s*)?[0-9]+(?:/[0-9]{4})?`)
var automaticEdition = regexp.MustCompile(`(?i)\b(?:ordinanza|ordinance)\s+(?:n\.?\s*)?[0-9]+(?:/[0-9]{4})?\s+del\s+([0-9]{2}/[0-9]{2}/[0-9]{4})`)

func measureSelectors(document verificationDocument, m extraction.Measure) map[string]EvidenceSelection {
	fields := map[string]EvidenceSelection{}
	values := map[string]string{"subject": m.Subject}
	if m.Place != nil {
		values["place"] = *m.Place
	}
	for _, evidence := range m.Evidence {
		if !slices.Contains([]string{"kind", "subject", "place"}, evidence.Field) {
			continue
		}
		for _, literal := range document.Texts {
			if literal.Selection.ResourceURL != evidence.ResourceURL || evidence.Page != nil && literal.Selection.Page != *evidence.Page {
				continue
			}
			quote := cleanLiteral(evidence.Quote)
			start := strings.Index(literal.Text, quote)
			if start < 0 {
				continue
			}
			needle := values[evidence.Field]
			if evidence.Field == "kind" {
				for _, word := range strings.Fields(quote) {
					word = strings.Trim(word, ".,;:!?()\"«»")
					if normalizedField("kind", word) == m.Kind {
						needle = word
						break
					}
				}
			}
			if needle == "" {
				continue
			}
			offset := strings.Index(quote, needle)
			if offset < 0 {
				continue
			}
			s := literal.Selection
			s.StartByte = start + offset
			s.EndByte = s.StartByte + len(needle)
			if s.EndByte-s.StartByte > 2000 {
				continue
			}
			fields[evidence.Field] = s
		}
	}
	// Only an explicit act-number/date clause establishes edition. Display and
	// publication dates, and validity dates elsewhere in the text, are excluded.
	var identityLiterals []verificationLiteral
	for _, literal := range document.Texts {
		owner, ok := fields["subject"]
		if !ok || owner.ResourceURL != literal.Selection.ResourceURL || owner.Page != literal.Selection.Page || owner.JSONPointer != literal.Selection.JSONPointer {
			continue
		}
		matches := automaticEdition.FindAllStringSubmatchIndex(literal.Text, -1)
		if len(matches) != 1 {
			continue
		}
		identityLiterals = append(identityLiterals, literal)
	}
	if len(identityLiterals) == 1 {
		literal := identityLiterals[0]
		match := automaticEdition.FindStringSubmatchIndex(literal.Text)
		ref := automaticAct.FindStringIndex(literal.Text[match[0]:match[1]])
		if ref == nil {
			return fields
		}
		s := literal.Selection
		s.StartByte = match[0] + ref[0]
		s.EndByte = match[0] + ref[1]
		fields["reference"] = s
		s = literal.Selection
		s.StartByte = match[2]
		s.EndByte = match[3]
		if _, err := time.Parse("02/01/2006", literal.Text[s.StartByte:s.EndByte]); err == nil {
			fields["edition"] = s
		}
	}
	return fields
}
