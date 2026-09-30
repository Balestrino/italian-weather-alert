package publicquery

import (
	"context"
	"encoding/json"
	"errors"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5"
)

type documentMetadata struct {
	Kind         string           `json:"kind"`
	Publication  metadataTemporal `json:"publication"`
	Modification metadataTemporal `json:"modification"`
}

type metadataTemporal struct {
	Original   string  `json:"original"`
	Precision  string  `json:"precision"`
	Instant    *string `json:"instant"`
	Date       *string `json:"date"`
	EndInstant *string `json:"end_instant"`
	Timezone   *string `json:"timezone"`
	Assumption *string `json:"assumption"`
}

func (s *Store) Document(ctx context.Context, query DocumentQuery) (DocumentResult, error) {
	qt, err := normalizeTime(query.QueryTime)
	if err != nil || query.DocumentID == "" {
		return DocumentResult{}, ErrInvalidParameters
	}
	documentID, err := strconv.ParseInt(query.DocumentID, 10, 64)
	if err != nil || documentID < 1 {
		return DocumentResult{}, ErrUnknownIdentifier
	}
	var versionID int64
	if query.VersionID != "" {
		versionID, err = strconv.ParseInt(query.VersionID, 10, 64)
		if err != nil || versionID < 1 {
			return DocumentResult{}, ErrUnknownIdentifier
		}
		err = s.pool.QueryRow(ctx, `SELECT id FROM retained_versions WHERE id=$1 AND document_id=$2 AND first_acquired_at<=$3`, versionID, documentID, qt.KnownAt).Scan(&versionID)
	} else {
		err = s.pool.QueryRow(ctx, `SELECT v.id FROM retained_versions v JOIN retained_documents d ON d.id=v.document_id JOIN `+s.sourcesSQL()+` s ON s.id=d.source_id
 WHERE v.document_id=$1 AND v.first_acquired_at<=$2 AND `+s.visibilitySQL()+` ORDER BY v.first_acquired_at DESC,v.id DESC LIMIT 1`, documentID, qt.KnownAt).Scan(&versionID)
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return DocumentResult{}, ErrUnknownIdentifier
	}
	if err != nil {
		return DocumentResult{}, err
	}
	result := DocumentResult{}
	result.Document, err = s.documentVersion(ctx, versionID, qt)
	if err != nil {
		return DocumentResult{}, err
	}
	result.Versions = []Document{}
	if query.IncludeVersions {
		rows, rowsErr := s.pool.Query(ctx, `SELECT v.id FROM retained_versions v JOIN retained_documents d ON d.id=v.document_id JOIN `+s.sourcesSQL()+` s ON s.id=d.source_id
 WHERE v.document_id=$1 AND v.first_acquired_at<=$2 AND `+s.visibilitySQL()+` ORDER BY v.first_acquired_at,v.id`, documentID, qt.KnownAt)
		if rowsErr != nil {
			return DocumentResult{}, rowsErr
		}
		defer rows.Close()
		for rows.Next() {
			var id int64
			if err = rows.Scan(&id); err != nil {
				return DocumentResult{}, err
			}
			value, loadErr := s.documentVersion(ctx, id, qt)
			if loadErr != nil {
				return DocumentResult{}, loadErr
			}
			result.Versions = append(result.Versions, value)
		}
		if err = rows.Err(); err != nil {
			return DocumentResult{}, err
		}
	}
	result.History, err = s.history(ctx, qt.KnownAt, nil, historyScope{DocumentID: documentID})
	return result, err
}

func (s *Store) documentVersion(ctx context.Context, versionID int64, qt QueryTime) (Document, error) {
	var value Document
	var numericDocumentID int64
	var issuer *string
	var metadata []byte
	err := s.pool.QueryRow(ctx, `SELECT d.id,v.id,d.source_id,issuer.name,publisher.name,d.official_url,v.content_hash,v.first_acquired_at,v.metadata
 FROM retained_versions v JOIN retained_documents d ON d.id=v.document_id
 JOIN `+s.sourcesSQL()+` s ON s.id=d.source_id JOIN registry_channels c ON c.id=s.channel_id
 JOIN registry_authorities publisher ON publisher.id=c.publisher_id
 LEFT JOIN registry_authorities issuer ON issuer.id=v.issuer_id
 WHERE v.id=$1 AND v.first_acquired_at<=$2 AND `+s.visibilitySQL()+``, versionID, qt.KnownAt).Scan(&numericDocumentID, &versionID, &value.SourceID, &issuer, &value.Publisher, &value.OfficialURL, &value.SHA256, &value.AcquiredAt, &metadata)
	if errors.Is(err, pgx.ErrNoRows) {
		return Document{}, ErrUnknownIdentifier
	}
	if err != nil {
		return Document{}, err
	}
	value.ID, value.VersionID, value.Issuer = stringID(numericDocumentID), stringID(versionID), issuer
	value.AcquiredAt = value.AcquiredAt.UTC()
	value.CopyURL = nil // Task 6.5 owns application-mediated copy authorization.
	value.Publication, value.Modification, value.Kind = metadataFields(metadata)
	interpretation, interpretedAt, runID, err := s.documentInterpretation(ctx, versionID, qt.KnownAt)
	if err != nil {
		return Document{}, err
	}
	value.InterpretedAt, value.InterpretationRunID = interpretedAt, runID
	value.Quality, err = s.quality(ctx, value.SourceID, qt, interpretation)
	if err != nil {
		return Document{}, err
	}
	value.Attachments, err = s.attachments(ctx, versionID)
	return value, err
}

func metadataFields(raw []byte) (Temporal, Temporal, string) {
	publication, modification := unknownTemporal(), unknownTemporal()
	kind := "document"
	var metadata documentMetadata
	if json.Unmarshal(raw, &metadata) != nil {
		return publication, modification, kind
	}
	if metadata.Kind != "" {
		kind = metadata.Kind
	}
	return temporalFromMetadata(metadata.Publication), temporalFromMetadata(metadata.Modification), kind
}

func temporalFromMetadata(value metadataTemporal) Temporal {
	if value.Precision == "" {
		return unknownTemporal()
	}
	result := Temporal{Original: value.Original, Precision: value.Precision, Date: value.Date, Timezone: value.Timezone, Assumption: value.Assumption}
	if value.Instant != nil {
		if parsed, err := time.Parse(time.RFC3339, *value.Instant); err == nil {
			parsed = parsed.UTC()
			result.Instant = &parsed
		}
	}
	if value.EndInstant != nil {
		if parsed, err := time.Parse(time.RFC3339, *value.EndInstant); err == nil {
			parsed = parsed.UTC()
			result.EndInstant = &parsed
		}
	}
	return result
}

func (s *Store) documentInterpretation(ctx context.Context, versionID int64, knownAt time.Time) (Dimension, *time.Time, *string, error) {
	var status, reason string
	var createdAt time.Time
	var runID int64
	err := s.pool.QueryRow(ctx, `SELECT e.status,e.reason_code,e.created_at,e.run_id FROM extraction_results e
 WHERE e.document_version_id=$1 AND e.created_at<=registry_interpretation_cutoff((SELECT d.source_id FROM retained_versions v JOIN retained_documents d ON d.id=v.document_id WHERE v.id=$1),$2) ORDER BY e.created_at DESC,e.run_id DESC LIMIT 1`, versionID, knownAt).Scan(&status, &reason, &createdAt, &runID)
	if errors.Is(err, pgx.ErrNoRows) {
		return Dimension{State: "not_processed", Limitations: []string{"no interpretation was available at the knowledge boundary"}, Evidence: []Evidence{}}, nil, nil, nil
	}
	if err != nil {
		return Dimension{}, nil, nil, err
	}
	state := "failed"
	if status == "extracted" || status == "not_applicable" {
		state = "supported"
	}
	createdAt = createdAt.UTC()
	id := stringID(runID)
	assessment, err := s.preserveDefectWarning(ctx, versionID, createdAt, Dimension{State: state, Limitations: []string{reason}, Evidence: []Evidence{}})
	return assessment, &createdAt, &id, err
}

func (s *Store) attachments(ctx context.Context, versionID int64) ([]Attachment, error) {
	rows, err := s.pool.Query(ctx, `SELECT url,object_hash,missing FROM retained_resources WHERE version_id=$1 AND role='attachment' ORDER BY url`, versionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := []Attachment{}
	for rows.Next() {
		var value Attachment
		var hash *string
		var missing string
		if err = rows.Scan(&value.OfficialURL, &hash, &missing); err != nil {
			return nil, err
		}
		switch {
		case hash != nil:
			value.Status = "acquired"
		case missing == "forbidden":
			value.Status = "restricted"
		default:
			value.Status = "unavailable"
		}
		result = append(result, value)
	}
	return result, rows.Err()
}
