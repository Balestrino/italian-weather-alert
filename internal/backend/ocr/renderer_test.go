package ocr

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestPopplerRendererUsesPrivateFilesAndOrdersPages(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("shell fixture is POSIX-only")
	}
	dir := t.TempDir()
	command := filepath.Join(dir, "fake-pdftoppm")
	script := []byte("#!/bin/sh\nlast=''\nfor value in \"$@\"; do last=$value; done\nprintf two > \"${last}-2.png\"\nprintf one > \"${last}-1.png\"\n")
	if err := os.WriteFile(command, script, 0700); err != nil {
		t.Fatal(err)
	}
	pages, err := (PopplerRenderer{Command: command, DPI: 144}).Render(context.Background(), []byte("synthetic PDF"))
	if err != nil || len(pages) != 2 || pages[0].Number != 1 || string(pages[0].Bytes) != "one" || string(pages[1].Bytes) != "two" {
		t.Fatalf("unexpected rendered pages: %#v %v", pages, err)
	}
}
