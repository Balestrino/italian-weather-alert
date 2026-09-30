// Package publiccopy mediates access to retained document bytes without ever
// exposing object-store locations or credentials.
package publiccopy

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/Balestrino/italian-weather-alert/internal/documents"
)

var (
	ErrDisabled   = errors.New("public retained-copy access is disabled")
	ErrRestricted = errors.New("source policy restricts public retained-copy access")
	ErrUnknown    = errors.New("retained document version is unknown")
	ErrStorage    = errors.New("retained copy storage is unavailable")
)

type Content struct {
	Bytes     []byte
	MediaType string
	SHA256    string
}

type Service struct {
	pool    *pgxpool.Pool
	docs    *documents.Store
	enabled bool
	baseURL string
}

func New(pool *pgxpool.Pool, docs *documents.Store, enabled bool, baseURL string) (*Service, error) {
	baseURL = strings.TrimRight(baseURL, "/")
	if pool == nil {
		return nil, errors.New("public copy database is required")
	}
	if enabled {
		parsed, err := url.Parse(baseURL)
		if docs == nil || err != nil || parsed.Scheme == "" || parsed.Host == "" || parsed.User != nil || parsed.RawQuery != "" || parsed.Fragment != "" || parsed.Path != "" || (parsed.Scheme != "http" && parsed.Scheme != "https") {
			return nil, errors.New("invalid public copy configuration")
		}
	}
	return &Service{pool: pool, docs: docs, enabled: enabled, baseURL: baseURL}, nil
}

// Link evaluates current policy every time. Denial is represented as a null
// link so public document metadata remains available.
func (s *Service) Link(ctx context.Context, documentID, versionID string) (*string, error) {
	if !s.enabled {
		return nil, nil
	}
	_, err := s.authorize(ctx, documentID, versionID)
	if errors.Is(err, ErrDisabled) || errors.Is(err, ErrRestricted) || errors.Is(err, ErrUnknown) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	link := fmt.Sprintf("%s/v1/documents/%s/versions/%s/content", s.baseURL, url.PathEscape(documentID), url.PathEscape(versionID))
	return &link, nil
}

func (s *Service) Read(ctx context.Context, documentID, versionID string) (Content, error) {
	resource, err := s.authorize(ctx, documentID, versionID)
	if err != nil {
		return Content{}, err
	}
	bytes, err := s.docs.Read(ctx, resource.versionID, resource.officialURL)
	if err != nil {
		return Content{}, fmt.Errorf("read retained public copy: %w", ErrStorage)
	}
	return Content{Bytes: bytes, MediaType: resource.mediaType, SHA256: resource.sha256}, nil
}

type authorizedResource struct {
	versionID   int64
	officialURL string
	mediaType   string
	sha256      string
}

func (s *Service) authorize(ctx context.Context, documentID, versionID string) (authorizedResource, error) {
	document, err := strconv.ParseInt(documentID, 10, 64)
	if err != nil || document < 1 {
		return authorizedResource{}, ErrUnknown
	}
	version, err := strconv.ParseInt(versionID, 10, 64)
	if err != nil || version < 1 {
		return authorizedResource{}, ErrUnknown
	}
	var resource authorizedResource
	var publicEnabled, copiesPermitted bool
	err = s.pool.QueryRow(ctx, `SELECT v.id,r.url,r.media_type,r.object_hash,(s.public_enabled AND EXISTS (SELECT 1 FROM retained_acquisitions pa JOIN registry_events pe ON pe.source_id=pa.source_id AND pe.revision=pa.configuration AND pe.kind='acceptance' WHERE pa.document_id=v.document_id AND pa.version_id=v.id)),
 COALESCE((c.body->'policy'->>'copies_permitted')::boolean,false)
 FROM retained_versions v JOIN retained_documents d ON d.id=v.document_id
 JOIN registry_public_sources s ON s.id=d.source_id
 LEFT JOIN registry_configurations c ON c.source_id=s.id AND c.revision=s.active_revision
 JOIN retained_resources r ON r.version_id=v.id AND r.role='original' AND r.object_hash IS NOT NULL
 WHERE d.id=$1 AND v.id=$2`, document, version).Scan(&resource.versionID, &resource.officialURL, &resource.mediaType, &resource.sha256, &publicEnabled, &copiesPermitted)
	if errors.Is(err, pgx.ErrNoRows) {
		return authorizedResource{}, ErrUnknown
	}
	if err != nil {
		return authorizedResource{}, err
	}
	if !s.enabled {
		return authorizedResource{}, ErrDisabled
	}
	if !publicEnabled || !copiesPermitted {
		return authorizedResource{}, ErrRestricted
	}
	return resource, nil
}
