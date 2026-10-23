package acquisition

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/Balestrino/italian-weather-alert/internal/backend/registry"
)

const fixtureOrigin = "https://albo.example.it"

func alboContract(origin string) *registry.JCityGovAlboContract {
	return &registry.JCityGovAlboContract{MunicipalityISTAT: "050004", ServesSource: "calcinaia-municipal", Origin: origin, ListPath: "/web/trasparenza/papca-ap/-/papca/igrid/1/2",
		DetailPathPrefix: "/web/trasparenza/papca-ap/-/papca/display/", Categories: map[string]string{"ordinanza": "ORDINANZE"}, PageBound: 12}
}

func alboFixture(t *testing.T, name string) []byte {
	t.Helper()
	body, err := os.ReadFile("testdata/jcitygov-albo/" + name)
	if err != nil {
		t.Fatal(err)
	}
	return body
}

var fixtureAtob = regexp.MustCompile(`atob\('([A-Za-z0-9+/=]+)'\)`)

// rehost points fixture links, including base64 download links, at origin.
func rehost(body []byte, origin string) []byte {
	out := fixtureAtob.ReplaceAllStringFunc(string(body), func(m string) string {
		raw, _ := base64.StdEncoding.DecodeString(fixtureAtob.FindStringSubmatch(m)[1])
		return "atob('" + base64.StdEncoding.EncodeToString([]byte(strings.ReplaceAll(string(raw), fixtureOrigin, origin))) + "')"
	})
	return []byte(strings.ReplaceAll(out, fixtureOrigin, origin))
}

func TestParseAlboListAndDetail(t *testing.T) {
	contract := alboContract(fixtureOrigin)
	page, err := ParseAlboList(alboFixture(t, "list-page-3.html"), contract)
	if err != nil || page.Number != 3 || page.Pages != 120 || page.Total != 5962 || len(page.Records) != 5 {
		t.Fatalf("list page: %+v %v", page, err)
	}
	ordinance, decree := page.Records[2], page.Records[3]
	if ordinance.RecordID != "9000660" || ordinance.Category != "ORDINANZE" || ordinance.Subcategory != "ORDINANZA" || ordinance.ActYear != 2026 || ordinance.ActNumber != 38 ||
		ordinance.RegistryNumber != 660 || !ordinance.PublishedFrom.Equal(time.Date(2026, 9, 9, 0, 0, 0, 0, time.UTC)) || !strings.HasPrefix(ordinance.Object, "EMERGENZA DI PROTEZIONE CIVILE") ||
		ordinance.DetailURL != fixtureOrigin+"/web/trasparenza/papca-ap/-/papca/display/9000660?p_p_state=pop_up" {
		t.Fatalf("ordinance row: %+v", ordinance)
	}
	if decree.Category != "DECRETI" || decree.Object != ordinance.Object || decree.ActNumber != 9 {
		t.Fatalf("decree row: %+v", decree)
	}
	if generic := page.Records[1]; generic.Category != "GENERICO" || generic.ActNumber != 0 {
		t.Fatalf("act without number: %+v", generic)
	}
	for name, want := range map[string]struct {
		number, registry int
		date             time.Time
	}{"detail-38.html": {38, 660, time.Date(2026, 9, 9, 0, 0, 0, 0, time.UTC)}, "detail-34.html": {34, 629, time.Date(2026, 8, 20, 0, 0, 0, 0, time.UTC)}} {
		detail, err := ParseAlboDetail(alboFixture(t, name), contract)
		if err != nil || detail.Category != "ORDINANZE" || detail.ActNumber != want.number || detail.RegistryNumber != want.registry || !detail.ActDate.Equal(want.date) || len(detail.Attachments) != 1 {
			t.Fatalf("%s: %+v %v", name, detail, err)
		}
		attachment := detail.Attachments[0]
		if attachment.Description != "Documento principale" || !strings.Contains(attachment.UnsignedURL, "downloadSigned=false") || !strings.HasPrefix(attachment.UnsignedURL, fixtureOrigin+"/") {
			t.Fatalf("%s attachment: %+v", name, attachment)
		}
	}
	foreign, err := ParseAlboDetail(alboFixture(t, "detail-38.html"), alboContract("https://other.example.it"))
	if err != nil || foreign.Attachments[0].UnsignedURL != "" {
		t.Fatalf("download of another origin kept: %+v %v", foreign.Attachments, err)
	}
	empty, err := ParseAlboList(alboFixture(t, "search-empty.html"), contract)
	if err != nil || empty.Total != 0 || len(empty.Records) != 0 {
		t.Fatalf("empty search: %+v %v", empty, err)
	}
}

func TestParseAlboFailsClosed(t *testing.T) {
	contract := alboContract(fixtureOrigin)
	list := string(alboFixture(t, "list-page-3.html"))
	detail := string(alboFixture(t, "detail-38.html"))
	for name, body := range map[string]string{
		"truncated list":     list[:strings.Index(list, "<tr class=\"master-detail-list-line")] + "<p>Pagina 3 di 120 (5962 risultati)</p>",
		"renamed cell":       strings.Replace(list, `class="annonumeroregistrazione number"`, `class="registro"`, 1),
		"no pagination":      strings.Replace(list, "Pagina 3 di 120", "Page 3 of 120", 1),
		"bad date":           strings.Replace(list, "09/09/2026 <br />", "9 settembre <br />", 1),
		"no registry":        strings.Replace(detail, `class="ap-numeroRegistro"`, `class="ap-altro"`, 1),
		"no record id":       strings.Replace(detail, `data-id-pubblicazione="9000660"`, ``, 1),
		"unparseable date":   strings.Replace(detail, " 09/09/2026 </span></td> </tr>\n</table>", " settembre </span></td> </tr>\n</table>", 1),
		"detail without obj": strings.Replace(detail, `class="ap-oggetto"`, `class="ap-nota"`, 1),
	} {
		t.Run(name, func(t *testing.T) {
			var err error
			if strings.Contains(name, "list") || strings.Contains(name, "cell") || strings.Contains(name, "pagination") || name == "bad date" {
				_, err = ParseAlboList([]byte(body), contract)
			} else {
				_, err = ParseAlboDetail([]byte(body), contract)
			}
			if !errors.Is(err, ErrAlboMarkup) {
				t.Fatalf("accepted changed markup: %v", err)
			}
		})
	}
}

// alboServer serves a synthetic albo of n records, one per day backwards
// from newest, with session-bound pagination like the reviewed portal.
func alboServer(t *testing.T, n int, newest time.Time) (*httptest.Server, *int) {
	t.Helper()
	requests := 0
	row := func(i int) string {
		day := newest.AddDate(0, 0, -i)
		number := 1000 - i
		return fmt.Sprintf(`<tr class="master-detail-list-line" data-id="%d"> <td class="annonumeroregistrazione number">2026/%d</td> <td class="annonumero number">2026/%d</td> <td class="categoria text"><span class="categoria categoria_categoria">ORDINANZE</span><span class="categoria categoria_sottocategoria">ORDINANZA</span></td> <td class="oggetto text">ATTO %d</td> <td class="periodo-pubblicazione date"> %s <br /> 31/12/2031 </td> </tr>`, 5000000+number, number, number, number, day.Format("02/01/2006"))
	}
	render := func(page int) string {
		pages := (n + alboPageSize - 1) / alboPageSize
		var rows []string
		for i := (page - 1) * alboPageSize; i < page*alboPageSize && i < n; i++ {
			rows = append(rows, row(i))
		}
		return fmt.Sprintf("<html><body><table>%s</table><p>Pagina %d di %d (%d risultati)</p></body></html>", strings.Join(rows, "\n"), page, pages, n)
	}
	var server *httptest.Server
	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests++
		if r.UserAgent() != AlboUserAgent {
			http.Error(w, "unidentified client", http.StatusForbidden)
			return
		}
		session, err := r.Cookie("JSESSIONID")
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/web/trasparenza/papca-ap/-/papca/igrid/1/2":
			http.SetCookie(w, &http.Cookie{Name: "JSESSIONID", Value: "fixture", Path: "/"})
			fmt.Fprint(w, render(1))
		case r.Method == http.MethodPost && r.URL.Query().Get(alboParam+"action") == "eseguiPaginazione":
			page, _ := strconv.Atoi(r.FormValue("hidden_page_to"))
			if err != nil || session.Value != "fixture" || r.FormValue("hidden_page_size") != "50" || page < 1 {
				http.Error(w, "session required", http.StatusInternalServerError)
				return
			}
			fmt.Fprint(w, render(page))
		case r.Method == http.MethodPost && r.URL.Query().Get(alboParam+"action") == "eseguiOrdinamentoLista":
			if err != nil || r.FormValue(alboParam+"numeroRegistrazioneDa") != "660" {
				w.Write(rehost(alboFixture(t, "search-empty.html"), server.URL))
				return
			}
			w.Write(rehost(alboFixture(t, "search-single.html"), server.URL))
		case strings.HasPrefix(r.URL.Path, "/web/trasparenza/papca-ap/-/papca/display/9000660"):
			w.Write(rehost(alboFixture(t, "detail-38.html"), server.URL))
		case r.URL.Path == alboActionPath && r.URL.Query().Get("p_p_resource_id") == "downloadAllegato":
			w.Header().Set("Content-Type", "application/pdf")
			w.Write([]byte("%PDF-synthetic ordinance"))
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(server.Close)
	return server, &requests
}

func TestJCityGovAlboWindowRegistryDetailAndDownload(t *testing.T) {
	ctx := context.Background()
	newest := time.Date(2026, 10, 22, 0, 0, 0, 0, time.UTC)
	server, requests := alboServer(t, 1000, newest)
	albo := &JCityGovAlbo{Contract: alboContract(server.URL), Client: server.Client()}
	from, to := time.Date(2026, 8, 13, 0, 0, 0, 0, time.UTC), time.Date(2026, 8, 20, 0, 0, 0, 0, time.UTC)
	records, err := albo.Window(ctx, from, to, 12)
	if err != nil || len(records) != 8 || !records[0].PublishedFrom.Equal(to) || !records[7].PublishedFrom.Equal(from) {
		t.Fatalf("window: %d records %v", len(records), err)
	}
	if albo.Reads() > 13 {
		t.Fatalf("window used %d requests", albo.Reads())
	}
	bounded := &JCityGovAlbo{Contract: alboContract(server.URL), Client: server.Client()}
	if _, err = bounded.Window(ctx, from, to, 3); !errors.Is(err, ErrAlboBound) {
		t.Fatalf("page bound not enforced: %v", err)
	}
	recent := &JCityGovAlbo{Contract: alboContract(server.URL), Client: server.Client()}
	if records, err = recent.Window(ctx, newest.AddDate(0, 0, -2), newest, 12); err != nil || len(records) != 3 || recent.Reads() != 2 {
		t.Fatalf("recent window: %d %v reads=%d", len(records), err, recent.Reads())
	}
	found, err := albo.Registry(ctx, 2026, 660)
	if err != nil || len(found) != 1 || found[0].ActNumber != 38 || found[0].RecordID != "9000660" {
		t.Fatalf("registry search: %+v %v", found, err)
	}
	if found, err = albo.Registry(ctx, 2026, 1); err != nil || len(found) != 0 {
		t.Fatalf("empty registry search: %+v %v", found, err)
	}
	detail, raw, err := albo.Detail(ctx, "9000660")
	if err != nil || len(raw) == 0 || detail.ActNumber != 38 || detail.DetailURL != server.URL+"/web/trasparenza/papca-ap/-/papca/display/9000660?p_p_state=pop_up" {
		t.Fatalf("detail: %+v %v", detail, err)
	}
	body, media, err := albo.Download(ctx, detail.Attachments[0].UnsignedURL)
	if err != nil || media != "application/pdf" || !strings.HasPrefix(string(body), "%PDF") {
		t.Fatalf("download: %q %q %v", body, media, err)
	}
	signed := strings.Replace(detail.Attachments[0].UnsignedURL, "downloadSigned=false", "downloadSigned=true", 1)
	for _, raw := range []string{signed, "https://other.example.it" + alboActionPath + "?p_p_id=" + alboPortlet, server.URL + "/elsewhere.pdf"} {
		if _, _, err = albo.Download(ctx, raw); !errors.Is(err, ErrInvalidConfiguration) {
			t.Fatalf("download outside contract allowed: %s %v", raw, err)
		}
	}
	if _, _, err = albo.Detail(ctx, "../1"); !errors.Is(err, ErrInvalidConfiguration) {
		t.Fatalf("detail outside contract allowed: %v", err)
	}
	if _, _, err = albo.Detail(ctx, "1"); !errors.Is(err, ErrAlboUnavailable) {
		t.Fatalf("missing detail not unavailable: %v", err)
	}
	if *requests == 0 {
		t.Fatal("no requests served")
	}
}

func TestJCityGovAlboRequiresSession(t *testing.T) {
	server, _ := alboServer(t, 10, time.Date(2026, 10, 22, 0, 0, 0, 0, time.UTC))
	albo := &JCityGovAlbo{Contract: alboContract(server.URL), Client: server.Client()}
	albo.session = server.Client() // a session without the list cookie
	if _, err := albo.Page(context.Background(), 1); !errors.Is(err, ErrAlboUnavailable) {
		t.Fatalf("page without session accepted: %v", err)
	}
}

func TestAlboPreviewReadsOnlyTheFirstPage(t *testing.T) {
	server, requests := alboServer(t, 120, time.Date(2026, 10, 22, 0, 0, 0, 0, time.UTC))
	contract := alboContract(server.URL)
	cfg := registry.Configuration{JCityGovAlbo: contract, AccessMethod: registry.JCityGovAlboAccess, URL: contract.ListURL(), Sections: []string{contract.ListURL()}}
	engine := &Engine{Albo: server.Client()}
	report, state, err := engine.acquire(context.Background(), "albo", 1, cfg)
	if err != nil || !state.reachable || !state.contentRecognized || len(report.Documents) != 0 || len(report.Sections) != 1 || report.Sections[0].SectionURL != contract.ListURL() || *requests != 2 {
		t.Fatalf("preview %+v %+v %v requests=%d", report, state, err, *requests)
	}
	broken := &Engine{Albo: server.Client()}
	cfg.JCityGovAlbo = alboContract(server.URL + "/missing")
	if _, state, err = broken.acquire(context.Background(), "albo", 1, cfg); err == nil || state.errorCode != "albo_unavailable" {
		t.Fatalf("unreachable albo preview: %+v %v", state, err)
	}
}
