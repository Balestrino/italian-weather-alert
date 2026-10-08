package publicquery

import (
	"context"
	"errors"
	"sort"
)

func (s *Store) MunicipalitySituation(ctx context.Context, query SituationQuery) (Situation, error) {
	s = s.forMunicipality(query.MunicipalityISTAT)
	qt, err := normalizeTime(query.QueryTime)
	if err != nil || len(query.MunicipalityISTAT) != 6 {
		return Situation{}, ErrInvalidParameters
	}
	s = s.withQueryReads(qt)
	result := Situation{}
	result.Municipality, err = s.municipality(ctx, query.MunicipalityISTAT, qt.KnownAt)
	if err != nil {
		return Situation{}, err
	}
	result.LocalMeasures, err = s.measures(ctx, query.MunicipalityISTAT, qt)
	if err != nil {
		return Situation{}, err
	}
	result.OperationalPhases, err = s.phases(ctx, query.MunicipalityISTAT, qt)
	if err != nil {
		return Situation{}, err
	}
	result.RegionalProducts, err = s.regional(ctx, query.MunicipalityISTAT, "", "", "", "", qt)
	if err != nil {
		return Situation{}, err
	}
	result.DocumentsRequiringAttention = []Document{}
	seen := map[string]bool{}
	for _, measure := range result.LocalMeasures {
		for _, documentID := range measure.NewerUninterpretedDocumentIDs {
			if seen[documentID] {
				continue
			}
			seen[documentID] = true
			document, documentErr := s.Document(ctx, DocumentQuery{QueryTime: qt, DocumentID: documentID})
			if documentErr != nil {
				if errors.Is(documentErr, ErrUnknownIdentifier) {
					continue
				}
				return Situation{}, documentErr
			}
			result.DocumentsRequiringAttention = append(result.DocumentsRequiringAttention, document.Document)
		}
	}
	// First/uninterpreted notices also matter when no previous measure exists.
	documents, documentErr := s.searchDocuments(ctx, SearchQuery{Kind: "attention", QueryTime: qt, MunicipalityISTAT: query.MunicipalityISTAT})
	if documentErr != nil {
		return Situation{}, documentErr
	}
	for _, document := range documents {
		if seen[document.ID] || document.Quality.Interpretation.State == "supported" {
			continue
		}
		seen[document.ID] = true
		result.DocumentsRequiringAttention = append(result.DocumentsRequiringAttention, document)
	}
	coverage, err := s.Coverage(ctx, CoverageQuery{QueryTime: qt, MunicipalityISTAT: query.MunicipalityISTAT})
	if err != nil {
		return Situation{}, err
	}
	result.Coverage = coverage.Sources
	result.History = coverage.History
	return result, nil
}

func (s *Store) Search(ctx context.Context, query SearchQuery) (SearchResult, error) {
	s = s.forMunicipality(query.MunicipalityISTAT)
	qt, err := normalizeTime(query.QueryTime)
	if err != nil || !validStatus(query.Status) || !validProduct(query.Product) || !validRisk(query.Risk) || (query.Kind != "document" && query.Kind != "measure" && query.Kind != "regional") || (query.MunicipalityISTAT != "" && len(query.MunicipalityISTAT) != 6) || (query.From != nil && query.To != nil && !query.From.Before(*query.To)) {
		return SearchResult{}, ErrInvalidParameters
	}
	if query.Kind != "regional" && (query.Product != "" || query.Risk != "" || query.Zone != "") {
		return SearchResult{}, ErrInvalidParameters
	}
	if query.Kind == "document" && query.Status != "" {
		return SearchResult{}, ErrInvalidParameters
	}
	s = s.withQueryReads(qt)
	query.QueryTime = qt
	result := SearchResult{Kind: query.Kind, Documents: []Document{}, Measures: []Measure{}, Regional: []RegionalWarning{}}
	switch query.Kind {
	case "document":
		result.Documents, err = s.searchDocuments(ctx, query)
	case "measure":
		result.Measures, err = s.measures(ctx, query.MunicipalityISTAT, qt)
		if err == nil {
			result.Measures = filterMeasure(result.Measures, query)
			if query.SourceID != "" {
				result.Measures, err = s.filterMeasuresBySource(ctx, result.Measures, query.SourceID)
			}
		}
	case "regional":
		result.Regional, err = s.regional(ctx, query.MunicipalityISTAT, query.Zone, query.Product, query.Risk, query.SourceID, qt, query.From != nil || query.To != nil || query.Status == "expired" || query.Status == "superseded")
		if err == nil {
			result.Regional = filterRegional(result.Regional, query)
		}
	}
	if err != nil {
		return SearchResult{}, err
	}
	result.History, err = s.history(ctx, qt.KnownAt, query.From, historyScope{Municipality: query.MunicipalityISTAT, SourceID: query.SourceID, Product: query.Product})
	return result, err
}

func (s *Store) filterMeasuresBySource(ctx context.Context, values []Measure, sourceID string) ([]Measure, error) {
	result := []Measure{}
	for _, value := range values {
		var matches bool
		if err := s.pool.QueryRow(ctx, `SELECT source_id=$2 FROM domain_local_measures WHERE id=$1`, value.ID, sourceID).Scan(&matches); err != nil {
			return nil, err
		}
		if matches {
			result = append(result, value)
		}
	}
	return result, nil
}

func (s *Store) searchDocuments(ctx context.Context, query SearchQuery) ([]Document, error) {
	additionalScope := ""
	if query.Kind == "attention" {
		additionalScope = " AND s.product_id='municipal' AND " + municipalNoticeSQL
	}
	rows, err := s.pool.Query(ctx, `SELECT DISTINCT ON (d.id) v.id
 FROM retained_documents d JOIN retained_versions v ON v.document_id=d.id
 JOIN `+s.sourcesSQL()+` s ON s.id=d.source_id
 WHERE (`+s.visibilitySQL()+`) AND v.first_acquired_at<=$1 AND ($2='' OR s.id=$2)
   AND ($3='' OR s.territory=$3 OR s.product_id<>'municipal')`+additionalScope+`
 ORDER BY d.id,v.first_acquired_at DESC,v.id DESC`, query.KnownAt, query.SourceID, query.MunicipalityISTAT)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var ids []int64
	for rows.Next() {
		var id int64
		if err = rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	rows.Close()
	if err = s.preloadDocumentVersions(ctx, ids, query.KnownAt); err != nil {
		return nil, err
	}
	result := []Document{}
	for _, id := range ids {
		value, loadErr := s.documentVersion(ctx, id, query.QueryTime)
		if loadErr != nil {
			return nil, loadErr
		}
		if (query.From != nil || query.To != nil) && (value.Publication.Precision == "unknown" || (value.Publication.Instant == nil && value.Publication.Date == nil)) {
			value.Quality.Interpretation.Limitations = append(value.Quality.Interpretation.Limitations, "publication time is imprecise; the requested publication interval cannot exclude this document")
		}
		if !intersects(value.Publication, query.From, query.To) {
			continue
		}
		result = append(result, value)
	}
	sort.Slice(result, func(i, j int) bool {
		if !result[i].AcquiredAt.Equal(result[j].AcquiredAt) {
			return result[i].AcquiredAt.Before(result[j].AcquiredAt)
		}
		return result[i].VersionID < result[j].VersionID
	})
	return result, nil
}
