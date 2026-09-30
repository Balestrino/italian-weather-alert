package documents

import (
	"bytes"
	"encoding/json"
	"errors"
	"testing"
)

func TestManifest(t *testing.T) {
	a := Acquisition{ID: "check", SourceID: "source", Configuration: 1, URL: "https://example.org/a", Metadata: json.RawMessage(`{"date":"2026-01-01","n":9007199254740993}`), Resources: []Resource{{URL: "https://example.org/a", Role: "original", Required: true, SourceID: "source", Configuration: 1, MediaType: "text/html", Bytes: []byte{}}, {URL: "https://example.org/b", Role: "attachment", Required: true, SourceID: "source", Configuration: 1, Missing: "unavailable"}}}
	prepared, _, hash, complete, err := prepare(a)
	if err != nil || complete {
		t.Fatalf("missing resource state: %v", err)
	}
	raw, err := json.Marshal(prepared)
	if err != nil {
		t.Fatal(err)
	}
	var replay Acquisition
	if err = json.Unmarshal(raw, &replay); err != nil {
		t.Fatal(err)
	}
	_, _, got, _, err := prepare(replay)
	if err != nil || got != hash {
		t.Fatal("empty originals or metadata precision lost in journal")
	}
	if !bytes.Contains(replay.Metadata, []byte("9007199254740993")) {
		t.Fatal("metadata number rounded")
	}
	a.Configuration = 2
	for i := range a.Resources {
		a.Resources[i].Configuration = 2
	}
	a.Resources[0], a.Resources[1] = a.Resources[1], a.Resources[0]
	_, _, got, _, err = prepare(a)
	if err != nil || got != hash {
		t.Fatal("resource order or policy revision duplicated content")
	}
	a.Resources[0].Bytes = []byte("PDF")
	a.Resources[0].Missing = ""
	a.Resources[0].MediaType = "application/pdf"
	_, _, got, complete, err = prepare(a)
	if err != nil || !complete || got == hash {
		t.Fatal("attachment recovery not represented")
	}
	a.Resources = append(a.Resources, a.Resources[0])
	if _, _, _, _, err = prepare(a); !errors.Is(err, ErrInvalid) {
		t.Fatal("duplicate resource accepted")
	}
}

func TestInvalidManifest(t *testing.T) {
	base := func() Acquisition {
		return Acquisition{ID: "id", SourceID: "source", Configuration: 1, URL: "https://example.org/a", Resources: []Resource{{URL: "https://example.org/a", Role: "original", Required: true, SourceID: "source", Configuration: 1, MediaType: "text/html", Bytes: []byte("original")}}}
	}
	for _, mutate := range []func(*Acquisition){
		func(a *Acquisition) { a.Resources = nil },
		func(a *Acquisition) { a.Resources[0].Bytes = nil },
		func(a *Acquisition) { a.Resources[0].Required = false },
		func(a *Acquisition) { a.Resources[0].Missing = "unavailable" },
		func(a *Acquisition) { a.Resources[0].Configuration = 2 },
		func(a *Acquisition) { a.Metadata = json.RawMessage(`[]`) },
		func(a *Acquisition) { a.URL = "https://user:secret@example.org/a" },
	} {
		a := base()
		mutate(&a)
		if _, _, _, _, err := prepare(a); !errors.Is(err, ErrInvalid) {
			t.Fatal("invalid input accepted")
		}
	}
}
