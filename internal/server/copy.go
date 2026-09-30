package server

import (
	"context"
	"fmt"
	"mime"
	"net/http"
	"strconv"

	"github.com/Balestrino/italian-weather-alert/internal/publiccopy"
	"github.com/Balestrino/italian-weather-alert/internal/publicquery"
)

func (s *publicService) copyHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if s.copies == nil {
			response, status := s.failure(s.now().UTC(), publiccopy.ErrDisabled, nil)
			reply(w, status, response)
			return
		}
		documentID, versionID := r.PathValue("document_id"), r.PathValue("version_id")
		content, err := s.copies.Read(r.Context(), documentID, versionID)
		if err != nil {
			response, status := s.failure(s.now().UTC(), err, nil)
			reply(w, status, response)
			return
		}
		mediaType := content.MediaType
		if _, _, err := mime.ParseMediaType(mediaType); err != nil {
			mediaType = "application/octet-stream"
		}
		w.Header().Set("Content-Type", mediaType)
		w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="document-%s-version-%s"`, documentID, versionID))
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Cache-Control", "no-store")
		w.Header().Set("ETag", `"sha256:`+content.SHA256+`"`)
		w.Header().Set("Content-Length", strconv.Itoa(len(content.Bytes)))
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(content.Bytes)
	})
}

func (s *publicService) applyCopyLinks(ctx context.Context, data any) (any, error) {
	if s.copies == nil {
		return data, nil
	}
	cache := map[string]*string{}
	decorate := func(document *publicquery.Document) error {
		document.CopyURL = nil
		key := document.ID + "/" + document.VersionID
		if link, ok := cache[key]; ok {
			document.CopyURL = link
			return nil
		}
		link, err := s.copies.Link(ctx, document.ID, document.VersionID)
		if err != nil {
			return err
		}
		cache[key] = link
		document.CopyURL = link
		return nil
	}
	switch value := data.(type) {
	case []publicquery.Document:
		for index := range value {
			if err := decorate(&value[index]); err != nil {
				return nil, err
			}
		}
		return value, nil
	case documentData:
		if err := decorate(&value.Document); err != nil {
			return nil, err
		}
		for index := range value.Versions {
			if err := decorate(&value.Versions[index]); err != nil {
				return nil, err
			}
		}
		return value, nil
	case publicquery.Situation:
		for index := range value.DocumentsRequiringAttention {
			if err := decorate(&value.DocumentsRequiringAttention[index]); err != nil {
				return nil, err
			}
		}
		return value, nil
	}
	return data, nil
}
