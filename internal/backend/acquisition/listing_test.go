package acquisition

import (
	"fmt"
	"testing"
	"time"

	"github.com/Balestrino/italian-weather-alert/internal/backend/registry"
)

func TestCascinaListingUnpaddedDays(t *testing.T) {
	cfg := registry.Configuration{Discovery: registry.Discovery{DocumentPathPrefixes: []string{"/it/news/"}, ListingItemClass: "card-body", ListingDateClass: "h5", ListingDateLayout: "2 Jan 2006", ListingDateLocale: "it"}}
	for _, tc := range []struct {
		day  string
		want int
	}{{"8", 8}, {"08", 8}, {"18", 18}, {"32", 0}} {
		t.Run(tc.day, func(t *testing.T) {
			body := []byte(fmt.Sprintf(`<div class="card-body"><a href="/it/news?type=3">Avvisi</a><span class="h5 card-pretitle">%s ottobre 2024</span><a href="/it/news/42/avviso"><h1 class="h5">Avviso</h1></a></div>`, tc.day))
			items := discoverDocuments("https://municipal.example/it/news-category/42", body, cfg, []Link{{URL: "https://municipal.example/it/news/42/avviso"}})
			if len(items) != 1 {
				t.Fatalf("discovery: %#v", items)
			}
			if tc.want == 0 {
				if items[0].PublicationDate != nil {
					t.Fatal("invalid date accepted")
				}
				return
			}
			want := time.Date(2024, 10, tc.want, 0, 0, 0, 0, time.UTC)
			if items[0].PublicationDate == nil || !items[0].PublicationDate.Equal(want) {
				t.Fatalf("date: %#v", items[0])
			}
		})
	}
}

func TestCalcinaiaListingDateExtraction(t *testing.T) {
	body := []byte(`<div class="card-body"><span class="data">20 Ago 2026</span><a href="/novita/avviso-di-criticita">Avviso</a></div>`)
	cfg := registry.Configuration{Discovery: registry.Discovery{DocumentPathPrefixes: []string{"/novita/"}, ListingItemClass: "card-body", ListingDateClass: "data", ListingDateLayout: "02 Jan 2006", ListingDateLocale: "it"}}
	items := discoverDocuments("https://www.comune.calcinaia.pi.it/tipi-di-notizia/notizie", body, cfg, nil)
	want := time.Date(2026, 8, 20, 0, 0, 0, 0, time.UTC)
	if len(items) != 1 || items[0].URL != "https://www.comune.calcinaia.pi.it/novita/avviso-di-criticita" || items[0].PublicationDate == nil || !items[0].PublicationDate.Equal(want) {
		t.Fatalf("unexpected extracted listing: %#v", items)
	}
}

func TestLivornoListingFullItalianMonthAndDateClass(t *testing.T) {
	body := []byte(`<div class="card-body"><div class="card-pretitle card-pretitle-link">Comunicati</div><span class="h5 card-pretitle">10 settembre 2026</span><a href="/it/news/133640/allerta-arancione" class="link-detail"><h1 class="h5 card-title">Allerta</h1></a></div>`)
	cfg := registry.Configuration{Discovery: registry.Discovery{DocumentPathPrefixes: []string{"/it/news/"}, ListingItemClass: "card-body", ListingDateClass: "h5", ListingDateLayout: "02 Jan 2006", ListingDateLocale: "it"}}
	items := discoverDocuments("https://www.comune.livorno.it/it/news-category/133640", body, cfg, nil)
	want := time.Date(2026, 9, 10, 0, 0, 0, 0, time.UTC)
	if len(items) != 1 || items[0].URL != "https://www.comune.livorno.it/it/news/133640/allerta-arancione" || items[0].PublicationDate == nil || !items[0].PublicationDate.Equal(want) {
		t.Fatalf("unexpected Livorno listing: %#v", items)
	}
}

func TestPisaListingDateAfterLabel(t *testing.T) {
	body := []byte(`<div class="card card-teaser"><div class="card-body"><div class="text-paragraph-card"><p>Codice giallo fino al 17 settembre</p><p><strong>Data di pubblicazione:</strong> 16/09/2026</p></div></div><a href="/Novita/Notizie/Allerta-meteo">Ulteriori dettagli</a></div>`)
	cfg := registry.Configuration{Discovery: registry.Discovery{DocumentPathPrefixes: []string{"/Novita/Notizie/"}, ListingItemClass: "card", ListingDateClass: "text-paragraph-card", ListingDateLabel: "Data di pubblicazione:", ListingDateLayout: "02/01/2006", ListingDateLocale: "it"}}
	items := discoverDocuments("https://www.comune.pisa.it/Novita/Notizie", body, cfg, nil)
	want := time.Date(2026, 9, 16, 0, 0, 0, 0, time.UTC)
	if len(items) != 1 || items[0].URL != "https://www.comune.pisa.it/Novita/Notizie/Allerta-meteo" || items[0].PublicationDate == nil || !items[0].PublicationDate.Equal(want) {
		t.Fatalf("unexpected Pisa listing: %#v", items)
	}
}
