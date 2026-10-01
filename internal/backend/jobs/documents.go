package jobs

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Balestrino/italian-weather-alert/internal/backend/documents"
)

const (
	DocumentQueue = "documents"
	DocumentKind  = "retain_acquisition"
)

type documentPayload struct {
	AcquisitionID string `json:"acquisition_id"`
}

// EnqueuePendingDocuments reconciles the durable acquisition journal with the
// queue. It closes the crash window between staging an acquisition and enqueue.
func EnqueuePendingDocuments(ctx context.Context, queue *Store, retained *documents.Store, now time.Time) (int, error) {
	ids, err := retained.Pending(ctx, 1000)
	if err != nil {
		return 0, err
	}
	for _, id := range ids {
		payload, _ := json.Marshal(documentPayload{AcquisitionID: id})
		if _, err = queue.Enqueue(ctx, EnqueueRequest{Queue: DocumentQueue, Kind: DocumentKind, IdempotencyKey: id, Payload: payload, MaxAttempts: 3, RetryBase: time.Second, AvailableAt: now.UTC()}); err != nil {
			return 0, err
		}
	}
	return len(ids), nil
}

func DocumentHandler(retained *documents.Store) Handler {
	return func(ctx context.Context, job Job) (Result, error) {
		var payload documentPayload
		if json.Unmarshal(job.Payload, &payload) != nil || payload.AcquisitionID == "" {
			return Result{}, &HandlerError{Failure{Code: "invalid_acquisition_job", Detail: "acquisition identifier is missing", Temporary: false}}
		}
		version, err := retained.Recover(ctx, payload.AcquisitionID)
		if err != nil {
			return Result{}, &HandlerError{Failure{Code: "retention_failed", Detail: "retained acquisition could not be finalized", Temporary: true}}
		}
		effectPayload, _ := json.Marshal(struct {
			VersionID   int64  `json:"version_id"`
			ContentHash string `json:"content_hash"`
		}{version.ID, version.Hash})
		return JSONResult(struct {
			VersionID int64 `json:"version_id"`
		}{version.ID}, Effect{Key: fmt.Sprintf("document-version:%d", version.ID), Kind: "document_version_ready", Payload: effectPayload})
	}
}
