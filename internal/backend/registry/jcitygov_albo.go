package registry

import (
	"net/url"
	"path"
	"regexp"
	"slices"
	"strings"
)

// JCityGovAlboContract scopes a municipal albo pretorio published on the
// JCityGov "papca" platform. It authorizes only the retrieval of acts that a
// notice of the served primary source explicitly references; the albo is never
// listed exhaustively, interpreted or used as an alert channel.
type JCityGovAlboContract struct {
	MunicipalityISTAT string            `json:"municipality_istat"`
	ServesSource      string            `json:"serves_source"`
	Origin            string            `json:"origin"`
	ListPath          string            `json:"list_path"`
	DetailPathPrefix  string            `json:"detail_path_prefix"`
	Categories        map[string]string `json:"categories"`
	PageBound         int               `json:"page_bound"`
}

const JCityGovAlboAccess = "jcitygov-albo"

// The JCityGov albo portlet and the action path serving its downloads.
const (
	JCityGovAlboPortlet    = "jcitygovalbopubblicazioni_WAR_jcitygovalbiportlet"
	JCityGovAlboActionPath = "/web/trasparenza/papca-ap"
)

// ReferenceKinds are the act kinds a notice reference can name.
var ReferenceKinds = []string{"ordinanza", "decreto", "delibera", "determina"}

var alboCategory = regexp.MustCompile(`^[A-Z][A-Z ]{1,40}$`)

// ListURL is the paged "Storico atti" list of the albo.
func (p *JCityGovAlboContract) ListURL() string { return p.Origin + p.ListPath }

// AllowsDetail reports whether raw is a record detail page of this albo.
func (p *JCityGovAlboContract) AllowsDetail(raw string) bool {
	u, err := url.Parse(raw)
	if err != nil || u.Scheme+"://"+u.Host != p.Origin || u.User != nil || u.Fragment != "" || !strings.HasPrefix(u.Path, p.DetailPathPrefix) {
		return false
	}
	id := strings.TrimPrefix(u.Path, p.DetailPathPrefix)
	return id != "" && strings.Trim(id, "0123456789") == "" && len(id) <= 18
}

// AllowsDownload reports whether raw is an unsigned attachment download of this albo.
func (p *JCityGovAlboContract) AllowsDownload(raw string) bool {
	u, err := url.Parse(raw)
	if err != nil || u.Scheme+"://"+u.Host != p.Origin || u.Path != JCityGovAlboActionPath || u.User != nil || u.Fragment != "" {
		return false
	}
	q := u.Query()
	param := "_" + JCityGovAlboPortlet + "_"
	return q.Get("p_p_id") == JCityGovAlboPortlet && q.Get("p_p_resource_id") == "downloadAllegato" && q.Get(param+"downloadSigned") == "false" && q.Get(param+"id") != ""
}

func (p *JCityGovAlboContract) Valid(c Configuration) bool {
	if p == nil || !municipalityISTAT.MatchString(p.MunicipalityISTAT) || strings.TrimSpace(p.ServesSource) == "" ||
		c.AccessMethod != JCityGovAlboAccess || c.CittadinoInformato != nil || c.RegionalProduct != nil || c.Attachments != nil ||
		c.ExpectedPublication != nil || c.LocalProcessing != nil || c.InferenceEligibility != nil ||
		c.Policy.CopiesPermitted || !c.Policy.CollectionPermitted || !c.Policy.RetentionPermitted || p.PageBound < 1 || p.PageBound > 20 {
		return false
	}
	u, err := url.Parse(p.Origin)
	if err != nil || u.Scheme != "https" || u.Host == "" || u.User != nil || u.Path != "" || u.RawQuery != "" || u.Fragment != "" || strings.ContainsAny(u.Host, "*%\\") {
		return false
	}
	for _, value := range []string{p.ListPath, p.DetailPathPrefix} {
		if !strings.HasPrefix(value, "/") || strings.ContainsAny(value, "?#%\\") {
			return false
		}
	}
	if path.Clean(p.ListPath) != p.ListPath || p.DetailPathPrefix == "/" || path.Clean(p.DetailPathPrefix)+"/" != p.DetailPathPrefix ||
		c.URL != p.ListURL() || !slices.Equal(c.Sections, []string{p.ListURL()}) || len(p.Categories) == 0 {
		return false
	}
	for kind, category := range p.Categories {
		if !slices.Contains(ReferenceKinds, kind) || !alboCategory.MatchString(category) {
			return false
		}
	}
	r := c.Referral
	return r != nil && validEvidence(r.Evidence) && r.Destination == c.URL && r.Territory == p.MunicipalityISTAT &&
		r.ProductID == "municipal" && slices.Equal(r.Sections, c.Sections) && strings.TrimSpace(r.Context) != ""
}
