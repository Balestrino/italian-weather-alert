package documents

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/url"
	"regexp"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

var errObjectMissing = errors.New("retained object missing")
var hashPattern = regexp.MustCompile(`^[0-9a-f]{64}$`)

type S3 struct {
	client *minio.Client
	bucket string
}

func NewS3(endpoint, bucket, accessKey, secretKey string) (*S3, error) {
	u, err := url.Parse(endpoint)
	if err != nil || u.Host == "" || (u.Path != "" && u.Path != "/") || u.User != nil || u.RawQuery != "" || u.Fragment != "" || (u.Scheme != "http" && u.Scheme != "https") || bucket == "" || accessKey == "" || secretKey == "" {
		return nil, ErrInvalid
	}
	c, err := minio.New(u.Host, &minio.Options{Creds: credentials.NewStaticV4(accessKey, secretKey, ""), Secure: u.Scheme == "https", Region: "us-east-1", BucketLookup: minio.BucketLookupPath})
	if err != nil {
		return nil, ErrStorage
	}
	return &S3{client: c, bucket: bucket}, nil
}

// Initialize is explicit; no request or normal service startup creates buckets
// or changes bucket policy. Newly created buckets have no public policy.
func (s *S3) Initialize(ctx context.Context) error {
	exists, err := s.client.BucketExists(ctx, s.bucket)
	if err != nil {
		return ErrStorage
	}
	if exists {
		return nil
	}
	if err = s.client.MakeBucket(ctx, s.bucket, minio.MakeBucketOptions{Region: "us-east-1"}); err != nil {
		if exists, e := s.client.BucketExists(ctx, s.bucket); e != nil || !exists {
			return ErrStorage
		}
	}
	return nil
}

func (s *S3) Read(ctx context.Context, hash string, size int64) ([]byte, error) {
	if !hashPattern.MatchString(hash) || size < 0 || size > MaxAcquisitionBytes {
		return nil, ErrInvalid
	}
	obj, err := s.client.GetObject(ctx, s.bucket, objectKey(hash), minio.GetObjectOptions{})
	if err != nil {
		return nil, ErrStorage
	}
	defer obj.Close()
	b, err := io.ReadAll(io.LimitReader(obj, size+1))
	if err != nil {
		if minio.ToErrorResponse(err).Code == "NoSuchKey" {
			return nil, errObjectMissing
		}
		return nil, ErrStorage
	}
	if int64(len(b)) != size || digest(b) != hash {
		return nil, ErrStorage
	}
	return b, nil
}

func (s *S3) Ensure(ctx context.Context, hash string, b []byte) error {
	if !hashPattern.MatchString(hash) || digest(b) != hash || len(b) > MaxAcquisitionBytes {
		return ErrInvalid
	}
	_, err := s.Read(ctx, hash, int64(len(b)))
	if err == nil {
		return nil
	}
	if !errors.Is(err, errObjectMissing) {
		return err
	}
	_, err = s.client.PutObject(ctx, s.bucket, objectKey(hash), bytes.NewReader(b), int64(len(b)), minio.PutObjectOptions{ContentType: "application/octet-stream", DisableMultipart: true, SendContentMd5: true})
	if err != nil {
		return ErrStorage
	}
	_, err = s.Read(ctx, hash, int64(len(b)))
	return err
}

// Delete is idempotent. PostgreSQL removes all references and durably queues
// the hash before this object-store effect is attempted.
func (s *S3) Delete(ctx context.Context, hash string) error {
	if !hashPattern.MatchString(hash) {
		return ErrInvalid
	}
	if err := s.client.RemoveObject(ctx, s.bucket, objectKey(hash), minio.RemoveObjectOptions{}); err != nil {
		return ErrStorage
	}
	return nil
}
