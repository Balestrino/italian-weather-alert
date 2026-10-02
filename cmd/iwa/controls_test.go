package main

import (
	"context"
	"errors"
	"github.com/Balestrino/italian-weather-alert/internal/backend/domain"
	"io"
	"strings"
	"testing"
)

func TestControlArgumentsAndRole(t *testing.T) {
	for _, args := range [][]string{
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
