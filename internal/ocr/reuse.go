package ocr

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/Balestrino/italian-weather-alert/internal/inference"
)

func digest(body []byte) string { hash := sha256.Sum256(body); return hex.EncodeToString(hash[:]) }

type renderedIdentity struct {
	Number    int
	MediaType string
	Identity  ArtifactIdentity
}

func (r *Runner) manifestKey(resourceHash, mediaType string) string {
	encoded, _ := json.Marshal([]string{"ocr-render-v1", resourceHash, mediaType, r.ProviderScope, r.Model, r.ConfigurationVersion, r.RendererIdentity})
	return digest(encoded)
}

func (r *Runner) identities(pages []PageImage) []renderedIdentity {
	result := make([]renderedIdentity, 0, len(pages))
	for _, page := range pages {
		request, _ := json.Marshal(pageRequest(r.Model, page))
		result = append(result, renderedIdentity{page.Number, page.MediaType, ArtifactIdentity{Scope: r.ProviderScope, Model: r.Model, Configuration: r.ConfigurationVersion, Renderer: r.RendererIdentity, InputSHA256: digest(page.Bytes), RequestSHA256: digest(request)}})
	}
	return result
}

// warmManifest skips local rendering only when every immutable artifact exists.
// Incomplete manifests fall back to rendering; already finished pages still reuse.
func (s *Store) warmManifest(ctx context.Context, key string) ([]renderedIdentity, bool, error) {
	var body []byte
	err := s.pool.QueryRow(ctx, `SELECT pages FROM ocr_render_manifests WHERE manifest_key=$1`, key).Scan(&body)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}
	var pages []renderedIdentity
	if json.Unmarshal(body, &pages) != nil || len(pages) == 0 {
		return nil, false, ErrConflict
	}
	for index, page := range pages {
		key, err := page.Identity.Key()
		if err != nil || page.Number != index+1 {
			return nil, false, ErrConflict
		}
		var ready bool
		if err = s.pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM ocr_artifacts WHERE artifact_key=$1 AND state='ready')`, key).Scan(&ready); err != nil {
			return nil, false, err
		}
		if !ready {
			return nil, false, nil
		}
	}
	return pages, true, nil
}

func (s *Store) putManifest(ctx context.Context, key string, pages []renderedIdentity, now time.Time) error {
	if len(pages) == 0 {
		return ErrInvalid
	}
	for index, page := range pages {
		if _, err := page.Identity.Key(); err != nil || page.Number != index+1 {
			return ErrInvalid
		}
	}
	body, _ := json.Marshal(pages)
	_, err := s.pool.Exec(ctx, `INSERT INTO ocr_render_manifests(manifest_key,pages,created_at) VALUES($1,$2,$3) ON CONFLICT DO NOTHING`, key, body, now)
	if err != nil {
		return err
	}
	var same bool
	err = s.pool.QueryRow(ctx, `SELECT pages=$2::jsonb FROM ocr_render_manifests WHERE manifest_key=$1`, key, body).Scan(&same)
	if err == nil && !same {
		return ErrConflict
	}
	return err
}

// callOwned keeps ownership through response validation and artifact publication.
// A lost heartbeat cancels the provider context; the fenced completion prevents
// a stale worker from publishing even if the transport ignores cancellation.
func (r *Runner) callOwned(ctx context.Context, a Artifact, request inference.Request, save func(inference.Response) error) (inference.Response, error) {
	now := time.Now
	if r.Now != nil {
		now = r.Now
	}
	owned, cancel := context.WithCancel(ctx)
	done := make(chan error, 1)
	go func() {
		ticker := time.NewTicker(20 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-owned.Done():
				done <- nil
				return
			case <-ticker.C:
				beat, stop := context.WithTimeout(owned, 5*time.Second)
				err := r.Artifacts.RenewArtifact(beat, a, now(), time.Minute)
				stop()
				if err != nil {
					cancel()
					done <- err
					return
				}
			}
		}
	}()
	response, err := r.Adapter.Complete(owned, request)
	if err == nil && owned.Err() != nil {
		err = owned.Err()
	}
	if err == nil && owned.Err() == nil {
		err = save(response)
	}
	cancel()
	heartbeatErr := <-done
	if heartbeatErr != nil {
		return response, heartbeatErr
	}
	if err != nil {
		cleanup, stop := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
		defer stop()
		_, _ = r.Artifacts.pool.Exec(cleanup, `UPDATE ocr_artifacts SET lease_expires_at=$4 WHERE artifact_key=$1 AND owner_token=$2 AND generation=$3 AND state='pending'`, a.Key, a.Token, a.Generation, now())
	}
	return response, err
}
