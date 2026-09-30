// A bounded, read-only reconstruction of the two September 23 failed requests.
// Provider responses and retained input are written only to a new private file.
package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/Balestrino/italian-weather-alert/internal/classification"
	"github.com/Balestrino/italian-weather-alert/internal/config"
	"github.com/Balestrino/italian-weather-alert/internal/documents"
	"github.com/Balestrino/italian-weather-alert/internal/inference"
	"github.com/Balestrino/italian-weather-alert/internal/ocr"
)

type review struct {
	Version         int64
	OriginalRun     int64
	OriginalCall    int64
	RequestHash     string
	Segment         classification.Segment
	Request         inference.Request
	Response        string
	ReturnedModel   string
	FinishReason    string
	Usage           inference.Usage
	LegacyAccepted  bool
	CurrentAccepted bool
	Reason          string
	ErrorCode       string
	SentAt          *time.Time
}
type report struct {
	CreatedAt time.Time
	Live      bool
	MaxCalls  int
	Calls     int
	Reviews   []review
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "classification review failed; inspect private evidence if created")
		os.Exit(1)
	}
}
func hashRequest(r inference.Request) string {
	b, _ := json.Marshal(r)
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:])
}
func matchedRequest(model string, s classification.Segment, want string) (inference.Request, error) {
	r, e := classification.SegmentRequest(model, s)
	if e != nil {
		return r, e
	}
	r.ChatTemplateKwargs = &inference.ChatTemplateOptions{EnableThinking: false}
	if hashRequest(r) != want {
		return r, errors.New("historical request mismatch")
	}
	return r, nil
}
func persist(f *os.File, r report) error {
	b, e := json.MarshalIndent(r, "", "  ")
	if e != nil {
		return e
	}
	if _, e = f.Seek(0, 0); e != nil {
		return e
	}
	if e = f.Truncate(0); e != nil {
		return e
	}
	if _, e = f.Write(append(b, '\n')); e != nil {
		return e
	}
	return f.Sync()
}
func run() error {
	out := flag.String("out", "", "new private evidence file")
	live := flag.Bool("live", false, "send only the two hash-matched failed requests; no OCR or retry")
	flag.Parse()
	if *out == "" {
		return errors.New("output required")
	}
	f, e := os.OpenFile(*out, os.O_RDWR|os.O_CREATE|os.O_EXCL, 0600)
	if e != nil {
		return e
	}
	defer f.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	c, e := config.Load()
	if e != nil {
		return e
	}
	if c.Role != "worker" && c.Role != "admin" {
		return errors.New("maintenance role required")
	}
	sc, e := config.LoadStorage()
	if e != nil {
		return e
	}
	pc, e := pgxpool.ParseConfig("postgres://iwa@localhost/iwa?sslmode=disable")
	if e != nil {
		return e
	}
	host, port, e := net.SplitHostPort(c.PostgresHost)
	if e != nil {
		return e
	}
	n, e := strconv.Atoi(port)
	if e != nil {
		return e
	}
	pc.ConnConfig.Host = host
	pc.ConnConfig.Port = uint16(n)
	pc.ConnConfig.Password = c.PostgresPassword
	pc.ConnConfig.RuntimeParams["default_transaction_read_only"] = "on"
	pool, e := pgxpool.NewWithConfig(ctx, pc)
	if e != nil {
		return e
	}
	defer pool.Close()
	objects, e := documents.NewS3(c.RustFSURL, sc.Bucket, sc.AccessKey, sc.SecretKey)
	if e != nil {
		return e
	}
	docs := documents.New(pool, objects)
	scans := ocr.NewStore(pool)
	result := report{CreatedAt: time.Now().UTC(), Live: *live, MaxCalls: 2}
	for _, id := range []int64{3419, 3476} {
		v, e := docs.Version(ctx, id)
		if e != nil {
			return e
		}
		content, e := classification.GatherContent(ctx, docs, scans, v)
		if e != nil || !content.Complete {
			return errors.New("incomplete input")
		}
		segments, e := classification.SegmentContent(id, content)
		if e != nil {
			return e
		}
		x := review{Version: id}
		var ordinal int
		var model string
		e = pool.QueryRow(ctx, `SELECT c.id,c.run_id,c.ordinal,c.input_sha256,c.requested_model FROM processing_provider_calls c JOIN processing_runs r ON r.id=c.run_id JOIN processing_run_attempts a ON a.run_id=c.run_id AND a.number=c.attempt_number WHERE r.document_version_id=$1 AND r.stage='classification' AND a.error_code='classification_output_quotation' ORDER BY c.id DESC LIMIT 1`, id).Scan(&x.OriginalCall, &x.OriginalRun, &ordinal, &x.RequestHash, &model)
		if e != nil {
			return e
		}
		if ordinal < 1 || ordinal > len(segments) {
			return errors.New("segment unavailable")
		}
		x.Segment = segments[ordinal-1]
		x.Request, e = matchedRequest(model, x.Segment, x.RequestHash)
		if e != nil {
			return e
		}
		result.Reviews = append(result.Reviews, x)
	}
	if e = persist(f, result); e != nil {
		return e
	}
	if !*live {
		return nil
	}
	q, e := config.LoadQwen()
	if e != nil {
		return e
	}
	a, e := inference.NewOpenAIChat(q.Adapter, q.URL, q.APIKey, &http.Client{Timeout: 90 * time.Second})
	if e != nil {
		return e
	}
	for i := range result.Reviews {
		x := &result.Reviews[i]
		if x.Request.Model != q.Model || result.Calls >= result.MaxCalls {
			return errors.New("call boundary")
		}
		now := time.Now().UTC()
		x.SentAt = &now
		result.Calls++
		if e = persist(f, result); e != nil {
			return e
		}
		response, callErr := a.Complete(ctx, x.Request)
		x.Response = response.Content
		x.ReturnedModel = response.Model
		x.FinishReason = response.FinishReason
		x.Usage = response.Usage
		if callErr != nil {
			x.ErrorCode = "provider_error"
		} else {
			_, legacyErr := classification.ParseLegacyDecision(x.Response, x.Segment.Content())
			x.LegacyAccepted = legacyErr == nil
			_, currentErr := classification.ParseDecision(x.Response, x.Segment.Content())
			x.CurrentAccepted = currentErr == nil
			if legacyErr != nil {
				x.Reason = classification.InvalidOutputReason(x.Response, x.FinishReason, x.Segment.Content())
			}
		}
		if e = persist(f, result); e != nil {
			return e
		}
		if callErr != nil {
			return errors.New("provider stopped")
		}
	}
	return nil
}
