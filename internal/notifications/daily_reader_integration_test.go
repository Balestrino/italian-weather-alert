//go:build integration

package notifications

import (
	"context"
	"github.com/jackc/pgx/v5"
	"github.com/Balestrino/italian-weather-alert/internal/publicquery"
	"testing"
	"time"
)

func TestReportReaderTerritorialBoundaries(t *testing.T) {
	pool, _ := notificationTestDB(t)
	ctx := context.Background()
	_, err := pool.Exec(ctx, `CREATE TABLE territorial_regions(code text,name text,revision int,enabled bool);
 CREATE TABLE territorial_region_versions(region_code text,revision int,configuration jsonb);
 CREATE TABLE territorial_current_sources(source_id text,region_code text,profile text,municipality_istat text);
 CREATE TABLE registry_sources(id text,product_id text,public_enabled bool,collection_enabled bool,active_revision int,latest_revision int,delay_seconds_override int);
 CREATE TABLE registry_configurations(source_id text,revision int,body jsonb);
 CREATE TABLE acquisition_source_status(source_id text,last_complete_at timestamptz,last_error_code text);
 CREATE TABLE retained_documents(id bigint,source_id text,official_url text);
 CREATE TABLE retained_versions(id bigint,document_id bigint,first_acquired_at timestamptz,metadata jsonb);
 CREATE TABLE processing_runs(id bigint,document_version_id bigint,created_at timestamptz);
 CREATE TABLE classification_results(run_id bigint,status text);
 CREATE TABLE extraction_results(run_id bigint,status text);
 CREATE TABLE geography_municipalities(dataset_id text,istat text,name text);
 CREATE TABLE geography_zone_mappings(dataset_id text,municipality_dataset_id text,municipality_istat text,zone text,territorial_scope text);
 INSERT INTO territorial_regions VALUES('09','Toscana',1,true),('08','Disabled',1,false),('07','Unsupported',0,true);
 INSERT INTO territorial_region_versions VALUES('09',1,'{"municipality_dataset":"selected","zone_dataset":"zones","profiles":["municipal-html"]}');
 INSERT INTO registry_sources VALUES('one','municipal',false,true,1,1,NULL),('two','municipal',false,true,1,1,NULL),('suspended','municipal',false,false,1,1,NULL),('disabled-region','municipal',false,true,1,1,NULL);
 INSERT INTO registry_configurations SELECT id,1,'{}' FROM registry_sources;
 INSERT INTO territorial_current_sources VALUES('one','09','municipal-html','050004'),('two','09','municipal-html','050004'),('suspended','09','municipal-html','050005'),('disabled-region','08','municipal-html','050006');
 INSERT INTO geography_municipalities VALUES('selected','050004','Calcinaia'),('historical','050004','Old name');
 INSERT INTO geography_zone_mappings VALUES('zones','selected','050004','A4','whole_municipality');`)
	if err != nil {
		t.Fatal(err)
	}
	calls := 0
	reader := ReportReader{Pool: pool, facts: func(ctx context.Context, tx pgx.Tx, region string, datasets map[string]string, at time.Time) ([]publicquery.RegionalWarning, []publicquery.Measure, []publicquery.OperationalPhase, error) {
		calls++
		var ro string
		if err := tx.QueryRow(ctx, "SHOW transaction_read_only").Scan(&ro); err != nil || ro != "on" {
			t.Fatal("snapshot not read-only", err)
		}
		if region != "09" || datasets["municipality_registry"] != "selected" {
			t.Fatal("wrong scope")
		}
		return nil, nil, nil, nil
	}}
	s, err := reader.Read(ctx, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	if len(s.Regions) != 2 || calls != 1 {
		t.Fatalf("disabled region or unsupported scope lost: %+v", s)
	}
	var r ReportRegion
	for _, candidate := range s.Regions {
		if candidate.Code == "09" {
			r = candidate
		}
	}
	if len(r.Sources) != 2 || len(r.Municipalities) != 1 || r.Municipalities[0].Name != "Calcinaia" || r.Municipalities[0].Zones[0] != "A4" || r.Sources[0].Public {
		t.Fatalf("collection boundary or selected geography lost: %+v", r)
	}
}
