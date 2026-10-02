package backoffice

import (
	"context"
	"encoding/base64"
	"errors"
	"github.com/Balestrino/italian-weather-alert/internal/backend/domain"
	"github.com/Balestrino/italian-weather-alert/internal/backend/registry"
	"github.com/Balestrino/italian-weather-alert/internal/backend/territory"
	"io"
	"net/http"
	"net/url"
	"slices"
	"strconv"
	"strings"
	"time"
)

type territorySetupPage struct {
	Region                                                domain.Region
	Base, Actor, CSV, SHA256, Municipality, AuthorityName string
	Preview                                               *domain.MunicipalityImportPreview
	Metadata                                              domain.MunicipalityImport
}

func territoryForm(w http.ResponseWriter, r *http.Request, allowed ...string) (result error) {
	defer func() {
		if result != nil && r.MultipartForm != nil {
			r.MultipartForm.RemoveAll()
		}
	}()
	r.Body = http.MaxBytesReader(w, r.Body, 2*domain.MaxMunicipalityImportBytes+(1<<20))
	var err error
	if strings.HasPrefix(r.Header.Get("Content-Type"), "multipart/form-data") {
		err = r.ParseMultipartForm(1 << 20)
	} else if strings.HasPrefix(r.Header.Get("Content-Type"), "application/x-www-form-urlencoded") {
		err = r.ParseForm()
	} else {
		return domain.ErrInvalid
	}
	if err != nil {
		return domain.ErrInvalid
	}
	for k, v := range r.PostForm {
		if !slices.Contains(allowed, k) || len(v) != 1 {
			return domain.ErrInvalid
		}
	}
	return nil
}
func setupRegion(r *http.Request, a AdminRuntime) (domain.Region, error) {
	if a.TerritoryConfig == nil {
		return domain.Region{}, registryUnavailable
	}
	return a.TerritoryConfig.Region(r.Context(), r.PathValue("region"))
}
func adminTerritorySetupRoutes(mux *http.ServeMux, a AdminRuntime) {
	mux.HandleFunc("GET /admin/regions/{region}/setup", func(w http.ResponseWriter, r *http.Request) {
		region, err := setupRegion(r, a)
		if err != nil {
			territoryFailure(w, r, err)
			return
		}
		renderUI(w, territoryTemplates, "territory-setup", territorySetupPage{Region: region, Base: "/admin/regions/" + region.Code, Metadata: domain.MunicipalityImport{VerifiedAt: time.Now()}})
	})
	for _, action := range []string{"preview", "adopt"} {
		mux.HandleFunc("POST /admin/regions/{region}/import/"+action, func(w http.ResponseWriter, r *http.Request) {
			if err := territoryForm(w, r, "actor", "expected_revision", "official_url", "version", "verified_at", "expected_count", "complete", "completeness_evidence", "csv", "csv_base64", "sha256"); err != nil {
				territoryFailure(w, r, err)
				return
			}
			if r.MultipartForm != nil {
				defer r.MultipartForm.RemoveAll()
			}
			region, err := setupRegion(r, a)
			if err != nil {
				territoryFailure(w, r, err)
				return
			}
			revision, err := territoryRevision(r.PostForm.Get("expected_revision"))
			if err != nil {
				territoryFailure(w, r, err)
				return
			}
			if revision != region.Revision {
				territoryFailure(w, r, domain.ErrConflict)
				return
			}
			actor := strings.TrimSpace(r.PostForm.Get("actor"))
			verified, err := time.Parse(time.DateOnly, r.PostForm.Get("verified_at"))
			if err != nil || actor == "" {
				territoryFailure(w, r, domain.ErrInvalid)
				return
			}
			count, err := strconv.Atoi(r.PostForm.Get("expected_count"))
			if err != nil || (r.PostForm.Get("complete") != "" && r.PostForm.Get("complete") != "true") {
				territoryFailure(w, r, domain.ErrInvalid)
				return
			}
			meta := domain.MunicipalityImport{Region: region.Code, OfficialURL: r.PostForm.Get("official_url"), Version: r.PostForm.Get("version"), VerifiedAt: verified, ExpectedCount: count, Complete: r.PostForm.Get("complete") == "true", CompletenessEvidence: r.PostForm.Get("completeness_evidence")}
			csv := r.PostForm.Get("csv")
			if encoded := r.PostForm.Get("csv_base64"); encoded != "" {
				decoded, e := base64.StdEncoding.DecodeString(encoded)
				if e != nil || csv != "" || r.MultipartForm != nil || len(decoded) > domain.MaxMunicipalityImportBytes {
					territoryFailure(w, r, domain.ErrInvalid)
					return
				}
				csv = string(decoded)
			}
			if r.MultipartForm != nil {
				if len(r.MultipartForm.File) != 1 || len(r.MultipartForm.File["file"]) != 1 || csv != "" {
					territoryFailure(w, r, domain.ErrInvalid)
					return
				}
				f, _, e := r.FormFile("file")
				if e != nil {
					territoryFailure(w, r, domain.ErrInvalid)
					return
				}
				body, e := io.ReadAll(io.LimitReader(f, domain.MaxMunicipalityImportBytes+1))
				f.Close()
				if e != nil {
					territoryFailure(w, r, domain.ErrInvalid)
					return
				}
				csv = string(body)
			}
			preview, err := domain.PreviewMunicipalities([]byte(csv), meta)
			if err != nil {
				territoryFailure(w, r, err)
				return
			}
			if action == "adopt" {
				_, err = a.TerritoryConfig.AdoptMunicipalities(r.Context(), []byte(csv), meta, r.PostForm.Get("sha256"), revision, actor)
				if err != nil {
					territoryFailure(w, r, err)
					return
				}
				http.Redirect(w, r, "/admin/regions/"+region.Code+"/setup", http.StatusSeeOther)
				return
			}
			renderUI(w, territoryTemplates, "territory-import-preview", territorySetupPage{Region: region, Base: "/admin/regions/" + region.Code, Actor: actor, CSV: csv, SHA256: preview.SHA256, Preview: &preview, Metadata: meta})
		})
	}
	mux.HandleFunc("POST /admin/regions/{region}/configuration", func(w http.ResponseWriter, r *http.Request) {
		if err := territoryForm(w, r, "actor", "expected_revision", "municipality_dataset", "zone_dataset", "postal_dataset", "municipal_profile", "regional_profile", "comparison_profile"); err != nil {
			territoryFailure(w, r, err)
			return
		}
		region, err := setupRegion(r, a)
		if err != nil {
			territoryFailure(w, r, err)
			return
		}
		revision, err := territoryRevision(r.PostForm.Get("expected_revision"))
		if err != nil {
			territoryFailure(w, r, err)
			return
		}
		cfg := domain.RegionConfiguration{MunicipalityDataset: r.PostForm.Get("municipality_dataset"), ZoneDataset: r.PostForm.Get("zone_dataset"), PostalDataset: r.PostForm.Get("postal_dataset"), Profiles: []string{}}
		for field, profile := range map[string]string{"municipal_profile": "municipal-html", "regional_profile": "toscana-cfr", "comparison_profile": "dpc-comparison"} {
			v := r.PostForm.Get(field)
			if v != "" && v != "true" {
				territoryFailure(w, r, domain.ErrInvalid)
				return
			}
			if v == "true" {
				cfg.Profiles = append(cfg.Profiles, profile)
			}
		}
		slices.Sort(cfg.Profiles)
		if _, err = a.TerritoryConfig.ConfigureRegion(r.Context(), region.Code, revision, cfg, r.PostForm.Get("actor")); err != nil {
			territoryFailure(w, r, err)
			return
		}
		http.Redirect(w, r, "/admin/regions/"+region.Code+"/setup", http.StatusSeeOther)
	})
	mux.HandleFunc("POST /admin/regions/{region}/enabled", func(w http.ResponseWriter, r *http.Request) {
		if err := territoryForm(w, r, "actor", "expected_revision", "enabled", "return_to"); err != nil {
			territoryFailure(w, r, err)
			return
		}
		region, err := setupRegion(r, a)
		if err != nil {
			territoryFailure(w, r, err)
			return
		}
		revision, err := territoryRevision(r.PostForm.Get("expected_revision"))
		raw := r.PostForm.Get("enabled")
		if err != nil || (raw != "true" && raw != "false") {
			territoryFailure(w, r, domain.ErrInvalid)
			return
		}
		back, err := territoryControlReturn(r.PostForm.Get("return_to"), region.Code, "configuration")
		if err != nil {
			territoryFailure(w, r, err)
			return
		}
		ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
		defer cancel()
		if _, err = a.TerritoryConfig.SetRegionEnabled(ctx, region.Code, revision, raw == "true", r.PostForm.Get("actor")); err != nil {
			territoryFailure(w, r, err)
			return
		}
		http.Redirect(w, r, back, http.StatusSeeOther)
	})
	mux.HandleFunc("POST /admin/regions/{region}/municipalities/{istat}/enabled", func(w http.ResponseWriter, r *http.Request) {
		if err := territoryForm(w, r, "actor", "expected_revision", "enabled", "return_to"); err != nil {
			territoryFailure(w, r, err)
			return
		}
		region, err := setupRegion(r, a)
		if err != nil {
			territoryFailure(w, r, err)
			return
		}
		revision, err := territoryRevision(r.PostForm.Get("expected_revision"))
		raw := r.PostForm.Get("enabled")
		if err != nil || (raw != "true" && raw != "false") {
			territoryFailure(w, r, domain.ErrInvalid)
			return
		}
		back, err := territoryControlReturn(r.PostForm.Get("return_to"), region.Code, "municipalities")
		if err != nil {
			territoryFailure(w, r, err)
			return
		}
		ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
		defer cancel()
		if _, err = a.TerritoryConfig.SetMunicipalityEnabled(ctx, region.Code, r.PathValue("istat"), revision, raw == "true", r.PostForm.Get("actor")); err != nil {
			territoryFailure(w, r, err)
			return
		}
		http.Redirect(w, r, back, http.StatusSeeOther)
	})
	adminTerritoryNewSourceRoutes(mux, a)
}

type sourceIdentityCreator interface {
	CreateSourceWithIdentity(context.Context, registry.Source, registry.Configuration, string, registry.Authority, registry.Channel) error
}

func adminTerritoryNewSourceRoutes(mux *http.ServeMux, a AdminRuntime) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		region, err := setupRegion(r, a)
		if err != nil {
			territoryFailure(w, r, err)
			return
		}
		istat := r.URL.Query().Get("municipality")
		name := region.Name
		territoryID := region.Code
		authorityID := "region-" + region.Code
		if istat != "" {
			if a.Territories == nil {
				territoryFailure(w, r, registryUnavailable)
				return
			}
			m, e := a.Territories.Municipality(r.Context(), region.Code, istat)
			if e != nil {
				territoryFailure(w, r, e)
				return
			}
			if m.Historical {
				territoryFailure(w, r, domain.ErrInvalid)
				return
			}
			name = "Comune di " + m.Name
			territoryID = istat
			authorityID = "municipality-" + istat
		}
		base := "/admin/regions/" + region.Code
		if r.Method == "GET" {
			renderUI(w, territoryTemplates, "territory-new-source", territorySetupPage{Region: region, Base: base, Municipality: istat, AuthorityName: name})
			return
		}
		if err = territoryForm(w, r, "actor", "source_id", "authority_name", "authority_url", "url", "sections", "product", "attribution"); err != nil {
			territoryFailure(w, r, err)
			return
		}
		creator, ok := a.Registry.(sourceIdentityCreator)
		if !ok {
			territoryFailure(w, r, registryUnavailable)
			return
		}
		id := strings.TrimSpace(r.PostForm.Get("source_id"))
		if id == "" || len(id) > 120 || strings.ContainsAny(id, "/\\?# ") {
			territoryFailure(w, r, domain.ErrInvalid)
			return
		}
		product := r.PostForm.Get("product")
		if istat != "" {
			if product != "municipal" {
				territoryFailure(w, r, domain.ErrInvalid)
				return
			}
		} else if !slices.Contains([]string{"criticality", "vigilance", "monitoring", "dpc_comparison"}, product) {
			territoryFailure(w, r, domain.ErrInvalid)
			return
		}
		sections := []string{}
		for _, line := range strings.Split(r.PostForm.Get("sections"), "\n") {
			if line = strings.TrimSpace(line); line != "" {
				sections = append(sections, line)
			}
		}
		source := registry.Source{ID: id, AuthorityID: authorityID, ChannelID: "source-" + id, ProductID: product, Territory: territoryID}
		cfg := registry.Configuration{URL: r.PostForm.Get("url"), Sections: sections, AccessMethod: "html", Attribution: r.PostForm.Get("attribution"), Limitations: []string{"Fonte in bozza: evidenze e confini di acquisizione da verificare"}}
		authority := registry.Authority{ID: authorityID, Name: r.PostForm.Get("authority_name"), OfficialURL: r.PostForm.Get("authority_url")}
		channel := registry.Channel{ID: source.ChannelID, PublisherID: authorityID, Platform: "official", URL: cfg.URL}
		if err = creator.CreateSourceWithIdentity(r.Context(), source, cfg, r.PostForm.Get("actor"), authority, channel); err != nil {
			if errors.Is(err, registry.ErrConflict) {
				err = domain.ErrConflict
			} else if errors.Is(err, registry.ErrInvalid) || errors.Is(err, territory.ErrAssociation) {
				err = domain.ErrInvalid
			}
			territoryFailure(w, r, err)
			return
		}
		if istat != "" {
			base += "/municipalities/" + istat
		}
		http.Redirect(w, r, "/admin/sources/"+url.PathEscape(id)+"?return_to="+url.QueryEscape(base+"?tab=configuration"), http.StatusSeeOther)
	}
	mux.HandleFunc("GET /admin/regions/{region}/sources/new", handler)
	mux.HandleFunc("POST /admin/regions/{region}/sources/new", handler)
}

// Restrict native-form returns to the same territorial workspace. Validate before mutation.
func territoryControlReturn(raw, region, fallbackTab string) (string, error) {
	if raw == "" {
		return "/admin/regions/" + region + "?tab=" + fallbackTab, nil
	}
	u, err := url.Parse(raw)
	if err != nil || u.IsAbs() || u.Host != "" || u.Fragment != "" || u.RawPath != "" || len(raw) > 4096 {
		return "", domain.ErrInvalid
	}
	q, err := url.ParseQuery(u.RawQuery)
	if err != nil {
		return "", domain.ErrInvalid
	}
	for k, v := range q {
		if len(v) != 1 {
			return "", domain.ErrInvalid
		}
		if u.Path == "/admin/regions" {
			if k != "state" || !slices.Contains([]string{"all", "enabled", "disabled"}, v[0]) {
				return "", domain.ErrInvalid
			}
		} else {
			if u.Path != "/admin/regions/"+region || !slices.Contains([]string{"tab", "search", "province", "coverage", "after"}, k) {
				return "", domain.ErrInvalid
			}
		}
	}
	if u.Path == "/admin/regions" {
		return u.String(), nil
	}
	if u.Path != "/admin/regions/"+region || !slices.Contains([]string{"configuration", "municipalities"}, q.Get("tab")) {
		return "", domain.ErrInvalid
	}
	if q.Get("tab") == "configuration" && len(q) != 1 {
		return "", domain.ErrInvalid
	}
	if len(q.Get("search")) > 200 || len(q.Get("province")) > 100 || len(q.Get("after")) > 2048 || !slices.Contains([]string{"", "none", "configured", "collecting"}, q.Get("coverage")) {
		return "", domain.ErrInvalid
	}
	return u.String(), nil
}
