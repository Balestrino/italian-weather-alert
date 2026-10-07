package ocr

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

const maxPageImageBytes = 5 << 20

// PopplerRenderingVersion identifies the immutable rasterization policy in caches.
const PopplerRenderingVersion = "png-144dpi-bounded-gray-v2"

type Renderer interface {
	Render(context.Context, []byte) ([]PageImage, error)
}

// PopplerRenderer invokes pdftoppm without a shell. The retained PDF is only
// materialized in a private temporary directory and removed after rendering.
type PopplerRenderer struct {
	Command string
	DPI     int
}

func (r PopplerRenderer) Render(ctx context.Context, pdf []byte) ([]PageImage, error) {
	command := r.Command
	if command == "" {
		command = "pdftoppm"
	}
	dpi := r.DPI
	if dpi == 0 {
		dpi = 144
	}
	if dpi < 72 || dpi > 300 || len(pdf) == 0 {
		return nil, ErrInvalid
	}
	path, err := exec.LookPath(command)
	if err != nil {
		return nil, ErrRendererUnavailable
	}
	dir, err := os.MkdirTemp("", "iwa-ocr-")
	if err != nil {
		return nil, ErrRendererUnavailable
	}
	defer os.RemoveAll(dir)
	input := filepath.Join(dir, "original.pdf")
	if err = os.WriteFile(input, pdf, 0600); err != nil {
		return nil, ErrRendererUnavailable
	}
	prefix := filepath.Join(dir, "page")
	cmd := exec.CommandContext(ctx, path, "-png", "-r", fmt.Sprint(dpi), input, prefix)
	cmd.Stdout = nil
	cmd.Stderr = nil
	if err = cmd.Run(); err != nil {
		if errors.Is(ctx.Err(), context.Canceled) || errors.Is(ctx.Err(), context.DeadlineExceeded) {
			return nil, ctx.Err()
		}
		return nil, ErrRasterization
	}
	paths, err := filepath.Glob(prefix + "-*.png")
	if err != nil || len(paths) == 0 || len(paths) > 500 {
		return nil, ErrRasterization
	}
	type renderedPage struct {
		path   string
		number int
	}
	entries := make([]renderedPage, 0, len(paths))
	for _, name := range paths {
		suffix := strings.TrimSuffix(strings.TrimPrefix(name, prefix+"-"), ".png")
		number, parseErr := strconv.Atoi(suffix)
		if parseErr != nil || number < 1 {
			return nil, ErrRasterization
		}
		entries = append(entries, renderedPage{path: name, number: number})
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].number < entries[j].number })
	pages := make([]PageImage, 0, len(entries))
	var total int64
	for index, entry := range entries {
		if entry.number != index+1 {
			return nil, ErrRasterization
		}
		info, statErr := os.Stat(entry.path)
		if statErr != nil {
			return nil, ErrRasterization
		}
		if info.Size() > maxPageImageBytes && dpi > 72 {
			// Retry only this physical page, once, as grayscale at minimum DPI.
			// The provider limit still applies to the replacement image.
			retryPrefix := filepath.Join(dir, fmt.Sprintf("retry-%d", entry.number))
			retry := exec.CommandContext(ctx, path, "-png", "-r", "72", "-gray", "-f", fmt.Sprint(entry.number), "-l", fmt.Sprint(entry.number), input, retryPrefix)
			if err = retry.Run(); err != nil {
				if ctx.Err() != nil {
					return nil, ctx.Err()
				}
				return nil, ErrRasterization
			}
			matches, matchErr := filepath.Glob(retryPrefix + "-*.png")
			if matchErr != nil || len(matches) != 1 {
				return nil, ErrRasterization
			}
			physicalPage, parseErr := strconv.Atoi(strings.TrimSuffix(strings.TrimPrefix(matches[0], retryPrefix+"-"), ".png"))
			if parseErr != nil || physicalPage != entry.number {
				return nil, ErrRasterization
			}
			entry.path = matches[0]
			info, statErr = os.Stat(entry.path)
		}
		if statErr != nil || info.Size() <= 0 || info.Size() > maxPageImageBytes {
			return nil, ErrRasterization
		}
		body, readErr := os.ReadFile(entry.path)
		if readErr != nil || len(body) == 0 {
			return nil, ErrRasterization
		}
		total += int64(len(body))
		if len(body) > maxPageImageBytes || total > 256<<20 || !strings.HasSuffix(entry.path, ".png") {
			return nil, ErrRasterization
		}
		pages = append(pages, PageImage{Number: index + 1, MediaType: "image/png", Bytes: body})
	}
	return pages, nil
}
