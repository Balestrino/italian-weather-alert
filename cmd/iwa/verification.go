package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"os"
	"strconv"
	"time"

	"github.com/Balestrino/italian-weather-alert/internal/backend/config"
	"github.com/Balestrino/italian-weather-alert/internal/backend/documents"
	"github.com/Balestrino/italian-weather-alert/internal/backend/domain"
	"github.com/jackc/pgx/v5/pgxpool"
)

func readVerificationRequest(reader io.Reader) (domain.VerificationRequest, error) {
	var request domain.VerificationRequest
	body, err := io.ReadAll(io.LimitReader(reader, 1048577))
	if err != nil || len(body) > 1048576 {
		return request, domain.ErrInvalid
	}
	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.DisallowUnknownFields()
	if err = decoder.Decode(&request); err != nil {
		return request, domain.ErrInvalid
	}
	var extra any
	if err = decoder.Decode(&extra); err != io.EOF {
		return request, domain.ErrInvalid
	}
	return request, nil
}

func executeVerification(ctx context.Context, pool *pgxpool.Pool, c config.Config, args []string, output io.Writer) bool {
	if c.Role != "admin" {
		slog.Error("multi-source verification requires admin role")
		return false
	}
	ctx, cancel := context.WithTimeout(ctx, 40*time.Second)
	defer cancel()
	store := domain.New(pool)
	var value any
	var err error
	switch args[0] {
	case "verify-candidate":
		file, openErr := os.Open(args[1])
		if openErr != nil {
			slog.Error("verification input unavailable")
			return false
		}
		defer file.Close()
		request, readErr := readVerificationRequest(file)
		if readErr != nil {
			slog.Error("verification input invalid")
			return false
		}
		sc, loadErr := config.LoadStorage()
		if loadErr != nil {
			slog.Error("storage configuration invalid")
			return false
		}
		objects, storageErr := documents.NewS3(c.RustFSURL, sc.Bucket, sc.AccessKey, sc.SecretKey)
		if storageErr != nil {
			slog.Error("storage initialization failed")
			return false
		}
		value, err = store.VerifyMultiSource(ctx, documents.New(pool, objects), request, time.Now())
	case "verification-receipt":
		id, parseErr := strconv.ParseInt(args[1], 10, 64)
		if parseErr != nil {
			err = domain.ErrInvalid
		} else {
			value, err = store.VerificationReceipt(ctx, id)
		}
	case "verification-history":
		after, parseErr := strconv.ParseInt(args[3], 10, 64)
		if parseErr != nil {
			err = domain.ErrInvalid
		} else {
			value, err = store.VerificationHistory(ctx, args[1], args[2], after, 100)
		}
	default:
		err = domain.ErrInvalid
	}
	if err != nil {
		code := "verification_failed"
		if errors.Is(err, domain.ErrInvalid) || errors.Is(err, domain.ErrVerificationEvidence) {
			code = "invalid_evidence"
		} else if errors.Is(err, domain.ErrConflict) {
			code = "verification_request_conflict"
		}
		slog.Error("multi-source verification failed", "code", code)
		return false
	}
	return json.NewEncoder(output).Encode(value) == nil
}
