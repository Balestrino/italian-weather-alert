package ocr

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
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

func TestPopplerRendererRetriesOnlyOversizedPhysicalPageOnce(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("shell fixture is POSIX-only")
	}
	for _, oversizedRetry := range []bool{false, true} {
		t.Run(fmt.Sprint(oversizedRetry), func(t *testing.T) {
			dir := t.TempDir()
			command := filepath.Join(dir, "fake-pdftoppm")
			trace := filepath.Join(dir, "trace")
			retryBody := `printf recovered > "${last}-2.png"`
			if oversizedRetry {
				retryBody = `dd if=/dev/zero of="${last}-2.png" bs=1048576 count=6 2>/dev/null`
			}
			script := "#!/bin/sh\necho \"$*\" >> '" + trace + "'\nlast=''\nfor value in \"$@\"; do last=$value; done\nif [ \"$3\" = 72 ]; then\n" + retryBody + "\nelse\nprintf normal > \"${last}-1.png\"\ndd if=/dev/zero of=\"${last}-2.png\" bs=1048576 count=6 2>/dev/null\nfi\n"
			if err := os.WriteFile(command, []byte(script), 0700); err != nil {
				t.Fatal(err)
			}
			pages, err := (PopplerRenderer{Command: command}).Render(context.Background(), []byte("synthetic PDF"))
			if oversizedRetry {
				if !errors.Is(err, ErrRasterization) {
					t.Fatalf("oversized replacement accepted: %v", err)
				}
			} else if err != nil || len(pages) != 2 || pages[1].Number != 2 || string(pages[1].Bytes) != "recovered" {
				t.Fatalf("page recovery failed: %#v %v", pages, err)
			}
			calls, err := os.ReadFile(trace)
			if err != nil || strings.Count(string(calls), "\n") != 2 || !strings.Contains(string(calls), "-r 72 -gray -f 2 -l 2") {
				t.Fatalf("unbounded or wrong-page retry: %q %v", calls, err)
			}
		})
	}
}
