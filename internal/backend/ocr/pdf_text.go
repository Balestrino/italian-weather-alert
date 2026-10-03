package ocr

import (
	"bytes"
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"
)

const PDFTextVersion = "poppler-pdf-text-v1"
const maxPDFTextBytes = 6 << 20

var ErrNativeText = errors.New("complete native PDF text unavailable")

type TextPage struct {
	Number int
	Text   string
}

type TextExtractor interface {
	Extract(context.Context, []byte) ([]TextPage, error)
}

// PopplerText accepts only reviewed text-only documents. Image detection rejects
// scans and mixed PDFs; vector content still requires the source's review.
type PopplerText struct{}

func (PopplerText) Extract(ctx context.Context, body []byte) ([]TextPage, error) {
	if len(body) == 0 || len(body) > 32<<20 {
		return nil, ErrNativeText
	}
	ctx, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel()
	dir, err := os.MkdirTemp("", "iwa-pdf-text-")
	if err != nil {
		return nil, ErrNativeText
	}
	defer os.RemoveAll(dir)
	path := filepath.Join(dir, "original.pdf")
	if os.WriteFile(path, body, 0600) != nil {
		return nil, ErrNativeText
	}
	info, err := popplerOutput(ctx, 64<<10, "pdfinfo", path)
	if err != nil {
		return nil, ErrNativeText
	}
	count, encrypted := 0, ""
	for _, line := range strings.Split(string(info), "\n") {
		key, value, ok := strings.Cut(line, ":")
		if !ok {
			continue
		}
		switch strings.TrimSpace(key) {
		case "Pages":
			count, _ = strconv.Atoi(strings.TrimSpace(value))
		case "Encrypted":
			encrypted = strings.TrimSpace(value)
		}
	}
	if count < 1 || count > 500 || encrypted != "no" {
		return nil, ErrNativeText
	}
	images, err := popplerOutput(ctx, 1<<20, "pdfimages", "-list", path)
	if err != nil || !noPDFImages(images) {
		return nil, ErrNativeText
	}
	text, err := popplerOutput(ctx, maxPDFTextBytes, "pdftotext", "-layout", "-enc", "UTF-8", "-eol", "unix", path, "-")
	if err != nil || !utf8.Valid(text) {
		return nil, ErrNativeText
	}
	parts := strings.Split(string(text), "\f")
	if len(parts) != count+1 || strings.TrimSpace(parts[count]) != "" {
		return nil, ErrNativeText
	}
	pages := make([]TextPage, 0, count)
	for i, text := range parts[:count] {
		pages = append(pages, TextPage{Number: i + 1, Text: strings.TrimSpace(text)})
	}
	if !completeTextPages(pages) {
		return nil, ErrNativeText
	}
	return pages, nil
}

func completeTextPages(pages []TextPage) bool {
	if len(pages) == 0 || len(pages) > 500 {
		return false
	}
	total := 0
	for i, p := range pages {
		total += len(p.Text)
		if p.Number != i+1 || !utf8.ValidString(p.Text) || total > maxPDFTextBytes {
			return false
		}
		letters := 0
		for _, r := range p.Text {
			if r == unicode.ReplacementChar || unicode.IsControl(r) && r != '\n' && r != '\t' && r != '\r' {
				return false
			}
			if unicode.IsLetter(r) {
				letters++
			}
		}
		if letters < 20 {
			return false
		}
	}
	return true
}

func noPDFImages(body []byte) bool {
	lines := strings.Split(strings.TrimSpace(string(body)), "\n")
	if len(lines) != 2 || !strings.HasPrefix(strings.TrimSpace(lines[0]), "page") || !strings.Contains(lines[0], "num") || !strings.Contains(lines[0], "type") {
		return false
	}
	separator := strings.TrimSpace(lines[1])
	return len(separator) > 10 && strings.Trim(separator, "-") == ""
}

type boundedOutput struct {
	bytes.Buffer
	limit int
}

func (b *boundedOutput) Write(p []byte) (int, error) {
	if b.Len()+len(p) > b.limit {
		return 0, ErrNativeText
	}
	return b.Buffer.Write(p)
}

func popplerOutput(ctx context.Context, limit int, name string, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Env = append(os.Environ(), "LC_ALL=C")
	out := &boundedOutput{limit: limit}
	cmd.Stdout = out
	if err := cmd.Run(); err != nil {
		return nil, err
	}
	return out.Bytes(), nil
}
