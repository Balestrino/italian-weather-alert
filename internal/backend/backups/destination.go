package backups

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"io"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"
)

type Completion struct {
	Format      int       `json:"format"`
	ID          string    `json:"id"`
	SHA256      string    `json:"sha256"`
	Size        int64     `json:"bytes"`
	CompletedAt time.Time `json:"completed_at"`
}
type Destination interface {
	Publish(context.Context, Bundle, time.Time) error
	Prune(context.Context, time.Time, string) error
}
type S3Destination struct {
	client         *minio.Client
	bucket, prefix string
}

var runPattern = regexp.MustCompile(`^[0-9a-f]{32}$`)
var digestPattern = regexp.MustCompile(`^[0-9a-f]{64}$`)

func NewS3Destination(c Config) (*S3Destination, error) {
	if !c.Enabled || c.Validate() != nil {
		return nil, ErrConfig
	}
	u, _ := url.Parse(c.Endpoint)
	client, err := minio.New(u.Host, &minio.Options{Creds: credentials.NewStaticV4(c.AccessKey, c.SecretKey, ""), Secure: u.Scheme == "https", Region: c.Region, BucketLookup: minio.BucketLookupPath})
	if err != nil {
		return nil, ErrConfig
	}
	return &S3Destination{client, c.Bucket, c.Prefix + "/"}, nil
}
func (s *S3Destination) key(id, suffix string) string { return s.prefix + id + suffix }
func (s *S3Destination) verify(ctx context.Context, key, hash string, size int64) error {
	obj, err := s.client.GetObject(ctx, s.bucket, key, minio.GetObjectOptions{})
	if err != nil {
		return ErrBackup
	}
	defer obj.Close()
	h := sha256.New()
	n, err := io.Copy(h, io.LimitReader(obj, size+1))
	if err != nil || n != size || hex.EncodeToString(h.Sum(nil)) != hash {
		return ErrBackup
	}
	return nil
}
func (s *S3Destination) Publish(ctx context.Context, b Bundle, at time.Time) error {
	if !runPattern.MatchString(b.Manifest.ID) || !digestPattern.MatchString(b.SHA256) || b.Size < 1 {
		return ErrBackup
	}
	f, err := os.Open(b.Path)
	if err != nil {
		return ErrBackup
	}
	defer f.Close()
	key := s.key(b.Manifest.ID, ".tar")
	// No bucket creation or policy changes: destination access is provisioned by
	// the operator. Completion is published only after full download verification.
	if _, err = s.client.PutObject(ctx, s.bucket, key, f, b.Size, minio.PutObjectOptions{ContentType: "application/x-tar"}); err != nil {
		return ErrBackup
	}
	if s.verify(ctx, key, b.SHA256, b.Size) != nil {
		return ErrBackup
	}
	raw, err := json.Marshal(Completion{1, b.Manifest.ID, b.SHA256, b.Size, at.UTC()})
	if err != nil {
		return ErrBackup
	}
	key = s.key(b.Manifest.ID, ".complete.json")
	if _, err = s.client.PutObject(ctx, s.bucket, key, bytes.NewReader(raw), int64(len(raw)), minio.PutObjectOptions{ContentType: "application/json"}); err != nil {
		return ErrBackup
	}
	hash := sha256.Sum256(raw)
	return s.verify(ctx, key, hex.EncodeToString(hash[:]), int64(len(raw)))
}
func (s *S3Destination) Prune(ctx context.Context, cutoff time.Time, keep string) error {
	// List fully before mutation, avoiding iterator errors after partial pruning.
	var expired []string
	archives := map[string]time.Time{}
	complete := map[string]bool{}
	for obj := range s.client.ListObjects(ctx, s.bucket, minio.ListObjectsOptions{Prefix: s.prefix, Recursive: true}) {
		if obj.Err != nil {
			return ErrBackup
		}
		if strings.HasSuffix(obj.Key, ".tar") {
			id := strings.TrimSuffix(strings.TrimPrefix(obj.Key, s.prefix), ".tar")
			if runPattern.MatchString(id) {
				archives[id] = obj.LastModified
			}
			continue
		}
		if !strings.HasSuffix(obj.Key, ".complete.json") {
			continue
		}
		id := strings.TrimSuffix(strings.TrimPrefix(obj.Key, s.prefix), ".complete.json")
		complete[id] = true
		if id == keep || !runPattern.MatchString(id) {
			continue
		}
		r, err := s.client.GetObject(ctx, s.bucket, obj.Key, minio.GetObjectOptions{})
		if err != nil {
			return ErrBackup
		}
		raw, err := io.ReadAll(io.LimitReader(r, 4097))
		_ = r.Close()
		if err != nil || len(raw) > 4096 {
			return ErrBackup
		}
		var c Completion
		if json.Unmarshal(raw, &c) != nil || c.Format != 1 || c.ID != id || !digestPattern.MatchString(c.SHA256) || c.Size < 1 || c.CompletedAt.IsZero() {
			return ErrBackup
		}
		if c.CompletedAt.Before(cutoff) {
			expired = append(expired, id)
		}
	}
	for _, id := range expired {
		// A failed deletion must never leave a complete marker for a missing bundle.
		if s.client.RemoveObject(ctx, s.bucket, s.key(id, ".complete.json"), minio.RemoveObjectOptions{}) != nil {
			return ErrBackup
		}
		if s.client.RemoveObject(ctx, s.bucket, s.key(id, ".tar"), minio.RemoveObjectOptions{}) != nil {
			return ErrBackup
		}
	}
	// Incomplete transfers and interrupted deletions have no completion marker.
	// Bound their retention too, without touching the current run or other keys.
	for id, at := range archives {
		if id != keep && !complete[id] && at.Before(cutoff) {
			if s.client.RemoveObject(ctx, s.bucket, s.key(id, ".tar"), minio.RemoveObjectOptions{}) != nil {
				return ErrBackup
			}
		}
	}
	return nil
}
