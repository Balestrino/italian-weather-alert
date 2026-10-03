package registry

import (
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// CittadinoInformatoContract is an explicit municipality-scoped opt-in. The
// operator's recorded choice is acquisition policy evidence, not a license.
type CittadinoInformatoContract struct {
	MunicipalityISTAT string   `json:"municipality_istat"`
	MunicipalitySlug  string   `json:"municipality_slug"`
	Publisher         string   `json:"publisher"`
	Updates           bool     `json:"updates"`
	Risks             bool     `json:"risks"`
	PageSize          int      `json:"page_size"`
	UpdatesSince      string   `json:"updates_since,omitempty"` // Explicit API display-date filter, not operative validity.
	AttachmentPaths   []string `json:"attachment_paths,omitempty"`
}

const CittadinoInformatoAccess = "cittadino-informato-api"

var cittadinoSlug = regexp.MustCompile(`^[a-z][a-z0-9]*(?:-[a-z0-9]+)*$`)
var municipalityISTAT = regexp.MustCompile(`^[0-9]{6}$`)

func (p *CittadinoInformatoContract) BaseURL() string {
	return "https://cittadinoinformato.it/" + p.MunicipalitySlug + "/"
}

func (p *CittadinoInformatoContract) APIBase() string {
	return p.BaseURL() + "wp-json/cittadino/v2/"
}

func (p *CittadinoInformatoContract) Valid(c Configuration) bool {
	if p == nil || !municipalityISTAT.MatchString(p.MunicipalityISTAT) || !cittadinoSlug.MatchString(p.MunicipalitySlug) ||
		p.Publisher != "comune_"+strings.ReplaceAll(p.MunicipalitySlug, "-", "_") ||
		c.URL != p.BaseURL() || c.AccessMethod != CittadinoInformatoAccess || (!p.Updates && !p.Risks) ||
		p.PageSize < 1 || p.PageSize > 100 || c.Discovery.MaxPagesPerSection < 1 || c.Discovery.MaxPagesPerSection > 1000 ||
		c.Discovery.MaxDocuments < 1 || c.Discovery.MaxDocuments > 10000 || c.Discovery.BootstrapDays < 1 || c.Discovery.BootstrapDays > 366 ||
		c.RegionalProduct != nil || c.Attachments != nil || c.Policy.CopiesPermitted || !c.Policy.CollectionPermitted || !c.Policy.RetentionPermitted {
		return false
	}
	if p.UpdatesSince != "" {
		if _, err := time.Parse("2006-01-02", p.UpdatesSince); err != nil || !p.Updates {
			return false
		}
	}
	sections := []string{}
	if p.Updates {
		sections = append(sections, p.BaseURL()+"aggiornamenti/")
	}
	if p.Risks {
		sections = append(sections, p.BaseURL())
	}
	r := c.Referral
	if !sameSections(c.Sections, sections) || r == nil || !validEvidence(r.Evidence) || r.Destination != c.URL ||
		r.Territory != p.MunicipalityISTAT || r.ProductID != "municipal" || !sameSections(r.Sections, sections) || strings.TrimSpace(r.Context) == "" {
		return false
	}
	seen := map[string]bool{}
	for _, path := range p.AttachmentPaths {
		// Shared uploads require an explicit subdirectory; another municipality's
		// directory is never a dependency permission for this source.
		if !strings.HasSuffix(path, "/") || strings.ContainsAny(path, "%?#\\") || strings.Contains(path, "..") || seen[path] ||
			(!strings.HasPrefix(path, "/"+p.MunicipalitySlug+"/") && (!strings.HasPrefix(path, "/app/uploads/") || path == "/app/uploads/")) {
			return false
		}
		seen[path] = true
	}
	return true
}

func (p *CittadinoInformatoContract) AllowsAttachment(raw string) bool {
	u, err := url.Parse(raw)
	if err != nil || u.Scheme != "https" || u.Host != "cittadinoinformato.it" || u.User != nil || u.Fragment != "" || u.RawQuery != "" || u.Path != u.EscapedPath() || strings.Contains(u.Path, "..") {
		return false
	}
	for _, prefix := range p.AttachmentPaths {
		if strings.HasPrefix(u.Path, prefix) && len(u.Path) > len(prefix) && strings.HasSuffix(strings.ToLower(u.Path), ".pdf") {
			return true
		}
	}
	return false
}

// AllowsOriginal also checks query scoping. A shared REST host is not permission
// to retain another municipality's response under this source's identity.
func (p *CittadinoInformatoContract) AllowsOriginal(raw string) bool {
	u, err := url.Parse(raw)
	if err != nil || u.User != nil || u.Fragment != "" || !strings.HasPrefix(raw, p.APIBase()) {
		return false
	}
	path := strings.TrimPrefix(u.Path, "/"+p.MunicipalitySlug+"/wp-json/cittadino/v2/")
	if p.Risks && (path == "rischi/oggi" || path == "rischi/domani") {
		return u.RawQuery == ""
	}
	if !p.Updates {
		return false
	}
	q, err := url.ParseQuery(u.RawQuery)
	if err != nil {
		return false
	}
	if path == "aggiornamenti" {
		page, err := strconv.Atoi(q.Get("page"))
		if err != nil || page < 1 || q.Get("comune") != p.MunicipalitySlug || q.Get("ente") != p.Publisher || q.Get("per_page") != strconv.Itoa(p.PageSize) {
			return false
		}
		if p.UpdatesSince != "" {
			date, _ := time.Parse("2006-01-02", p.UpdatesSince)
			if len(q["data_inizio"]) != 1 || q.Get("data_inizio") != date.Format("02/01/2006") {
				return false
			}
			q.Del("data_inizio")
		}
		for _, key := range []string{"comune", "ente", "per_page", "page"} {
			if len(q[key]) != 1 {
				return false
			}
			q.Del(key)
		}
		return len(q) == 0
	}
	id, err := strconv.ParseInt(strings.TrimPrefix(path, "aggiornamenti/"), 10, 64)
	return err == nil && id > 0 && raw == p.APIBase()+"aggiornamenti/"+strconv.FormatInt(id, 10)+"?comune="+url.QueryEscape(p.MunicipalitySlug)
}
