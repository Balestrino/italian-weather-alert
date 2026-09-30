package inference

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"time"

	"github.com/Balestrino/italian-weather-alert/internal/processing"
)

type CheckpointStore interface {
	Checkpoint(context.Context, processing.SegmentCheckpoint) (processing.SegmentCheckpoint, bool, error)
	PutCheckpoint(context.Context, processing.SegmentCheckpoint, time.Time) error
	UseCheckpoint(context.Context, processing.SegmentCheckpoint, processing.Attempt, int, bool) error
}

func checkpointKey(stage, configuration string, request Request) processing.SegmentCheckpoint {
	body, _ := json.Marshal(request)
	hash := sha256.Sum256(body)
	return processing.SegmentCheckpoint{Stage: stage, Configuration: configuration, InputSHA256: hex.EncodeToString(hash[:])}
}
func LoadCheckpoint(ctx context.Context, store CheckpointStore, stage, configuration string, ordinal int, request Request) (Response, bool, error) {
	if store == nil {
		return Response{}, false, nil
	}
	a, ok := ctx.Value(attemptKey{}).(*callAttempt)
	if !ok {
		return Response{}, false, ErrInvalid
	}
	c, found, err := store.Checkpoint(ctx, checkpointKey(stage, configuration, request))
	if err != nil || !found {
		return Response{}, false, err
	}
	var response Response
	if json.Unmarshal(c.Response, &response) != nil {
		return Response{}, false, ErrInvalid
	}
	if err = store.UseCheckpoint(ctx, c, processing.Attempt{RunID: a.run, Number: a.number}, ordinal, true); err != nil {
		return Response{}, false, err
	}
	response.Usage = Usage{}
	return response, true, nil
}
func SaveCheckpoint(ctx context.Context, store CheckpointStore, stage, configuration string, ordinal int, request Request, response Response) error {
	if store == nil {
		return nil
	}
	a, ok := ctx.Value(attemptKey{}).(*callAttempt)
	if !ok {
		return ErrInvalid
	}
	c := checkpointKey(stage, configuration, request)
	c.CallID = response.CallID
	response.ReasoningContent = ""
	response.Diagnostic = Diagnostic{}
	c.Response, _ = json.Marshal(response)
	if err := store.PutCheckpoint(ctx, c, time.Now()); err != nil {
		return err
	}
	return store.UseCheckpoint(ctx, c, processing.Attempt{RunID: a.run, Number: a.number}, ordinal, false)
}
