package diagnostics

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Store struct{ pool *pgxpool.Pool }

func New(pool *pgxpool.Pool) *Store { return &Store{pool: pool} }

func (s *Store) Snapshot(ctx context.Context, sourceID string, at time.Time) (Snapshot, error) {
	if s == nil || s.pool == nil || strings.TrimSpace(sourceID) == "" || at.IsZero() {
		return Snapshot{}, ErrInvalid
	}
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.RepeatableRead, AccessMode: pgx.ReadOnly})
	if err != nil {
		return Snapshot{}, err
	}
	defer tx.Rollback(ctx)
	var collection bool
	var delay int
	if err = tx.QueryRow(ctx, `SELECT s.collection_enabled,COALESCE(s.delay_seconds_override,(c.body->>'delay_seconds')::integer,1800) FROM registry_sources s JOIN registry_configurations c ON c.source_id=s.id AND c.revision=COALESCE(s.active_revision,s.latest_revision) WHERE s.id=$1`, sourceID).Scan(&collection, &delay); errors.Is(err, pgx.ErrNoRows) {
		return Snapshot{}, ErrNotFound
	} else if err != nil {
		return Snapshot{}, err
	}
	result := Snapshot{SourceID: sourceID, ObservedAt: at.UTC()}
	var checkID *int64
	var started, finished, lastComplete *time.Time
	var reachable, recognized, complete *bool
	var errorCode *string
	err = tx.QueryRow(ctx, `SELECT l.id,l.started_at,l.finished_at,l.reachable,l.content_recognized,l.complete,l.error_code,
 (SELECT max(finished_at) FROM acquisition_checks WHERE source_id=$1 AND complete AND finished_at<=$2)
 FROM acquisition_checks l WHERE l.source_id=$1 AND l.finished_at<=$2 ORDER BY l.finished_at DESC,l.id DESC LIMIT 1`, sourceID, at.UTC()).Scan(&checkID, &started, &finished, &reachable, &recognized, &complete, &errorCode, &lastComplete)
	availabilityEvidence, _ := json.Marshal(map[string]any{"check_id": checkID, "started_at": started, "finished_at": finished, "error_code": errorCode})
	result.Availability = Signal{Evidence: availabilityEvidence, Limitations: []string{}}
	switch {
	case !collection:
		result.Availability.State = "suspended"
		result.Availability.Limitations = []string{"collection is disabled; availability is not inferred from retained facts"}
	case errors.Is(err, pgx.ErrNoRows):
		result.Availability.State = "unverified"
		result.Availability.Limitations = []string{"no completed source check is available"}
	case err != nil:
		return Snapshot{}, err
	case reachable != nil && recognized != nil && complete != nil && *reachable && *recognized && *complete && errorCode == nil:
		result.Availability.State = "available"
	default:
		result.Availability.State = "unavailable"
		result.Availability.Limitations = []string{"latest check is not complete and recognizable; retained fact validity is separate"}
	}
	timelinessEvidence, _ := json.Marshal(map[string]any{"last_complete_at": lastComplete, "delay_threshold_seconds": delay})
	result.Timeliness = Signal{Evidence: timelinessEvidence, Limitations: []string{}}
	switch {
	case !collection:
		result.Timeliness.State = "suspended"
		result.Timeliness.Limitations = []string{"collection is disabled"}
	case lastComplete == nil:
		result.Timeliness.State = "unverified"
		result.Timeliness.Limitations = []string{"no complete-check baseline is available"}
	case !at.UTC().Before(lastComplete.Add(time.Duration(delay) * time.Second)):
		result.Timeliness.State = "delayed"
		result.Timeliness.Limitations = []string{"complete-check delay threshold reached; this does not imply warning expiry"}
	default:
		result.Timeliness.State = "current"
	}
	var versions, unresolved int64
	if err = tx.QueryRow(ctx, `SELECT count(*),count(*) FILTER(WHERE classification_status<>'classified' OR (relevant AND extraction_status<>'extracted')) FROM (
 SELECT v.id,COALESCE(c.status,'not_processed') classification_status,COALESCE(c.relevant,false) relevant,COALESCE(e.status,'not_processed') extraction_status
 FROM retained_versions v JOIN retained_documents d ON d.id=v.document_id
 LEFT JOIN LATERAL (SELECT run_id,status,relevant FROM classification_results WHERE document_version_id=v.id ORDER BY created_at DESC,run_id DESC LIMIT 1) c ON true
 LEFT JOIN LATERAL (SELECT status FROM extraction_results WHERE document_version_id=v.id AND classification_run_id=c.run_id ORDER BY created_at DESC,run_id DESC LIMIT 1) e ON true
 WHERE d.source_id=$1 AND v.first_acquired_at<=$2) q`, sourceID, at.UTC()).Scan(&versions, &unresolved); err != nil {
		return Snapshot{}, err
	}
	var reviewedSuites int64
	var blocked bool
	if err = tx.QueryRow(ctx, `SELECT count(*),COALESCE(bool_or(NOT passed),false) FROM (
 SELECT DISTINCT ON (suite) suite,passed FROM registry_regressions
 WHERE source_id=$1 AND recorded_at<=$2 ORDER BY suite,recorded_at DESC,id DESC
) latest`, sourceID, at.UTC()).Scan(&reviewedSuites, &blocked); err != nil {
		return Snapshot{}, err
	}
	interpretationEvidence, _ := json.Marshal(map[string]any{"retained_versions": versions, "unresolved_versions": unresolved, "reviewed_regression_suites": reviewedSuites, "blocking_regression": blocked})
	result.Interpretation = Signal{Evidence: interpretationEvidence, Limitations: []string{"terminal model output alone is not correctness; reviewed regressions are required and remain bounded evidence"}}
	switch {
	case versions == 0:
		result.Interpretation.State = "unverified"
	case blocked || unresolved > 0:
		result.Interpretation.State = "degraded"
	case reviewedSuites == 0:
		result.Interpretation.State = "unverified"
	default:
		result.Interpretation.State = "supported"
	}
	if err = tx.Commit(ctx); err != nil {
		return Snapshot{}, err
	}
	return result, nil
}

func (s *Store) Compare(ctx context.Context, value Comparison) (Comparison, error) {
	if s == nil || s.pool == nil || value.RegionalVersionID < 1 || value.DPCVersionID < 1 || strings.TrimSpace(value.Scope) == "" || strings.TrimSpace(value.Actor) == "" || value.ComparedAt.IsZero() || len(value.Dimensions) != 4 {
		return Comparison{}, ErrInvalid
	}
	if err := validateDimensions(value.Dimensions); err != nil {
		return Comparison{}, err
	}
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return Comparison{}, err
	}
	defer tx.Rollback(ctx)
	regionalSource, regionalProduct, err := versionSource(ctx, tx, value.RegionalVersionID)
	if err != nil {
		return Comparison{}, err
	}
	dpcSource, dpcProduct, err := versionSource(ctx, tx, value.DPCVersionID)
	if err != nil {
		return Comparison{}, err
	}
	if (regionalProduct != "vigilance" && regionalProduct != "criticality" && regionalProduct != "monitoring") || dpcProduct != "dpc_comparison" || value.RegionalVersionID == value.DPCVersionID {
		return Comparison{}, ErrInvalid
	}
	for _, dimension := range value.Dimensions {
		if err = evidenceBelongs(ctx, tx, value.RegionalVersionID, dimension.Regional); err != nil {
			return Comparison{}, err
		}
		if err = evidenceBelongs(ctx, tx, value.DPCVersionID, dimension.DPC); err != nil {
			return Comparison{}, err
		}
	}
	value.RegionalSourceID, value.DPCSourceID = regionalSource, dpcSource
	wire, _ := json.Marshal(value.Dimensions)
	value.ComparedAt = value.ComparedAt.UTC().Truncate(time.Microsecond)
	if value.ID == "" {
		hash := sha256.Sum256([]byte(fmt.Sprintf("%d\x00%d\x00%s\x00%s\x00%s\x00%s", value.RegionalVersionID, value.DPCVersionID, value.Scope, value.Actor, value.ComparedAt.UTC().Format(time.RFC3339Nano), wire)))
		value.ID = "dpc-" + hex.EncodeToString(hash[:])[:24]
	}
	tag, err := tx.Exec(ctx, `INSERT INTO diagnostic_dpc_comparisons(id,regional_source_id,regional_version_id,dpc_source_id,dpc_version_id,scope,dimensions,actor,compared_at) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9) ON CONFLICT DO NOTHING`, value.ID, value.RegionalSourceID, value.RegionalVersionID, value.DPCSourceID, value.DPCVersionID, value.Scope, wire, value.Actor, value.ComparedAt)
	if err != nil {
		return Comparison{}, err
	}
	if tag.RowsAffected() == 0 {
		var stored Comparison
		var sameDimensions bool
		if err = tx.QueryRow(ctx, `SELECT regional_source_id,regional_version_id,dpc_source_id,dpc_version_id,scope,dimensions=$2::jsonb,actor,compared_at FROM diagnostic_dpc_comparisons WHERE id=$1`, value.ID, wire).Scan(&stored.RegionalSourceID, &stored.RegionalVersionID, &stored.DPCSourceID, &stored.DPCVersionID, &stored.Scope, &sameDimensions, &stored.Actor, &stored.ComparedAt); err != nil {
			return Comparison{}, err
		}
		if !sameDimensions || stored.RegionalSourceID != value.RegionalSourceID || stored.RegionalVersionID != value.RegionalVersionID || stored.DPCSourceID != value.DPCSourceID || stored.DPCVersionID != value.DPCVersionID || stored.Scope != value.Scope || stored.Actor != value.Actor || !stored.ComparedAt.Equal(value.ComparedAt) {
			return Comparison{}, ErrConflict
		}
	}
	if err = tx.Commit(ctx); err != nil {
		return Comparison{}, err
	}
	return value, nil
}

func versionSource(ctx context.Context, tx pgx.Tx, versionID int64) (string, string, error) {
	var source, product string
	err := tx.QueryRow(ctx, `SELECT d.source_id,s.product_id FROM retained_versions v JOIN retained_documents d ON d.id=v.document_id JOIN registry_sources s ON s.id=d.source_id WHERE v.id=$1`, versionID).Scan(&source, &product)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", "", ErrNotFound
	}
	return source, product, err
}

func evidenceBelongs(ctx context.Context, tx pgx.Tx, versionID int64, evidence Evidence) error {
	if evidence.ResourceURL == "" || evidence.Locator == "" || evidence.Page != nil && *evidence.Page < 1 {
		return ErrInvalid
	}
	var found bool
	if err := tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM retained_resources WHERE version_id=$1 AND url=$2 AND missing='')`, versionID, evidence.ResourceURL).Scan(&found); err != nil {
		return err
	}
	if !found {
		return ErrInvalid
	}
	return nil
}

func validateDimensions(values []Dimension) error {
	want := map[string]bool{"risk": true, "territory": true, "issuance": true, "validity": true}
	seen := map[string]bool{}
	for _, value := range values {
		if !want[value.Name] || seen[value.Name] || value.State != "same" && value.State != "different" && value.State != "not_comparable" {
			return ErrInvalid
		}
		seen[value.Name] = true
		if value.State == "not_comparable" {
			if strings.TrimSpace(value.Reason) == "" || value.RegionalValue != nil || value.DPCValue != nil {
				return ErrInvalid
			}
		} else {
			if value.RegionalValue == nil || value.DPCValue == nil || strings.TrimSpace(*value.RegionalValue) == "" || strings.TrimSpace(*value.DPCValue) == "" || strings.TrimSpace(value.Reason) != "" {
				return ErrInvalid
			}
		}
	}
	return nil
}
