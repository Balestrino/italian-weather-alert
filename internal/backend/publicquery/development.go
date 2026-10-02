package publicquery

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	"strings"
)

const developmentPublicationWarning = "Development publication: manual municipality override; production readiness is not certified."

// Environment is supplied by private runtime configuration, never public input.
func NewForEnvironment(pool *pgxpool.Pool, environment string) *Store {
	s := New(pool)
	s.development = environment == "development"
	return s
}
func (s *Store) forMunicipality(istat string) *Store {
	copy := *s
	if istat != "" {
		copy.municipalityScope = istat
	}
	return &copy
}
func sqlLiteral(value string) string { return "'" + strings.ReplaceAll(value, "'", "''") + "'" }
func (s *Store) developmentSourceSQL() string {
	if !s.development || s.region != "" {
		return "false"
	}
	scope := "true"
	if s.municipalityScope != "" {
		scope = "dm.istat=" + sqlLiteral(s.municipalityScope)
	}
	return `EXISTS(SELECT 1 FROM territorial_current_sources a
 JOIN registry_configurations dc ON dc.source_id=s.id AND dc.revision=s.active_revision
 JOIN development_publication_current_municipalities dm ON dm.region_code=a.region_code
 WHERE a.source_id=s.id AND a.region_code='09' AND ` + scope + `
 AND COALESCE((dc.body#>>'{policy,publication_permitted}')::boolean,false)
 AND COALESCE(NULLIF(dc.body->>'processing_profile',''),a.profile)=CASE WHEN s.product_id='municipal' THEN 'municipal-html' ELSE 'toscana-cfr' END
 AND (dc.body#>>'{regional_product,kind}' IS NULL OR dc.body#>>'{regional_product,kind}'=s.product_id)
 AND dc.body->'provenance' IS NOT NULL AND dc.body->'provenance'<>'null'::jsonb
 AND COALESCE(dc.body->'unresolved','[]'::jsonb) IN ('[]'::jsonb,'null'::jsonb)
 AND ((s.product_id='municipal' AND a.municipality_istat=dm.istat AND s.territory=dm.istat)
 OR (a.municipality_istat IS NULL AND s.product_id IN ('criticality','vigilance','monitoring'))))`
}
func (s *Store) developmentVisibilitySQL() string {
	return `(` + s.developmentSourceSQL() + `) AND EXISTS(SELECT 1 FROM retained_acquisitions da WHERE da.source_id=s.id AND da.configuration=s.active_revision AND da.version_id=v.id AND da.document_id=v.document_id)`
}

// Shared regional bulletins may be read, but their facts remain scoped to selected municipalities.
func (s *Store) regionalDevelopmentScopeSQL(mapping string) string {
	if !s.development || s.region != "" {
		return "true"
	}
	return `(` + strictVisibilitySQL + `) OR EXISTS(SELECT 1 FROM development_publication_current_municipalities dm JOIN geography_zone_mappings dz ON dz.municipality_istat=dm.istat AND dz.dataset_id=` + sqlLiteral(mapping) + ` WHERE dm.region_code='09' AND dz.zone=f.zone)`
}
func (s *Store) developmentSource(ctx context.Context, sourceID string) (bool, error) {
	if !s.development {
		return false, nil
	}
	var enabled bool
	err := s.pool.QueryRow(ctx, `SELECT NOT s.public_enabled AND (`+s.developmentSourceSQL()+`) FROM `+s.sourcesSQL()+` s WHERE s.id=$1`, sourceID).Scan(&enabled)
	return enabled, err
}

// Bind saved API/MCP views to environment and publication revision. Revocation also expires old views.
func (s *Store) PublicationScope(ctx context.Context) (string, error) {
	if !s.development {
		return "strict", nil
	}
	var scope string
	err := s.pool.QueryRow(ctx, `SELECT 'development:'||md5(concat_ws('|',
 (SELECT string_agg(region_code||':'||istat||':'||revision::text||':'||enabled::text,',' ORDER BY region_code,istat) FROM development_publication_municipalities),
 (SELECT string_agg(code||':'||revision::text,',' ORDER BY code) FROM territorial_regions),
 (SELECT string_agg(id||':'||COALESCE(active_revision,0)::text||':'||public_enabled::text,',' ORDER BY id) FROM registry_sources),
 (SELECT max(id)::text FROM territorial_source_associations)))`).Scan(&scope)
	return scope, err
}
