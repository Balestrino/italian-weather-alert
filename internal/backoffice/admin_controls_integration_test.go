//go:build integration

package backoffice

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/Balestrino/italian-weather-alert/internal/backend/domain"
	"github.com/Balestrino/italian-weather-alert/internal/backend/jobs"
	"github.com/Balestrino/italian-weather-alert/internal/backend/operations"
	"github.com/Balestrino/italian-weather-alert/internal/backend/registry"
)

// Exercise the shipped executable against the same isolated registry as the admin.
// Neither interface constructs a crawler/provider or runs background processes.
func TestCLIAdminActivationParity(t *testing.T) {
	ctx := context.Background()
	p := territorialAdminDB(t)
	dataset := adoptTestMunicipalities(t, p, "09", 2)
	geo, reg := domain.New(p), registry.New(p)
	if _, err := geo.ConfigureRegion(ctx, "09", 1, domain.RegionConfiguration{MunicipalityDataset: dataset, Profiles: []string{"municipal-html"}}, "fixture"); err != nil {
		t.Fatal(err)
	}
	h := HandlerWithAdministration(nil, AdminRuntime{Territories: operations.New(p), TerritoryConfig: geo, Registry: reg, Embedding: jobs.New(p)})
	post := func(path string, fields url.Values, status int, back string) {
		t.Helper()
		r := httptest.NewRequest("POST", "http://127.0.0.1"+path, strings.NewReader(fields.Encode()))
		r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		w := httptest.NewRecorder()
		h.ServeHTTP(w, r)
		if w.Code != status || (back != "" && w.Header().Get("Location") != back) {
			t.Fatalf("POST %s: %d %s %s", path, w.Code, w.Header().Get("Location"), w.Body)
		}
	}
	tmp := t.TempDir()
	binary := filepath.Join(tmp, "iwa")
	if out, err := exec.Command("go", "build", "-o", binary, "../../cmd/iwa").CombinedOutput(); err != nil {
		t.Fatalf("build: %v %s", err, out)
	}
	password := filepath.Join(tmp, "postgres-password")
	token := filepath.Join(tmp, "crawl-token")
	if err := os.WriteFile(password, []byte(p.Config().ConnConfig.Password), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(token, []byte("synthetic-unused-token"), 0600); err != nil {
		t.Fatal(err)
	}
	env := []string{}
	for _, v := range os.Environ() {
		if !strings.HasPrefix(v, "IWA_") && !strings.HasPrefix(v, "COMPOSE_") {
			env = append(env, v)
		}
	}
	env = append(env, "IWA_ROLE=admin", "IWA_POSTGRES_HOST="+fmt.Sprintf("%s:%d", p.Config().ConnConfig.Host, p.Config().ConnConfig.Port), "IWA_POSTGRES_PASSWORD_FILE="+password, "IWA_CRAWL_TOKEN_FILE="+token)
	cli := func(failure string, args ...string) map[string]any {
		t.Helper()
		cmd := exec.Command(binary, args...)
		cmd.Env = env
		out, err := cmd.CombinedOutput()
		if failure != "" {
			if err == nil || !strings.Contains(string(out), "code="+failure) {
				t.Fatalf("CLI %v: expected %s; %v %s", args, failure, err, out)
			}
			return nil
		}
		if err != nil {
			t.Fatalf("CLI %v: %v %s", args, err, out)
		}
		var state map[string]any
		if err = json.Unmarshal(out, &state); err != nil {
			t.Fatalf("CLI JSON: %v %s", err, out)
		}
		return state
	}
	cli("complete_municipality_register_required", "region-enable", "03", "0", "cli")
	cli("", "region-enable", "09", "2", "cli-region")
	post("/admin/regions/09/enabled", url.Values{"actor": {"admin-region"}, "expected_revision": {"3"}, "enabled": {"false"}, "return_to": {"/admin/regions?state=all"}}, 303, "/admin/regions?state=all")
	if cli("", "region-status", "09")["enabled"] != false {
		t.Fatal("admin region change not visible to CLI")
	}
	cli("revision_conflict", "region-enable", "09", "3", "stale")
	if cli("", "municipality-enable", "09", "909000", "0", "cli-municipality")["blocked_by"] != "region_disabled" {
		t.Fatal("parent gate not reported")
	}
	back := "/admin/regions/09?tab=municipalities&search=Comune&province=PR&coverage=none"
	post("/admin/regions/09/municipalities/909000/enabled", url.Values{"actor": {"admin-municipality"}, "expected_revision": {"1"}, "enabled": {"false"}, "return_to": {back}}, 303, back)
	if cli("", "municipality-status", "09", "909000")["enabled"] != false {
		t.Fatal("admin municipality change not visible to CLI")
	}
	post("/admin/regions/09/municipalities/909000/enabled", url.Values{"actor": {"stale"}, "expected_revision": {"1"}, "enabled": {"true"}}, 409, "")
	post("/admin/regions/09/municipalities/909000/enabled", url.Values{"actor": {""}, "expected_revision": {"2"}, "enabled": {"true"}}, 400, "")
	post("/admin/regions/09/municipalities/909000/enabled", url.Values{"actor": {"bad-return"}, "expected_revision": {"2"}, "enabled": {"true"}, "return_to": {"https://example.test/"}}, 400, "")
	cli("not_found", "municipality-enable", "03", "909000", "0", "foreign")
	cli("", "region-enable", "09", "4", "cli-region")
	cli("", "municipality-enable", "09", "909000", "2", "cli-municipality")
	if err := reg.CreateAuthority(ctx, registry.Authority{ID: "a", Name: "Synthetic", OfficialURL: "https://example.test"}); err != nil {
		t.Fatal(err)
	}
	if err := reg.CreateChannel(ctx, registry.Channel{ID: "c", PublisherID: "a", Platform: "web", URL: "https://example.test"}); err != nil {
		t.Fatal(err)
	}
	cfg := registry.Configuration{URL: "https://example.test", Sections: []string{"https://example.test"}, AccessMethod: "html", Attribution: "Synthetic fixture", Policy: registry.Policy{CollectionPermitted: true, RetentionPermitted: true, Evidence: &registry.Evidence{URL: "https://example.test", Locator: "synthetic", ObservedAt: time.Now()}}}
	if err := reg.CreateSource(ctx, registry.Source{ID: "source", AuthorityID: "a", ChannelID: "c", ProductID: "municipal", Territory: "909000"}, cfg, "fixture"); err != nil {
		t.Fatal(err)
	}
	if cli("", "embedding-status")["enabled"] != false || cli("", "source-embedding-status", "source")["enabled"] != false {
		t.Fatal("embedding controls did not default to disabled")
	}
	cli("", "embedding-enable", "0", "cli-global")
	post("/admin/sources/source/embedding", url.Values{"actor": {"admin-source"}, "expected_revision": {"0"}, "enabled": {"true"}}, 303, "/admin/embedding")
	if cli("", "source-embedding-status", "source")["enabled"] != true {
		t.Fatal("source embedding admin change not visible to CLI")
	}
	post("/admin/embedding", url.Values{"actor": {"admin-global"}, "expected_revision": {"1"}, "enabled": {"false"}}, 303, "/admin/embedding")
	if cli("", "embedding-status")["enabled"] != false || cli("", "source-embedding-status", "source")["enabled"] != true {
		t.Fatal("global disable erased source choice")
	}
	cli("revision_conflict", "embedding-enable", "1", "stale")
	cli("", "embedding-enable", "2", "cli-global")
	cli("", "source-embedding-disable", "source", "1", "cli-source")
	post("/admin/sources/source/embedding", url.Values{"actor": {"stale"}, "expected_revision": {"1"}, "enabled": {"true"}}, 409, "")
	post("/admin/embedding", url.Values{"actor": {""}, "expected_revision": {"3"}, "enabled": {"true"}}, 400, "")
	cli("", "embedding-disable", "3", "cli-global")
	embeddingRequest := httptest.NewRequest("GET", "http://127.0.0.1/admin/embedding", nil)
	embeddingRequest.Header.Set("Accept", "text/html")
	embeddingResponse := httptest.NewRecorder()
	h.ServeHTTP(embeddingResponse, embeddingRequest)
	if embeddingResponse.Code != 200 || !strings.Contains(embeddingResponse.Body.String(), "Disabilitato") || !strings.Contains(embeddingResponse.Body.String(), "/admin/sources/source/embedding") {
		t.Fatal("embedding form missing", embeddingResponse.Code, embeddingResponse.Body)
	}
	sourcePost := func(action, revision, actor string, status int) {
		t.Helper()
		post("/admin/sources/source/"+action, url.Values{"revision": {revision}, "actor": {actor}}, status, "")
	}
	cli("preview_or_policy_required", "source-enable-collection", "source", "1", "cli")
	sourcePost("enable-collection", "1", "admin", 409)
	if err := reg.RecordPreview(ctx, "source", 1, "synthetic-preview", registry.Evidence{URL: "https://example.test", Locator: "fixture", ObservedAt: time.Now()}); err != nil {
		t.Fatal(err)
	}
	if cli("", "source-enable-collection", "source", "1", "cli-source")["collection_enabled"] != true {
		t.Fatal("source enablement missing")
	}
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "http://127.0.0.1/admin/sources/source", nil)
	r.Header.Set("Accept", "application/json")
	h.ServeHTTP(w, r)
	var view adminSourceView
	if w.Code != 200 || json.Unmarshal(w.Body.Bytes(), &view) != nil || !view.State.CollectionEnabled || view.State.PublicEnabled {
		t.Fatalf("CLI state not visible in admin: %d %s", w.Code, w.Body)
	}
	sourcePost("suspend-collection", "1", "admin-source", 200)
	if cli("", "source-status", "source")["collection_enabled"] != false {
		t.Fatal("admin suspension not visible to CLI")
	}
	sourcePost("enable-collection", "1", "admin-enable", 200)
	cli("", "source-suspend-collection", "source", "1", "cli-suspend")
	w = httptest.NewRecorder()
	h.ServeHTTP(w, r)
	if w.Code != 200 || json.Unmarshal(w.Body.Bytes(), &view) != nil || view.State.CollectionEnabled {
		t.Fatal("CLI suspension not visible in admin")
	}
	// The active revision remains eligible while drafts exist; an obsolete inactive draft conflicts.
	for rev := 1; rev <= 2; rev++ {
		if _, err := reg.AppendConfiguration(ctx, "source", rev, cfg, "draft"); err != nil {
			t.Fatal(err)
		}
	}
	cli("revision_conflict", "source-enable-collection", "source", "2", "cli")
	sourcePost("enable-collection", "2", "admin", 409)
	denied := cfg
	denied.Policy.CollectionPermitted = false
	if _, err := reg.AppendConfiguration(ctx, "source", 3, denied, "policy"); err != nil {
		t.Fatal(err)
	}
	// Synthetic captured preview isolates the policy gate from the missing-preview gate.
	if _, err := p.Exec(ctx, `INSERT INTO registry_events(source_id,revision,kind,actor,evidence) VALUES('source',4,'preview','synthetic-capture','{}')`); err != nil {
		t.Fatal(err)
	}
	cli("preview_or_policy_required", "source-enable-collection", "source", "4", "cli")
	sourcePost("enable-collection", "4", "admin", 409)
	// Profile restrictions apply equally even to an otherwise valid, previewed source.
	if _, err := geo.ConfigureRegion(ctx, "09", 5, domain.RegionConfiguration{MunicipalityDataset: dataset, Profiles: []string{}}, "profile"); err != nil {
		t.Fatal(err)
	}
	cli("unsupported_profile", "source-enable-collection", "source", "1", "cli")
	sourcePost("enable-collection", "1", "admin", 409)
	events, err := reg.Events(ctx, "source")
	if err != nil {
		t.Fatal(err)
	}
	foundEnable, foundDisable := false, false
	for _, e := range events {
		foundEnable = foundEnable || (e.Kind == "collection_enabled" && e.Actor == "cli-source")
		foundDisable = foundDisable || (e.Kind == "collection_disabled" && e.Actor == "admin-source")
	}
	if !foundEnable || !foundDisable {
		t.Fatal("source audit attribution lost")
	}
	history, err := operations.New(p).TerritorialHistory(ctx, "09", "909000", operations.TerritorialHistoryFilter{Kind: "configuration"}, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, e := range history.Entries {
		if strings.HasPrefix(e.ID, "municipality-") && e.Actor == "admin-municipality" {
			found = true
		}
	}
	if !found {
		t.Fatal("municipal lifecycle missing from scoped history")
	}
	other, err := operations.New(p).TerritorialHistory(ctx, "09", "909001", operations.TerritorialHistoryFilter{Kind: "configuration"}, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range other.Entries {
		if strings.HasPrefix(e.ID, "municipality-") {
			t.Fatal("foreign municipality event leaked")
		}
	}
	for _, table := range []string{"processing_jobs", "acquisition_checks", "retained_versions"} {
		var n int
		if err := p.QueryRow(ctx, "SELECT count(*) FROM "+table).Scan(&n); err != nil || n != 0 {
			t.Fatalf("control caused work in %s: %d %v", table, n, err)
		}
	}
	st, err := reg.State(ctx, "source")
	if err != nil || st.PublicEnabled {
		t.Fatal("controls granted public access", err)
	}
	// Role errors must also produce a nonzero executable result without contacting dependencies.
	cmd := exec.Command(binary, "region-status", "09")
	cmd.Env = append(env, "IWA_ROLE=worker")
	if out, err := cmd.CombinedOutput(); err == nil || !strings.Contains(string(out), "admin_role_required") {
		t.Fatalf("role: %v %s", err, out)
	}
}

func TestTerritorialControlFormsRetainPagination(t *testing.T) {
	p := territorialAdminDB(t)
	adoptTestMunicipalities(t, p, "09", 125)
	ops := operations.New(p)
	page, err := ops.Municipalities(context.Background(), "09", operations.MunicipalityFilter{Limit: 50}, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	q := url.Values{"tab": {"municipalities"}, "search": {"Comune"}, "province": {"PR"}, "coverage": {"none"}, "after": {page.Next}}
	path := "/admin/regions/09?" + q.Encode()
	h := HandlerWithAdministration(nil, AdminRuntime{Territories: ops, TerritoryConfig: domain.New(p)})
	w := httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "http://127.0.0.1"+path, nil))
	if w.Code != 200 || !strings.Contains(w.Body.String(), "name=\"return_to\" value=\""+strings.ReplaceAll(path, "&", "&amp;")+"\"") {
		t.Fatalf("pagination return missing: %d %s", w.Code, w.Body)
	}
}
