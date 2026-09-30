package health

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestProbeAuthenticationAndRedirects(t *testing.T) {
	leaked := false
	target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { leaked = true }))
	defer target.Close()
	source := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer private-fixture-value" {
			t.Error("missing auth")
		}
		if r.URL.Path == "/redirect" {
			http.Redirect(w, r, target.URL, 302)
			return
		}
		w.WriteHeader(200)
	}))
	defer source.Close()
	if err := HTTP(source.URL, "private-fixture-value")(context.Background()); err != nil {
		t.Fatal(err)
	}
	if err := HTTP(source.URL+"/redirect", "private-fixture-value")(context.Background()); err == nil || leaked {
		t.Fatal("redirect followed")
	}
}
func TestProbeRespectsDeadline(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { <-r.Context().Done() }))
	defer srv.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()
	if err := HTTP(srv.URL, "")(ctx); err == nil {
		t.Fatal("deadline ignored")
	}
}
