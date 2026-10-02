package classification

import (
	"bytes"
	"strings"

	"github.com/Balestrino/italian-weather-alert/internal/backend/registry"
	"golang.org/x/net/html"
)

// articleHTML requires every reviewed selector to match exactly one nonempty
// container. Otherwise the full page is kept, so layout drift cannot hide text.
func articleHTML(body []byte, selectors []string) string {
	if len(selectors) == 0 {
		return visibleHTML(body)
	}
	root, err := html.Parse(bytes.NewReader(body))
	if err != nil {
		return visibleHTML(body)
	}
	selected := map[*html.Node]bool{}
	for _, selector := range selectors {
		if !registry.ValidBodySelector(selector) {
			return visibleHTML(body)
		}
		var matches []*html.Node
		var walk func(*html.Node)
		walk = func(n *html.Node) {
			if n.Type == html.ElementNode && matchesBodySelector(n, selector) {
				matches = append(matches, n)
			}
			for c := n.FirstChild; c != nil; c = c.NextSibling {
				walk(c)
			}
		}
		walk(root)
		if len(matches) != 1 || nodeVisibleText(matches[0]) == "" {
			return visibleHTML(body)
		}
		selected[matches[0]] = true
	}
	var texts []string
	var collect func(*html.Node)
	collect = func(n *html.Node) {
		if selected[n] {
			texts = append(texts, nodeVisibleText(n))
			return // A selected ancestor already includes its descendants.
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			collect(c)
		}
	}
	collect(root)
	return normalize(strings.Join(texts, " "))
}

func nodeVisibleText(n *html.Node) string {
	var b bytes.Buffer
	if html.Render(&b, n) != nil {
		return ""
	}
	return visibleHTML(b.Bytes())
}

func matchesBodySelector(n *html.Node, selector string) bool {
	index := strings.IndexAny(selector, ".#")
	if index < 0 {
		return n.Data == selector
	}
	if index > 0 && n.Data != selector[:index] {
		return false
	}
	key, token := "class", selector[index+1:]
	if selector[index] == '#' {
		key = "id"
	}
	for _, a := range n.Attr {
		if a.Key != key {
			continue
		}
		if key == "id" {
			return a.Val == token
		}
		for _, class := range strings.Fields(a.Val) {
			if class == token {
				return true
			}
		}
	}
	return false
}
