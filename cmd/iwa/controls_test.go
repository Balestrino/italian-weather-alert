package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/Balestrino/italian-weather-alert/internal/backend/acquisition"
	"github.com/Balestrino/italian-weather-alert/internal/backend/domain"
)

func TestControlArgumentsAndRole(t *testing.T) {
	for _, args := range [][]string{
		{"embedding-status"}, {"embedding-enable", "0", "operator"}, {"embedding-disable", "1", "operator"},
		{"source-embedding-status", "source"}, {"source-embedding-enable", "source", "0", "operator"}, {"source-embedding-disable", "source", "1", "operator"},
		{"source-status", "source"}, {"source-enable-collection", "source", "1", "operator"},
		{"source-suspend-collection", "source", "1", "operator"}, {"region-status", "09"},
		{"region-enable", "09", "0", "operator"}, {"region-disable", "09", "3", "operator"},
		{"municipality-status", "09", "050004"}, {"municipality-enable", "09", "050004", "0", "operator"},
		{"municipality-disable", "09", "050004", "2", "operator"},
	} {
		c, ok, err := parseControlCommand(args)
		if !ok || err != nil {
			t.Fatal(args, c, ok, err)
		}
		if err = executeControl(context.Background(), nil, "public", c, io.Discard); !errors.Is(err, errControlRole) {
			t.Fatal("role bypass", err)
		}
	}
	for _, args := range [][]string{
		{"embedding-status", "extra"}, {"embedding-enable", "-1", "actor"}, {"embedding-disable", "0", " "}, {"source-embedding-enable", "source", "x", "actor"},
		{"source-status"}, {"source-status", "s", "extra"}, {"source-enable-collection", "s", "0", "actor"},
		{"source-enable-collection", "s", "1", " "}, {"source-enable-collection", "s", "-1", "actor"},
		{"region-enable", "9", "0", "actor"}, {"region-enable", "09", "x", "actor"},
		{"municipality-enable", "09", "050004", "0"}, {"municipality-status", "03", "../050004"},
	} {
		_, ok, err := parseControlCommand(args)
		if !ok || !errors.Is(err, errControlArguments) {
			t.Fatal("invalid accepted", args, ok, err)
		}
	}
	if _, ok, err := parseControlCommand([]string{"worker"}); ok || err != nil {
		t.Fatal("existing command intercepted")
	}
	if controlErrorCode(errors.New("private credential detail")) != "control_unavailable" {
		t.Fatal("unsafe error")
	}
}

func TestDevelopmentPublicationCommands(t *testing.T) {
	for _, command := range []string{"municipality-development-publication-enable", "municipality-development-publication-disable", "municipality-development-publication-status"} {
		args := []string{command, "09", "050004"}
		if !strings.HasSuffix(command, "-status") {
			args = append(args, "0", "operator")
		}
		c, handled, err := parseControlCommand(args)
		if err != nil || !handled {
			t.Fatal(c, handled, err)
		}
		err = executeControlEnvironment(context.Background(), nil, "admin", "production", c, io.Discard)
		if !errors.Is(err, domain.ErrDevelopmentOnly) {
			t.Fatal("strict control", err)
		}
	}
}

func TestUnavailableTargetCommands(t *testing.T) {
	for _, args := range [][]string{
		{"source-unavailable-targets", "calcinaia-municipal"},
		{"source-tracked-targets", "calcinaia-municipal"},
		{"source-act-references", "calcinaia-municipal"},
		{"source-unavailable-decide", "calcinaia-municipal", "1", "exclude_stale_unavailable", "operator"},
		{"source-unavailable-decide", "calcinaia-municipal", "2", "resume", "operator"},
	} {
		c, ok, err := parseControlCommand(args)
		if !ok || err != nil || c.Source != "calcinaia-municipal" {
			t.Fatal(args, c, ok, err)
		}
		if err = executeControl(context.Background(), nil, "public", c, io.Discard); !errors.Is(err, errControlRole) {
			t.Fatal("role bypass", err)
		}
	}
	c, _, _ := parseControlCommand([]string{"source-unavailable-decide", "s", "3", "resume", "operator"})
	if c.Revision != 3 || c.Decision != "resume" || c.Actor != "operator" {
		t.Fatalf("decision arguments: %+v", c)
	}
	for _, args := range [][]string{
		{"source-unavailable-targets"}, {"source-unavailable-targets", "s", "extra"},
		{"source-tracked-targets"}, {"source-tracked-targets", "s", "1", "operator"},
		{"source-act-references"}, {"source-act-references", "s", "1", "operator"},
		{"source-unavailable-decide", "s", "0", "resume", "operator"}, {"source-unavailable-decide", "s", "x", "resume", "operator"},
		{"source-unavailable-decide", "s", "1", "delete", "operator"}, {"source-unavailable-decide", "s", "1", "resume", " "},
		{"source-unavailable-decide", "s", "1", "resume"},
	} {
		if _, ok, err := parseControlCommand(args); !ok || !errors.Is(err, errControlArguments) {
			t.Fatal("invalid accepted", args, ok, err)
		}
	}
	// Payload problems are rejected before any database access.
	for _, input := range []io.Reader{nil, strings.NewReader("not json"), strings.NewReader(`{"url":"https://x.example","unexpected":true}`)} {
		c, _, _ := parseControlCommand([]string{"source-unavailable-decide", "s", "1", "exclude_stale_unavailable", "operator"})
		c.Input = input
		if err := executeControl(context.Background(), nil, "admin", c, io.Discard); !errors.Is(err, errControlArguments) {
			t.Fatal("invalid payload accepted", err)
		}
	}
	if code := controlErrorCode(fmt.Errorf("wrapped: %w", acquisition.ErrTargetChanged)); code != "target_changed" {
		t.Fatal("disposition code", code)
	}
	if code := controlErrorCode(acquisition.ErrMeasureAcknowledgementRequired); code != "measure_acknowledgement_required" {
		t.Fatal("disposition code", code)
	}
}
