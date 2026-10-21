package acquisition

import (
	"context"
	"strings"
	"time"

	"github.com/Balestrino/italian-weather-alert/internal/backend/registry"
)

// FollowedTarget is a tracked target whose maintained measure state or
// explicit reference can keep it in revision checks, with the reason it is or
// is no longer selected by the current plan.
type FollowedTarget struct {
	URL                   string     `json:"url"`
	MeasureState          string     `json:"measure_state"`
	RelevanceState        string     `json:"relevance_state"`
	ExplicitReference     bool       `json:"explicit_reference"`
	PublicationDate       *time.Time `json:"publication_date"`
	AgeDays               int        `json:"age_days"`
	LimitDays             int        `json:"unresolved_limit_days"`
	InterpretationStateAt *time.Time `json:"interpretation_state_at"`
	Recheck               string     `json:"recheck"`
}

// FollowedTargets lists targets of the source's active configuration with a
// measure state other than none or an explicit reference. It reads the state
// derived by the last check and never refreshes or changes it.
func (s *TrackingStore) FollowedTargets(ctx context.Context, sourceID string, now time.Time) ([]FollowedTarget, error) {
	if strings.TrimSpace(sourceID) == "" || now.IsZero() {
		return nil, ErrInvalidConfiguration
	}
	var windowDays, limitDays int
	var bootstrap bool
	err := s.pool.QueryRow(ctx, `SELECT COALESCE(NULLIF(c.body->'discovery'->>'bootstrap_days','')::int,30),
 COALESCE(NULLIF(c.body->'discovery'->>'unresolved_measure_days','')::int,0),
 COALESCE((SELECT completed_at IS NULL FROM acquisition_bootstrap b WHERE b.source_id=r.id AND b.configuration=r.active_revision),false)
 FROM registry_sources r JOIN registry_configurations c ON c.source_id=r.id AND c.revision=r.active_revision WHERE r.id=$1`, sourceID).Scan(&windowDays, &limitDays, &bootstrap)
	if err != nil {
		return nil, err
	}
	limitDays = registry.Discovery{UnresolvedMeasureDays: limitDays}.EffectiveUnresolvedMeasureDays()
	cutoff := now.UTC().AddDate(0, 0, -windowDays)
	rows, err := s.pool.Query(ctx, `SELECT t.url,t.measure_state,t.relevance_state,t.explicit_reference,t.source_publication_date,
 GREATEST(0,EXTRACT(DAY FROM $2::timestamptz-COALESCE(t.source_publication_date::timestamptz,t.first_discovered_at)))::int,
 t.interpretation_state_at,
 CASE
   WHEN disposition.decision='exclude_stale_unavailable' AND t.last_seen_at<=disposition.observed_last_seen_at THEN 'not_rechecked'
   WHEN t.source_publication_date IS NULL OR t.source_publication_date>=$3::date THEN 'window'
   WHEN t.measure_state IN ('ongoing','unresolved') THEN t.measure_state
   WHEN $4 AND t.explicit_reference THEN 'bootstrap_reference'
   ELSE 'not_rechecked' END
 FROM acquisition_targets t JOIN registry_sources r ON r.id=t.source_id AND r.active_revision=t.configuration
 LEFT JOIN LATERAL (SELECT decision,observed_last_seen_at FROM acquisition_target_dispositions d
   WHERE d.source_id=t.source_id AND d.configuration=t.configuration AND d.url=t.url ORDER BY d.id DESC LIMIT 1) disposition ON true
 WHERE t.source_id=$1 AND (t.measure_state<>'none' OR t.explicit_reference)
 ORDER BY t.source_publication_date DESC NULLS LAST,t.url`, sourceID, now.UTC(), cutoff, bootstrap)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := []FollowedTarget{}
	for rows.Next() {
		item := FollowedTarget{LimitDays: limitDays}
		if err = rows.Scan(&item.URL, &item.MeasureState, &item.RelevanceState, &item.ExplicitReference, &item.PublicationDate, &item.AgeDays, &item.InterpretationStateAt, &item.Recheck); err != nil {
			return nil, err
		}
		result = append(result, item)
	}
	return result, rows.Err()
}
