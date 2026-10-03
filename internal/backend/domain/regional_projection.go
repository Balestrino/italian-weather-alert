package domain

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/Balestrino/italian-weather-alert/internal/backend/acquisition"
	"github.com/Balestrino/italian-weather-alert/internal/backend/documents"
	"github.com/jackc/pgx/v5"
	"time"
)

const RegionalProjectionLogic = "cfr-vector-v3"

type RegionalDocuments interface {
	Version(context.Context, int64) (documents.Version, error)
	Read(context.Context, int64, string) ([]byte, error)
}

type ProjectionReport struct {
	VersionID   int64    `json:"version_id"`
	Product     string   `json:"product"`
	Status      string   `json:"status"`
	Statement   string   `json:"statement"`
	Facts       int      `json:"facts"`
	Limitations []string `json:"limitations"`
}

// ProjectCFR runs only against retained evidence. Atomic publication prevents
// a reader from observing records before their validity and limitations exist.
// Source suspension is locked with the write, and every result identifies the
// exact version and parser. It never grants source acceptance/public access.
func (s *Store) ProjectCFR(ctx context.Context, docs RegionalDocuments, id int64) (ProjectionReport, error) {
	return s.ProjectCFRWithReader(ctx, docs, id, acquisition.PopplerVectorReader{})
}

func (s *Store) ProjectCFRWithReader(ctx context.Context, docs RegionalDocuments, id int64, reader acquisition.VectorPDFReader) (ProjectionReport, error) {
	out := ProjectionReport{VersionID: id, Status: "not_regional", Limitations: []string{}}
	var source, product, authority, publisher, platform string
	err := s.pool.QueryRow(ctx, `SELECT s.id,s.product_id,s.authority_id,c.publisher_id,c.platform FROM retained_versions v JOIN retained_documents d ON d.id=v.document_id JOIN registry_sources s ON s.id=d.source_id JOIN registry_channels c ON c.id=s.channel_id WHERE v.id=$1`, id).Scan(&source, &product, &authority, &publisher, &platform)
	if err != nil {
		return out, err
	}
	out.Product = product
	if product != "criticality" && product != "vigilance" && product != "monitoring" {
		return out, nil
	}
	var old ProjectionReport
	old.VersionID = id
	old.Product = product
	err = s.pool.QueryRow(ctx, `SELECT status,statement,limitations,(SELECT count(*) FROM domain_regional_records WHERE document_version_id=$1 AND projection_logic=$2) FROM domain_regional_projections WHERE document_version_id=$1 AND logic_version=$2`, id, RegionalProjectionLogic).Scan(&old.Status, &old.Statement, &old.Limitations, &old.Facts)
	if err == nil {
		return old, nil
	}
	if err != pgx.ErrNoRows {
		return out, err
	}
	version, err := docs.Version(ctx, id)
	if err != nil {
		return out, err
	}
	var original string
	for _, r := range version.Resources {
		if r.Role == "original" && r.SourceID == source {
			original = r.URL
		}
	}
	if original == "" {
		return out, ErrInvalid
	}
	body, err := docs.Read(ctx, id, original)
	if err != nil {
		return out, err
	}
	projection, parseErr := acquisition.ProjectRegionalHTML(product, body)
	if product == "criticality" && version.Complete {
		for _, resource := range version.Resources {
			if resource.SourceID == source && resource.Required && resource.Missing == "" && resource.MediaType == "application/pdf" {
				pdf, e := docs.Read(ctx, id, resource.URL)
				if e == nil {
					var evidence acquisition.VectorEvidence
					evidence, e = reader.Read(ctx, pdf)
					if e == nil {
						var maps acquisition.RegionalProjection
						maps, e = acquisition.ProjectCriticalityVector(body, resource.URL, evidence)
						if e == nil && parseErr == nil {
							maps, e = acquisition.MergeCriticalityMaps(maps, projection)
						}
						if e == nil {
							projection = maps
							parseErr = nil
						}
					}
				}
				if e != nil {
					projection.Limitations = append(projection.Limitations, "Retained PDF graphical interpretation failed; preceding explicit HTML evidence only.")
				}
				break
			}
		}
	}
	out.Status = "partial"
	out.Statement = projection.Statement
	out.Limitations = projection.Limitations
	if !version.Complete {
		parseErr = fmt.Errorf("required evidence incomplete")
	}
	if parseErr != nil {
		out.Status = "unsupported"
		out.Statement = ""
		out.Limitations = []string{"Retained regional product could not be fully parsed or required evidence is incomplete; consult the original. No alert level or all-clear is inferred."}
		projection.Facts = nil
	}
	if product == "monitoring" && out.Statement == "NESSUN AVVISO IN CORSO DI VALIDITÀ O EVENTO IN CORSO" && parseErr == nil {
		out.Status = "no_event"
	}
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return out, err
	}
	defer tx.Rollback(ctx)
	var suspended *time.Time
	var allowed bool
	if err = tx.QueryRow(ctx, `SELECT interpretation_suspended_at,territorial_source_allowed(id) FROM registry_sources WHERE id=$1 FOR SHARE`, source).Scan(&suspended, &allowed); err != nil {
		return out, err
	}
	if suspended != nil || !allowed {
		out.Status = "suspended"
		return out, nil
	}
	if _, err = tx.Exec(ctx, `SELECT pg_advisory_xact_lock(hashtextextended($1,730060))`, fmt.Sprintf("%s:%d", RegionalProjectionLogic, id)); err != nil {
		return out, err
	}
	limitations, _ := json.Marshal(out.Limitations)
	tag, err := tx.Exec(ctx, `INSERT INTO domain_regional_projections(document_version_id,logic_version,source_id,product,status,statement,limitations) VALUES($1,$2,$3,$4,$5,$6,$7) ON CONFLICT DO NOTHING`, id, RegionalProjectionLogic, source, product, out.Status, out.Statement, limitations)
	if err != nil {
		return out, err
	}
	if tag.RowsAffected() == 0 {
		return out, tx.Commit(ctx)
	}
	at := time.Now().UTC()
	for n, f := range projection.Facts {
		if f.EvidenceURL != "" {
			raw, _ := json.Marshal(map[string]any{"source_url": f.EvidenceURL, "page": f.Page, "locator": f.Locator})
			f.Locator = string(raw)
		}
		key := fmt.Sprintf("%s:%d:%d", RegionalProjectionLogic, id, n+1)
		_, err = tx.Exec(ctx, `INSERT INTO domain_regional_records(id,document_version_id,source_id,product,originating_authority_id,publisher_id,platform,municipal_republication,projection_logic,evidence_locator) VALUES($1,$2,$3,$4,$5,$6,$7,false,$8,$9)`, key, id, source, product, authority, publisher, platform, RegionalProjectionLogic, f.Locator)
		if err != nil {
			return out, err
		}
		_, err = tx.Exec(ctx, `INSERT INTO domain_regional_facts(regional_record_id,product,ordinal,risk,official_risk_label,zone,level) VALUES($1,$2,1,$3,$4,$5,$6)`, key, product, f.Risk, f.Label, f.Zone, f.Level)
		if err != nil {
			return out, err
		}
		var zone, assumption *string
		if f.Precision == "interval" {
			z, a := "Europe/Rome", "CFR official local bulletin times interpreted in Europe/Rome; original minute precision retained."
			zone, assumption = &z, &a
		}
		_, err = tx.Exec(ctx, `INSERT INTO domain_temporal_values(entity_kind,entity_id,meaning,original_expression,precision,instant,date_value,end_instant,timezone,assumption,evidence_document_version_id,created_at) VALUES('regional_record',$1,'validity',$2,$3,$4,$5,$6,$7,$8,$9,$10)`, key, f.Original, f.Precision, f.Start, dateValue(f.Date), f.End, zone, assumption, id, at)
		if err != nil {
			return out, err
		}
	}
	out.Facts = len(projection.Facts)
	return out, tx.Commit(ctx)
}
