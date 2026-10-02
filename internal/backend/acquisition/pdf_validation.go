package acquisition

import (
	"bytes"
	"context"
	"errors"
	"os"
	"os/exec"
	"regexp"
	"time"
)

var errInvalidPDF = errors.New("invalid attachment PDF")
var pdfPages = regexp.MustCompile(`(?m)^Pages:\s+[1-9][0-9]*\s*$`)

// Poppler is already part of the application image. Validation does not perform
// OCR or require readable text: scanned PDFs remain legitimate evidence.
func validateAttachmentPDF(ctx context.Context, page Page) error {
	if page.MediaType != "application/pdf" && page.MediaType != "application/octet-stream" && page.MediaType != "" {
		return errInvalidPDF
	}
	if !bytes.HasPrefix(page.HTML, []byte("%PDF-")) {
		return errInvalidPDF
	}
	tail := page.HTML
	if len(tail) > 1024 {
		tail = tail[len(tail)-1024:]
	}
	if !bytes.Contains(tail, []byte("%%EOF")) {
		return errInvalidPDF
	}
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "pdfinfo", "-")
	cmd.Env = []string{"PATH=" + os.Getenv("PATH"), "LANG=C", "LC_ALL=C"}
	cmd.Stdin = bytes.NewReader(page.HTML)
	out, err := cmd.Output()
	if err != nil || !pdfPages.Match(out) {
		return errInvalidPDF
	}
	return nil
}
