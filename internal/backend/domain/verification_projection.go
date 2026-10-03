package domain

import (
	"context"
	"encoding/json"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
)

// Projection always uses the primary's supported values and source identity.
// A secondary conflict never changes the originating level or creates a second
// measure. Missing primary facts cannot be supplied from the platform.
func (s *Store) projectVerificationPrimary(ctx context.Context, tx pgx.Tx, kind, municipality string, primary ChannelEvidence, at time.Time) (string, string, error) {
	for _, field := range requiredFacts(kind, primary.Fields) {
		if primary.Fields[field].Value == "" || primary.Fields[field].Value == "unknown" || primary.Fields[field].Value == "not_applicable" {
			return "", "diagnostic", nil
		}
	}
	if kind == "regional_record" && primary.Fields["product"].Value != ProductMonitoring && !slices.Contains([]string{"green", "yellow", "orange", "red"}, primary.Fields["level"].Value) {
		return "", "diagnostic", nil
	}
	fields := []string{"reference", "edition", "kind", "subject", "place", "phase", "validity"}
	if kind == "regional_record" {
		fields = []string{"product", "risk", "zone", "issuance", "validity", "level"}
	}
	identity := map[string]string{}
	for _, name := range fields {
		if f, ok := primary.Fields[name]; ok {
			identity[name] = f.Value
		}
	}
	key := fmt.Sprintf("%s:%s:%d:%s", VerificationLogic, kind, primary.VersionID, verificationHash(identity))
	if _, err := tx.Exec(ctx, "SELECT pg_advisory_xact_lock(hashtextextended($1,730069))", fmt.Sprintf("%s:%d", kind, primary.VersionID)); err != nil {
		return "", "diagnostic", err
	}
	if existing, err := s.matchVerificationProjection(ctx, tx, kind, primary); err != nil {
		return "", "diagnostic", err
	} else if existing != "" {
		if err = s.recordVerificationDetails(ctx, tx, kind, existing, primary, at); err != nil {
			return "", "diagnostic", err
		}
		return existing, "primary_projection", s.recordVerificationProjection(ctx, tx, kind, existing, primary, at)
	}
	var place *string
	if field, ok := primary.Fields["place"]; ok {
		value := cleanLiteral(field.Passage)
		place = &value
	}
	var err error
	switch kind {
	case "local_measure":
		_, err = tx.Exec(ctx, `INSERT INTO domain_local_measures(id,document_version_id,source_id,municipality_istat,issuing_authority_id,publisher_id,platform,kind,subject,place,recorded_at)
 VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11) ON CONFLICT DO NOTHING`, key, primary.VersionID, primary.SourceID, municipality, primary.AuthorityID, primary.PublisherID, primary.Platform, primary.Fields["kind"].Value, cleanLiteral(primary.Fields["subject"].Passage), place, at)
	case "operational_phase":
		_, err = tx.Exec(ctx, `INSERT INTO domain_operational_phases(id,document_version_id,source_id,municipality_istat,authority_id,publisher_id,platform,phase,recorded_at)
 VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9) ON CONFLICT DO NOTHING`, key, primary.VersionID, primary.SourceID, municipality, primary.AuthorityID, primary.PublisherID, primary.Platform, cleanLiteral(primary.Fields["phase"].Passage), at)
	case "regional_record":
		var sourceProduct string
		if err = tx.QueryRow(ctx, "SELECT product_id FROM registry_sources WHERE id=$1", primary.SourceID).Scan(&sourceProduct); err != nil {
			return "", "diagnostic", err
		}
		if sourceProduct != primary.Fields["product"].Value {
			return "", "diagnostic", ErrVerificationEvidence
		}
		_, err = tx.Exec(ctx, `INSERT INTO domain_regional_records(id,document_version_id,source_id,product,originating_authority_id,publisher_id,platform,municipal_republication,recorded_at)
 VALUES($1,$2,$3,$4,$5,$6,$7,false,$8) ON CONFLICT DO NOTHING`, key, primary.VersionID, primary.SourceID, sourceProduct, primary.AuthorityID, primary.PublisherID, primary.Platform, at)
		if err == nil {
			var level *string
			if sourceProduct != ProductMonitoring {
				l := primary.Fields["level"].Value
				level = &l
			}
			_, err = tx.Exec(ctx, `INSERT INTO domain_regional_facts(regional_record_id,product,ordinal,risk,official_risk_label,zone,level)
 VALUES($1,$2,1,$3,$4,$5,$6) ON CONFLICT DO NOTHING`, key, sourceProduct, primary.Fields["risk"].Value, primary.Fields["risk"].Passage, strings.ToUpper(cleanLiteral(primary.Fields["zone"].Passage)), level)
		}
	}
	if err != nil {
		return "", "diagnostic", err
	}
	if err = s.recordVerificationDetails(ctx, tx, kind, key, primary, at); err != nil {
		return "", "diagnostic", err
	}
	return key, "primary_projection", s.recordVerificationProjection(ctx, tx, kind, key, primary, at)
}

func (s *Store) recordVerificationDetails(ctx context.Context, tx pgx.Tx, kind, key string, primary ChannelEvidence, at time.Time) error {
	value := verificationTemporal(kind, key, primary.VersionID, primary.Fields["validity"], at)
	_, err := tx.Exec(ctx, `INSERT INTO domain_temporal_values(entity_kind,entity_id,meaning,original_expression,precision,instant,date_value,end_instant,condition,evidence_document_version_id,created_at)
 SELECT $1,$2,'validity',NULLIF($3,''),$4,$5,$6,$7,$8,$9,$10
 WHERE NOT EXISTS (
 SELECT 1 FROM (SELECT * FROM domain_temporal_values WHERE entity_kind=$1 AND entity_id=$2 AND meaning='validity' AND created_at<=$10 ORDER BY created_at DESC,id DESC LIMIT 1) latest
 WHERE COALESCE(original_expression,'')=$3 AND precision=$4 AND instant IS NOT DISTINCT FROM $5::timestamptz AND date_value IS NOT DISTINCT FROM $6::date
 AND end_instant IS NOT DISTINCT FROM $7::timestamptz AND condition IS NOT DISTINCT FROM $8::text AND timezone IS NULL AND assumption IS NULL)`, kind, key, value.Original, value.Precision, value.Instant, dateValue(value.Date), value.EndInstant, value.Condition, primary.VersionID, at)
	if err != nil {
		return err
	}
	if kind == "local_measure" {
		state := "supported"
		limitations := []string{}
		if value.Precision == "unknown" {
			state = "partial"
			limitations = append(limitations, "validity_not_established")
		}
		body, _ := json.Marshal(limitations)
		_, err = tx.Exec(ctx, `INSERT INTO domain_interpretation_events(local_measure_id,state,evidence_document_version_id,reason,limitations,actor,recorded_at)
 SELECT $1,$2,$3,'multi_source_primary_fields',$4,$5,$6
 WHERE NOT EXISTS (
 SELECT 1 FROM (SELECT * FROM domain_interpretation_events WHERE local_measure_id=$1 AND recorded_at<=$6 ORDER BY recorded_at DESC,id DESC LIMIT 1) latest
 WHERE state=$2 AND evidence_document_version_id=$3 AND actor=$5 AND limitations=$4::jsonb)`, key, state, primary.VersionID, body, VerificationLogic, at)
	}
	return err
}

func verificationTemporal(kind, key string, version int64, field VerifiedField, at time.Time) TemporalValue {
	value := TemporalValue{EntityKind: kind, EntityID: key, Meaning: "validity", Original: field.Passage, Precision: "unknown", EvidenceDocumentVersionID: version, CreatedAt: at}
	if start, end, ok := verificationInterval(field.Passage); ok {
		value.Precision = "interval"
		value.Instant = &start
		value.EndInstant = &end
		return value
	}
	for _, layout := range []string{"2006-01-02", "02/01/2006"} {
		if date, err := time.Parse(layout, cleanLiteral(field.Passage)); err == nil {
			value.Precision = "date"
			value.Date = &date
			return value
		}
	}
	if instant, err := time.Parse(time.RFC3339, cleanLiteral(field.Passage)); err == nil {
		instant = instant.UTC()
		value.Precision = "instant"
		value.Instant = &instant
		return value
	}
	if strings.HasPrefix(strings.ToLower(cleanLiteral(field.Passage)), "fino a ") || strings.HasPrefix(strings.ToLower(cleanLiteral(field.Passage)), "until ") {
		condition := field.Passage
		value.Precision = "conditional"
		value.Condition = &condition
	}
	return value
}

func (s *Store) matchVerificationProjection(ctx context.Context, tx pgx.Tx, kind string, primary ChannelEvidence) (string, error) {
	var rows pgx.Rows
	var err error
	switch kind {
	case "local_measure":
		rows, err = tx.Query(ctx, `SELECT id FROM domain_local_measures WHERE document_version_id=$1 AND source_id=$2 AND kind=$3
 AND lower(regexp_replace(trim(subject),'\s+',' ','g'))=$4 AND lower(regexp_replace(trim(COALESCE(place,'')),'\s+',' ','g'))=$5`, primary.VersionID, primary.SourceID, primary.Fields["kind"].Value, primary.Fields["subject"].Value, primary.Fields["place"].Value)
	case "operational_phase":
		rows, err = tx.Query(ctx, `SELECT id FROM domain_operational_phases WHERE document_version_id=$1 AND source_id=$2 AND lower(regexp_replace(trim(phase),'\s+',' ','g'))=$3`, primary.VersionID, primary.SourceID, primary.Fields["phase"].Value)
	default:
		var level *string
		if primary.Fields["product"].Value != ProductMonitoring {
			l := primary.Fields["level"].Value
			level = &l
		}
		rows, err = tx.Query(ctx, `SELECT r.id FROM domain_regional_records r JOIN domain_regional_facts f ON f.regional_record_id=r.id
 WHERE r.document_version_id=$1 AND r.source_id=$2 AND NOT r.municipal_republication AND r.product=$3 AND f.risk=$4 AND lower(f.zone)=$5 AND f.level IS NOT DISTINCT FROM $6`, primary.VersionID, primary.SourceID, primary.Fields["product"].Value, primary.Fields["risk"].Value, primary.Fields["zone"].Value, level)
	}
	if err != nil {
		return "", err
	}
	ids := []string{}
	for rows.Next() {
		var id string
		if err = rows.Scan(&id); err != nil {
			rows.Close()
			return "", err
		}
		if !slices.Contains(ids, id) {
			ids = append(ids, id)
		}
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return "", err
	}
	matches := []string{}
	for _, id := range ids {
		var original string
		err = tx.QueryRow(ctx, "SELECT COALESCE(original_expression,'') FROM domain_temporal_values WHERE entity_kind=$1 AND entity_id=$2 AND meaning='validity' ORDER BY id DESC LIMIT 1", kind, id).Scan(&original)
		if err != nil && err != pgx.ErrNoRows {
			return "", err
		}
		if normalizedField("validity", original) == primary.Fields["validity"].Value {
			matches = append(matches, id)
		}
	}
	if len(matches) == 1 {
		return matches[0], nil
	}
	if len(matches) > 1 {
		return "", ErrConflict
	}
	return "", nil
}

// Keys describe a primary publication's explicit scope, never a vote between
// secondary channels. Different acts, scopes and regional validity periods stay
// separate; a newer retained version of the same scope preserves its predecessor.
func (s *Store) recordVerificationProjection(ctx context.Context, tx pgx.Tx, kind, id string, primary ChannelEvidence, at time.Time) error {
	fields := []string{"reference", "subject", "place"}
	if kind == "regional_record" {
		fields = []string{"product", "risk", "zone", "issuance", "validity"}
	}
	identity := map[string]string{}
	for _, field := range fields {
		identity[field] = primary.Fields[field].Value
	}
	if kind != "regional_record" && identity["reference"] == "" {
		identity["publication"] = primary.URL
	}
	_, err := tx.Exec(ctx, `INSERT INTO domain_verification_projections(kind,source_id,logical_key,document_version_id,domain_record_id,recorded_at,acquisition_at)
 SELECT $1,$2,$3,$4,$5,$6,COALESCE((SELECT max(acquired_at) FROM retained_acquisitions WHERE version_id=$4 AND acquired_at<=$6),(SELECT first_acquired_at FROM retained_versions WHERE id=$4))
 ON CONFLICT DO NOTHING`, kind, primary.SourceID, verificationHash(identity), primary.VersionID, id, at)
	return err
}
