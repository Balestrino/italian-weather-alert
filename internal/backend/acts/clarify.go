package acts

import (
	"context"
	"encoding/json"
	"errors"
	"slices"
	"time"

	"github.com/jackc/pgx/v5"
)

// ClarificationActor names the logic that appends clarified validity.
const ClarificationActor = "referenced-act-clarification-v1"

// clarifyMeasure applies the act's clause to one notice measure. Only an
// unknown validity is clarified; a determinate one that differs from the act
// records a conflict and stays unchanged. It reports whether it clarified.
func (s *Store) clarifyMeasure(ctx context.Context, p Payload, m NoticeMeasure, actVersion int64, clause Clause, label string, at time.Time) (bool, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return false, err
	}
	defer tx.Rollback(ctx)
	if _, err = tx.Exec(ctx, "SELECT pg_advisory_xact_lock(hashtextextended($1,730101))", m.ID); err != nil {
		return false, err
	}
	var exists bool
	if err = tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM act_clarifications WHERE resolution_id=$1 AND local_measure_id=$2 AND parser_version=$3)`, p.ResolutionID, m.ID, ParserVersion).Scan(&exists); err != nil || exists {
		return false, err
	}
	precision, original := "unknown", ""
	var instant *time.Time
	err = tx.QueryRow(ctx, `SELECT precision,instant,COALESCE(original_expression,'') FROM domain_temporal_values WHERE entity_kind='local_measure' AND entity_id=$1 AND meaning='validity' ORDER BY created_at DESC,id DESC LIMIT 1`, m.ID).Scan(&precision, &instant, &original)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return false, err
	}
	disposition := clause.Match(m.Kind, m.Subject)
	record := func(outcome string) error {
		var ordinal, url, page any
		passage := ""
		if disposition != nil {
			ordinal, url, passage = disposition.Ordinal, clause.ResourceURL, bounded(clause.Passage+" "+disposition.Text, 2000)
			if clause.Page > 0 {
				page = clause.Page
			}
		}
		notice := ""
		if outcome == "conflict" {
			notice = bounded(original, 2000)
		}
		_, err := tx.Exec(ctx, `INSERT INTO act_clarifications(resolution_id,act_document_version_id,local_measure_id,outcome,disposition_ordinal,act_resource_url,act_page_number,act_passage,notice_passage,parser_version,recorded_at)
 VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)`, p.ResolutionID, actVersion, m.ID, outcome, ordinal, url, page, passage, notice, ParserVersion, at.UTC())
		return err
	}
	switch {
	case disposition == nil:
		err = record("not_matched")
	case precision != "unknown":
		if instant != nil && instant.Equal(clause.Start) {
			return false, tx.Commit(ctx)
		}
		err = record("conflict")
	default:
		if err = record("clarified"); err == nil {
			err = appendValidity(ctx, tx, m.ID, actVersion, clause, label, at)
		}
	}
	if err != nil {
		return false, err
	}
	return disposition != nil && precision == "unknown", tx.Commit(ctx)
}

func appendValidity(ctx context.Context, tx pgx.Tx, measure string, actVersion int64, clause Clause, label string, at time.Time) error {
	start := clause.Start.UTC()
	precision, condition := "conditional", &clause.Condition
	var end *time.Time
	if clause.End != nil {
		e := clause.End.UTC()
		precision, condition, end = "interval", nil, &e
	}
	assumption := "Clarified from the dispositive clause of " + label + " cited by the notice; Europe/Rome."
	if _, err := tx.Exec(ctx, `INSERT INTO domain_temporal_values(entity_kind,entity_id,meaning,original_expression,precision,instant,end_instant,timezone,assumption,condition,evidence_document_version_id,created_at)
 VALUES('local_measure',$1,'validity',$2,$3,$4,$5,'Europe/Rome',$6,$7,$8,$9)`, measure, bounded(clause.Passage, 2000), precision, start, end, assumption, condition, actVersion, at.UTC()); err != nil {
		return err
	}
	var raw []byte
	err := tx.QueryRow(ctx, `SELECT limitations FROM domain_interpretation_events WHERE local_measure_id=$1 ORDER BY id DESC LIMIT 1`, measure).Scan(&raw)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return err
	}
	var previous, limitations []string
	_ = json.Unmarshal(raw, &previous)
	resolved := []string{"validity_not_established", "valid_from"}
	if end != nil {
		resolved = append(resolved, "valid_until")
	}
	for _, l := range previous {
		if !slices.Contains(resolved, l) {
			limitations = append(limitations, l)
		}
	}
	state := "supported"
	if len(limitations) > 0 {
		state = "partial"
	} else {
		limitations = []string{}
	}
	encoded, _ := json.Marshal(limitations)
	_, err = tx.Exec(ctx, `INSERT INTO domain_interpretation_events(local_measure_id,state,evidence_document_version_id,reason,limitations,actor,recorded_at)
 VALUES($1,$2,$3,'clarified_by_referenced_act',$4,$5,$6)`, measure, state, actVersion, encoded, ClarificationActor, at.UTC())
	return err
}

func bounded(value string, limit int) string {
	if len(value) <= limit {
		return value
	}
	for limit > 0 && value[limit]&0xC0 == 0x80 {
		limit--
	}
	return value[:limit]
}
