// Package territory provides shared territorial execution boundaries without
// coupling source, job and domain stores to one another.
package territory

import (
	"context"
	"errors"
	"github.com/jackc/pgx/v5"
)

var ErrAssociation = errors.New("source territory is unresolved")
var ErrUnsupported = errors.New("processing profile is unsupported for this territory")
var ErrDisabled = errors.New("region is disabled")
var ErrPublicScope = errors.New("publication outside the supported public territory")

type Query interface {
	QueryRow(context.Context, string, ...any) pgx.Row
}

func Installed(ctx context.Context, q Query) (bool, error) {
	var yes bool
	err := q.QueryRow(ctx, `SELECT to_regclass('territorial_regions') IS NOT NULL`).Scan(&yes)
	return yes, err
}
func Compatible(region, product, profile string) bool {
	return (product == "municipal" && profile == "municipal-html") || (region == "09" && ((profile == "toscana-cfr" && (product == "vigilance" || product == "criticality" || product == "monitoring")) || (profile == "dpc-comparison" && product == "dpc_comparison")))
}
func Associate(ctx context.Context, tx pgx.Tx, source, product, territory, profile, actor string) error {
	installed, err := Installed(ctx, tx)
	if err != nil || !installed {
		return err
	}
	var region string
	var dataset, istat *string
	if product == "municipal" {
		var count int
		err = tx.QueryRow(ctx, `SELECT count(DISTINCT m.region_code),min(m.region_code),min(m.dataset_id) FROM territorial_municipalities m JOIN territorial_regions r ON r.code=m.region_code JOIN territorial_region_versions v ON v.region_code=r.code AND v.revision=r.revision WHERE m.istat=$1 AND m.dataset_id=v.configuration->>'municipality_dataset'`, territory).Scan(&count, &region, &dataset)
		if err != nil || count != 1 {
			return ErrAssociation
		}
		istat = &territory
	} else {
		err = tx.QueryRow(ctx, `SELECT code FROM territorial_regions WHERE code=$1 OR name=$1`, territory).Scan(&region)
		if err != nil {
			return ErrAssociation
		}
	}
	if profile == "" {
		if product == "municipal" {
			profile = "municipal-html"
		} else if region == "09" && product == "dpc_comparison" {
			profile = "dpc-comparison"
		} else if region == "09" {
			profile = "toscana-cfr"
		} else {
			profile = "unsupported"
		}
	}
	_, err = tx.Exec(ctx, `INSERT INTO territorial_source_associations(source_id,region_code,municipality_dataset_id,municipality_istat,profile,actor,recorded_at) VALUES($1,$2,$3,$4,$5,$6,COALESCE((SELECT min(created_at) FROM registry_configurations WHERE source_id=$1),clock_timestamp()))`, source, region, dataset, istat, profile, actor)
	return err
}
func CheckProfile(ctx context.Context, q Query, source, configuredProfile string) error {
	installed, err := Installed(ctx, q)
	if err != nil || !installed {
		return err
	}
	var region, product, profile string
	var listed bool
	err = q.QueryRow(ctx, `SELECT a.region_code,s.product_id,a.profile,COALESCE(v.configuration->'profiles' ? CASE WHEN $2='' THEN a.profile ELSE $2 END,false) FROM territorial_current_sources a JOIN registry_sources s ON s.id=a.source_id JOIN territorial_regions r ON r.code=a.region_code LEFT JOIN territorial_region_versions v ON v.region_code=r.code AND v.revision=r.revision WHERE a.source_id=$1`, source, configuredProfile).Scan(&region, &product, &profile, &listed)
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrAssociation
	}
	if err != nil {
		return err
	}
	if configuredProfile != "" {
		profile = configuredProfile
	}
	if !listed || !Compatible(region, product, profile) {
		return ErrUnsupported
	}
	return nil
}

// CheckExecution is the execution handoff boundary. The SQL function locks the
// region against concurrent disablement; work admitted here may finish.
func CheckExecution(ctx context.Context, q Query, source string) error {
	installed, err := Installed(ctx, q)
	if err != nil || !installed {
		return err
	}
	var allowed bool
	if err = q.QueryRow(ctx, `SELECT territorial_source_allowed($1)`, source).Scan(&allowed); err != nil {
		return err
	}
	if !allowed {
		return ErrDisabled
	}
	return nil
}
func CheckVersion(ctx context.Context, q Query, version int64) error {
	installed, err := Installed(ctx, q)
	if err != nil || !installed {
		return err
	}
	var allowed bool
	if err = q.QueryRow(ctx, `SELECT territorial_job_allowed(jsonb_build_object('document_version_id',$1::bigint))`, version).Scan(&allowed); err != nil {
		return err
	}
	if !allowed {
		return ErrDisabled
	}
	return nil
}

// Predicate is used only with application-owned SQL expressions, never request input.
func Predicate(ctx context.Context, q Query, expression string, job bool) (string, error) {
	installed, err := Installed(ctx, q)
	if err != nil || !installed {
		return "", err
	}
	fn := "territorial_source_allowed"
	if job {
		fn = "territorial_job_allowed"
	}
	return " AND " + fn + "(" + expression + ") ", nil
}

// CheckConfiguration verifies the selected parser as well as the declared profile.
// Drafts remain writable, but a municipal source cannot execute a regional adapter.
func CheckConfiguration(ctx context.Context, q Query, source, profile, regionalKind string) error {
	if err := CheckProfile(ctx, q, source, profile); err != nil {
		return err
	}
	if regionalKind == "" {
		return nil
	}
	installed, err := Installed(ctx, q)
	if err != nil || !installed {
		return err
	}
	var allowed bool
	err = q.QueryRow(ctx, `SELECT a.region_code='09' AND s.product_id=$2 AND COALESCE(NULLIF($3,''),a.profile)='toscana-cfr' FROM territorial_current_sources a JOIN registry_sources s ON s.id=a.source_id WHERE a.source_id=$1`, source, regionalKind, profile).Scan(&allowed)
	if err != nil {
		return err
	}
	if !allowed {
		return ErrUnsupported
	}
	return nil
}
