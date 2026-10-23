package acquisition

import (
	"bytes"
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/Balestrino/italian-weather-alert/internal/backend/documents"
	"github.com/Balestrino/italian-weather-alert/internal/backend/registry"
	"golang.org/x/net/html"
)

// Albo failures fail closed: a caller records the reference as unavailable
// and never infers an act from a partial or unrecognized page.
var (
	ErrAlboUnavailable = errors.New("albo unavailable")
	ErrAlboMarkup      = errors.New("albo markup not recognized")
	ErrAlboBound       = errors.New("albo page bound exceeded")
)

const (
	alboPortlet    = registry.JCityGovAlboPortlet
	alboParam      = "_" + alboPortlet + "_"
	alboActionPath = registry.JCityGovAlboActionPath
	alboPageSize   = 50
)

// AlboUserAgent identifies the service to albo portals.
const AlboUserAgent = "IWA/1.0 (+https://github.com/Balestrino/italian-weather-alert)"

// AlboRecord is one albo publication as listed or detailed. ActDate is zero
// when only list data is known; PublishedFrom is the list's publication start.
type AlboRecord struct {
	RecordID       string
	Category       string
	Subcategory    string
	ActNumber      int
	ActYear        int
	ActDate        time.Time
	RegistryYear   int
	RegistryNumber int
	Object         string
	PublishedFrom  time.Time
	DetailURL      string
}

type AlboAttachment struct {
	Key, MediaType, Title, Description string
	UnsignedURL                        string
}

type AlboDetail struct {
	AlboRecord
	Attachments []AlboAttachment
}

type AlboPage struct {
	Number, Pages, Total int
	Records              []AlboRecord
}

// JCityGovAlbo reads one albo within a reviewed contract. Every request is
// counted against the caller's page budget; list pagination is session-bound.
type JCityGovAlbo struct {
	Contract  *registry.JCityGovAlboContract
	Client    *http.Client
	UserAgent string
	session   *http.Client
	reads     int
}

func (a *JCityGovAlbo) client() (*http.Client, error) {
	if a.session != nil {
		return a.session, nil
	}
	if a.Contract == nil {
		return nil, ErrInvalidConfiguration
	}
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, err
	}
	base := a.Client
	if base == nil {
		base = http.DefaultClient
	}
	session := *base
	session.Jar = jar
	origin, _ := url.Parse(a.Contract.Origin)
	session.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		if !sameHTTPOrigin(origin, req.URL) || len(via) >= 5 {
			return errResourceRedirectOutsideOrigin
		}
		return nil
	}
	a.session = &session
	return a.session, nil
}

// Reads reports how many albo requests this client has made.
func (a *JCityGovAlbo) Reads() int { return a.reads }

func (a *JCityGovAlbo) do(ctx context.Context, method, target string, form url.Values, limit int64) ([]byte, string, error) {
	client, err := a.client()
	if err != nil {
		return nil, "", err
	}
	var body io.Reader
	if form != nil {
		body = strings.NewReader(form.Encode())
	}
	req, err := http.NewRequestWithContext(ctx, method, target, body)
	if err != nil {
		return nil, "", ErrInvalidConfiguration
	}
	if form != nil {
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}
	// Some portals reject the default Go client; identify the service instead.
	agent := a.UserAgent
	if agent == "" {
		agent = AlboUserAgent
	}
	req.Header.Set("User-Agent", agent)
	a.reads++
	resp, err := client.Do(req)
	if err != nil {
		return nil, "", fmt.Errorf("%w: request failed", ErrAlboUnavailable)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, "", fmt.Errorf("%w: HTTP %d", ErrAlboUnavailable, resp.StatusCode)
	}
	data, err := io.ReadAll(io.LimitReader(resp.Body, limit+1))
	if err != nil || int64(len(data)) > limit || len(data) == 0 {
		return nil, "", fmt.Errorf("%w: unreadable or oversized response", ErrAlboUnavailable)
	}
	return data, responseMediaType(resp.Header.Get("Content-Type")), nil
}

func (a *JCityGovAlbo) actionURL(action string) string {
	return a.Contract.Origin + alboActionPath + "?p_p_id=" + alboPortlet + "&p_p_lifecycle=1&p_p_state=pop_up&p_p_mode=view&" + alboParam + "action=" + action
}

// DetailURL is the official, session-free link of an albo record.
func (a *JCityGovAlbo) DetailURL(recordID string) string {
	return a.Contract.Origin + a.Contract.DetailPathPrefix + recordID + "?p_p_state=pop_up"
}

// Page returns list page n (1-based, 50 records) of all publications ordered
// by descending registry number. The list view is opened once per client.
func (a *JCityGovAlbo) Page(ctx context.Context, n int) (AlboPage, error) {
	if a.Contract == nil || n < 1 {
		return AlboPage{}, ErrInvalidConfiguration
	}
	if a.session == nil {
		if _, _, err := a.do(ctx, http.MethodGet, a.Contract.ListURL(), nil, 4<<20); err != nil {
			return AlboPage{}, err
		}
	}
	body, _, err := a.do(ctx, http.MethodPost, a.actionURL("eseguiPaginazione"), url.Values{"hidden_page_size": {strconv.Itoa(alboPageSize)}, "hidden_page_to": {strconv.Itoa(n)}}, 4<<20)
	if err != nil {
		return AlboPage{}, err
	}
	page, err := ParseAlboList(body, a.Contract)
	if err == nil && page.Number != n && page.Pages >= n {
		err = fmt.Errorf("%w: requested page %d, received %d", ErrAlboMarkup, n, page.Number)
	}
	return page, err
}

// Registry looks up the exact registry year and number.
func (a *JCityGovAlbo) Registry(ctx context.Context, year, number int) ([]AlboRecord, error) {
	if a.Contract == nil || year < 1900 || number < 1 {
		return nil, ErrInvalidConfiguration
	}
	if a.session == nil {
		if _, _, err := a.do(ctx, http.MethodGet, a.Contract.ListURL(), nil, 4<<20); err != nil {
			return nil, err
		}
	}
	form := url.Values{"submitButton": {"doFilter"}}
	for _, field := range []string{"oggetto", "mittente", "searchTag", "dataPubblicazioneDa", "dataPubblicazioneA"} {
		form.Set(alboParam+field, "")
	}
	form.Set(alboParam+"simpleSearchEnable", "true")
	form.Set(alboParam+"mostraSoloLista", "true")
	form.Set(alboParam+"annoRegistrazioneDa", strconv.Itoa(year))
	form.Set(alboParam+"numeroRegistrazioneDa", strconv.Itoa(number))
	body, _, err := a.do(ctx, http.MethodPost, a.actionURL("eseguiOrdinamentoLista"), form, 4<<20)
	if err != nil {
		return nil, err
	}
	page, err := ParseAlboList(body, a.Contract)
	return page.Records, err
}

// Window returns the records whose publication start lies within [from,to],
// reading at most bound list pages including the binary-search probes.
func (a *JCityGovAlbo) Window(ctx context.Context, from, to time.Time, bound int) ([]AlboRecord, error) {
	if a.Contract == nil || bound < 1 || to.Before(from) {
		return nil, ErrInvalidConfiguration
	}
	budget := bound
	read := func(n int) (AlboPage, error) {
		if budget == 0 {
			return AlboPage{}, ErrAlboBound
		}
		budget--
		return a.Page(ctx, n)
	}
	first, err := read(1)
	if err != nil {
		return nil, err
	}
	// Find the first page whose oldest record is not newer than to.
	low, high, start := 1, first.Pages, first
	if len(first.Records) > 0 && first.Records[len(first.Records)-1].PublishedFrom.After(to) {
		low = 2
		for low < high {
			mid := (low + high) / 2
			page, err := read(mid)
			if err != nil {
				return nil, err
			}
			if len(page.Records) > 0 && page.Records[len(page.Records)-1].PublishedFrom.After(to) {
				low = mid + 1
			} else {
				high, start = mid, page
			}
		}
		if start.Number != low {
			if start, err = read(low); err != nil {
				return nil, err
			}
		}
	}
	var result []AlboRecord
	for page := start; ; {
		oldest := time.Time{}
		for _, record := range page.Records {
			if !record.PublishedFrom.Before(from) && !record.PublishedFrom.After(to) {
				result = append(result, record)
			}
			oldest = record.PublishedFrom
		}
		if len(page.Records) == 0 || oldest.Before(from) || page.Number >= page.Pages {
			return result, nil
		}
		if page, err = read(page.Number + 1); err != nil {
			return nil, err
		}
	}
}

// Detail reads a record detail page and returns it with its raw markup.
func (a *JCityGovAlbo) Detail(ctx context.Context, recordID string) (AlboDetail, []byte, error) {
	if a.Contract == nil || !a.Contract.AllowsDetail(a.DetailURL(recordID)) {
		return AlboDetail{}, nil, ErrInvalidConfiguration
	}
	body, _, err := a.do(ctx, http.MethodGet, a.DetailURL(recordID), nil, 4<<20)
	if err != nil {
		return AlboDetail{}, nil, err
	}
	detail, err := ParseAlboDetail(body, a.Contract)
	if err == nil && detail.RecordID != recordID {
		err = fmt.Errorf("%w: detail for another record", ErrAlboMarkup)
	}
	detail.DetailURL = a.DetailURL(recordID)
	return detail, body, err
}

// Download fetches an unsigned attachment URL taken from a parsed detail.
func (a *JCityGovAlbo) Download(ctx context.Context, raw string) ([]byte, string, error) {
	if a.Contract == nil || !a.allowsDownload(raw) {
		return nil, "", ErrInvalidConfiguration
	}
	return a.do(ctx, http.MethodGet, raw, nil, documents.MaxAcquisitionBytes)
}

func (a *JCityGovAlbo) allowsDownload(raw string) bool { return a.Contract.AllowsDownload(raw) }

var (
	alboPaging = regexp.MustCompile(`Pagina (\d+) di (\d+) \((\d+) risultati\)`)
	alboPair   = regexp.MustCompile(`^(\d{4})/(\d+)$`)
	alboAtob   = regexp.MustCompile(`atob\('([A-Za-z0-9+/=]+)'\)`)
)

// ParseAlboList reads one list or search-result page.
func ParseAlboList(body []byte, contract *registry.JCityGovAlboContract) (AlboPage, error) {
	root, err := html.Parse(bytes.NewReader(body))
	if err != nil {
		return AlboPage{}, fmt.Errorf("%w: list", ErrAlboMarkup)
	}
	page := AlboPage{}
	match := alboPaging.FindSubmatch(body)
	if match == nil {
		return page, fmt.Errorf("%w: list pagination", ErrAlboMarkup)
	}
	page.Number, _ = strconv.Atoi(string(match[1]))
	page.Pages, _ = strconv.Atoi(string(match[2]))
	page.Total, _ = strconv.Atoi(string(match[3]))
	var failed error
	walkElements(root, func(n *html.Node) {
		if failed != nil || n.Data != "tr" || !hasClass(n, "master-detail-list-line") {
			return
		}
		record := AlboRecord{RecordID: alboAttr(n, "data-id")}
		cells := map[string]*html.Node{}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			if c.Type == html.ElementNode && c.Data == "td" {
				cells[strings.Fields(alboAttr(c, "class") + " x")[0]] = c
			}
		}
		registryCell, act, category, object, period := cells["annonumeroregistrazione"], cells["annonumero"], cells["categoria"], cells["oggetto"], cells["periodo-pubblicazione"]
		if record.RecordID == "" || strings.Trim(record.RecordID, "0123456789") != "" || registryCell == nil || act == nil || category == nil || object == nil || period == nil {
			failed = fmt.Errorf("%w: list row", ErrAlboMarkup)
			return
		}
		var ok bool
		if record.RegistryYear, record.RegistryNumber, ok = alboYearNumber(alboText(registryCell)); !ok {
			failed = fmt.Errorf("%w: registry number", ErrAlboMarkup)
			return
		}
		record.ActYear, record.ActNumber, _ = alboYearNumber(alboText(act))
		record.Category, record.Subcategory = alboCategory(category)
		record.Object = alboText(object)
		dates := strings.Fields(alboText(period))
		if len(dates) == 0 {
			failed = fmt.Errorf("%w: publication period", ErrAlboMarkup)
			return
		}
		if record.PublishedFrom, err = time.Parse("02/01/2006", dates[0]); err != nil {
			failed = fmt.Errorf("%w: publication date", ErrAlboMarkup)
			return
		}
		if contract != nil {
			record.DetailURL = contract.Origin + contract.DetailPathPrefix + record.RecordID + "?p_p_state=pop_up"
		}
		page.Records = append(page.Records, record)
	})
	if failed == nil && len(page.Records) == 0 && page.Total != 0 {
		failed = fmt.Errorf("%w: rows missing", ErrAlboMarkup)
	}
	return page, failed
}

// ParseAlboDetail reads the metadata table and the attachment table of a
// detail page. Only unsigned downloads of the contract's origin are kept.
func ParseAlboDetail(body []byte, contract *registry.JCityGovAlboContract) (AlboDetail, error) {
	root, err := html.Parse(bytes.NewReader(body))
	if err != nil {
		return AlboDetail{}, fmt.Errorf("%w: detail", ErrAlboMarkup)
	}
	detail := AlboDetail{}
	fields := map[string]string{}
	walkElements(root, func(n *html.Node) {
		if n.Data == "table" && alboAttr(n, "data-id-pubblicazione") != "" {
			detail.RecordID = alboAttr(n, "data-id-pubblicazione")
		}
		if n.Data != "tr" {
			return
		}
		if class := alboAttr(n, "class"); strings.HasPrefix(class, "ap-") {
			var cells []*html.Node
			for c := n.FirstChild; c != nil; c = c.NextSibling {
				if c.Type == html.ElementNode && c.Data == "td" {
					cells = append(cells, c)
				}
			}
			if len(cells) == 2 {
				fields[strings.TrimPrefix(class, "ap-")] = alboText(cells[1])
			}
		}
		if key := alboAttr(n, "data-chiave-allegato"); key != "" {
			var cells []string
			for c := n.FirstChild; c != nil; c = c.NextSibling {
				if c.Type == html.ElementNode && c.Data == "td" {
					cells = append(cells, alboText(c))
				}
			}
			attachment := AlboAttachment{Key: key, MediaType: alboAttr(n, "data-mimetype")}
			if len(cells) >= 2 {
				attachment.Title, attachment.Description = cells[0], cells[1]
			}
			walkElements(n, func(link *html.Node) {
				if link.Data != "a" || alboAttr(link, "title") != "Download versione non firmata" {
					return
				}
				if m := alboAtob.FindStringSubmatch(alboAttr(link, "onclick")); m != nil {
					if decoded, decodeErr := base64.StdEncoding.DecodeString(m[1]); decodeErr == nil {
						attachment.UnsignedURL = string(decoded)
					}
				}
			})
			detail.Attachments = append(detail.Attachments, attachment)
		}
	})
	var ok bool
	if detail.RegistryYear, ok = alboInt(fields["annoRegistro"]); !ok {
		return detail, fmt.Errorf("%w: registry year", ErrAlboMarkup)
	}
	if detail.RegistryNumber, ok = alboInt(fields["numeroRegistro"]); !ok {
		return detail, fmt.Errorf("%w: registry number", ErrAlboMarkup)
	}
	detail.Category, detail.Subcategory, detail.Object = fields["categoria"], fields["sottocategoria"], fields["oggetto"]
	detail.ActYear, _ = alboInt(fields["annoAtto"])
	detail.ActNumber, _ = alboInt(fields["numeroAtto"])
	if value := fields["dataAtto"]; value != "" {
		if detail.ActDate, err = time.Parse("02/01/2006", value); err != nil {
			return detail, fmt.Errorf("%w: act date", ErrAlboMarkup)
		}
	}
	if detail.RecordID == "" || detail.Category == "" || detail.Object == "" {
		return detail, fmt.Errorf("%w: detail fields", ErrAlboMarkup)
	}
	if contract != nil {
		probe := &JCityGovAlbo{Contract: contract}
		for i := range detail.Attachments {
			if !probe.allowsDownload(detail.Attachments[i].UnsignedURL) {
				detail.Attachments[i].UnsignedURL = ""
			}
		}
	}
	return detail, nil
}

func alboYearNumber(value string) (int, int, bool) {
	m := alboPair.FindStringSubmatch(strings.TrimSpace(value))
	if m == nil {
		return 0, 0, false
	}
	year, _ := strconv.Atoi(m[1])
	number, _ := strconv.Atoi(m[2])
	return year, number, number > 0
}

func alboInt(value string) (int, bool) {
	n, err := strconv.Atoi(strings.TrimSpace(value))
	return n, err == nil && n > 0
}

func alboCategory(cell *html.Node) (string, string) {
	var category, subcategory string
	walkElements(cell, func(n *html.Node) {
		switch {
		case n.Data == "span" && hasClass(n, "categoria_categoria"):
			category = alboText(n)
		case n.Data == "span" && hasClass(n, "categoria_sottocategoria"):
			subcategory = alboText(n)
		}
	})
	if category == "" {
		category = alboText(cell)
	}
	return category, subcategory
}

func alboAttr(n *html.Node, key string) string {
	for _, a := range n.Attr {
		if a.Key == key {
			return a.Val
		}
	}
	return ""
}

func alboText(n *html.Node) string {
	var b strings.Builder
	var visit func(*html.Node)
	visit = func(n *html.Node) {
		if n.Type == html.TextNode {
			b.WriteString(n.Data)
			b.WriteByte(' ')
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			visit(c)
		}
	}
	visit(n)
	return strings.Join(strings.Fields(b.String()), " ")
}
