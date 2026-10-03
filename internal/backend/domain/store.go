package domain

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Store struct{ pool *pgxpool.Pool }

func New(pool *pgxpool.Pool) *Store { return &Store{pool: pool} }

type sourceIdentity struct {
	authorityID, publisherID, platform, product, territory string
}

func source(ctx context.Context, q interface {
	QueryRow(context.Context, string, ...any) pgx.Row
}, id string) (sourceIdentity, error) {
	var value sourceIdentity
	err := q.QueryRow(ctx, `SELECT s.authority_id,c.publisher_id,c.platform,s.product_id,s.territory
 FROM registry_sources s JOIN registry_channels c ON c.id=s.channel_id WHERE s.id=$1`, id).Scan(&value.authorityID, &value.publisherID, &value.platform, &value.product, &value.territory)
	return value, err
}

func versionBelongsToSource(ctx context.Context, q interface {
	QueryRow(context.Context, string, ...any) pgx.Row
}, versionID int64, sourceID string) (bool, error) {
	var matches bool
	err := q.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM retained_versions v JOIN retained_documents d ON d.id=v.document_id WHERE v.id=$1 AND d.source_id=$2)`, versionID, sourceID).Scan(&matches)
	return matches, err
}

func (s *Store) PutRegional(ctx context.Context, value RegionalRecord) error {
	if !validRegional(value) {
		return ErrInvalid
	}
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	identity, err := source(ctx, tx, value.SourceID)
	if err != nil {
		return err
	}
	matches, err := versionBelongsToSource(ctx, tx, value.DocumentVersionID, value.SourceID)
	if err != nil {
		return err
	}
	if !matches {
		return ErrInvalid
	}
	municipalRepublication := identity.product == "municipal" && identity.authorityID != value.OriginatingAuthorityID
	if identity.platform == "cittadino-informato" || identity.product != value.Product && !municipalRepublication {
		return ErrInvalid
	}
	tag, err := tx.Exec(ctx, `INSERT INTO domain_regional_records(id,document_version_id,source_id,product,originating_authority_id,publisher_id,platform,municipal_republication)
 VALUES($1,$2,$3,$4,$5,$6,$7,$8) ON CONFLICT DO NOTHING`, value.ID, value.DocumentVersionID, value.SourceID, value.Product, value.OriginatingAuthorityID, identity.publisherID, identity.platform, municipalRepublication)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		var version int64
		var product, origin string
		if err = tx.QueryRow(ctx, "SELECT document_version_id,product,originating_authority_id FROM domain_regional_records WHERE id=$1", value.ID).Scan(&version, &product, &origin); err != nil {
			return err
		}
		if version != value.DocumentVersionID || product != value.Product || origin != value.OriginatingAuthorityID {
			return ErrConflict
		}
		return tx.Commit(ctx)
	}
	for _, fact := range value.Facts {
		if _, err = tx.Exec(ctx, `INSERT INTO domain_regional_facts(regional_record_id,product,ordinal,risk,official_risk_label,zone,level)
 VALUES($1,$2,$3,$4,$5,$6,$7)`, value.ID, value.Product, fact.Ordinal, fact.Risk, fact.OfficialRiskLabel, fact.Zone, fact.Level); err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

func (s *Store) Regional(ctx context.Context, id string) (RegionalRecord, bool, error) {
	var value RegionalRecord
	err := s.pool.QueryRow(ctx, `SELECT id,document_version_id,source_id,product,originating_authority_id,publisher_id,platform,municipal_republication
 FROM domain_regional_records WHERE id=$1`, id).Scan(&value.ID, &value.DocumentVersionID, &value.SourceID, &value.Product, &value.OriginatingAuthorityID, &value.PublisherID, &value.Platform, &value.MunicipalRepublication)
	if errors.Is(err, pgx.ErrNoRows) {
		return RegionalRecord{}, false, nil
	}
	if err != nil {
		return RegionalRecord{}, false, err
	}
	rows, err := s.pool.Query(ctx, `SELECT ordinal,risk,official_risk_label,zone,level FROM domain_regional_facts WHERE regional_record_id=$1 ORDER BY ordinal`, id)
	if err != nil {
		return RegionalRecord{}, false, err
	}
	defer rows.Close()
	for rows.Next() {
		var fact RegionalFact
		if err = rows.Scan(&fact.Ordinal, &fact.Risk, &fact.OfficialRiskLabel, &fact.Zone, &fact.Level); err != nil {
			return RegionalRecord{}, false, err
		}
		value.Facts = append(value.Facts, fact)
	}
	return value, true, rows.Err()
}

func (s *Store) PutLocalMeasure(ctx context.Context, value LocalMeasure) error {
	if !validLocal(value) {
		return ErrInvalid
	}
	identity, err := source(ctx, s.pool, value.SourceID)
	if err != nil {
		return err
	}
	matches, err := versionBelongsToSource(ctx, s.pool, value.DocumentVersionID, value.SourceID)
	if err != nil {
		return err
	}
	if identity.platform == "cittadino-informato" || identity.product != "municipal" || identity.territory != value.MunicipalityISTAT {
		return ErrInvalid
	}
	if !matches {
		return ErrInvalid
	}
	_, err = s.pool.Exec(ctx, `INSERT INTO domain_local_measures(id,document_version_id,source_id,municipality_istat,issuing_authority_id,publisher_id,platform,kind,subject,place,regional_record_id)
 VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)`, value.ID, value.DocumentVersionID, value.SourceID, value.MunicipalityISTAT, identity.authorityID, identity.publisherID, identity.platform, value.Kind, value.Subject, value.Place, value.RegionalRecordID)
	return err
}

func (s *Store) LocalMeasure(ctx context.Context, id string) (LocalMeasure, bool, error) {
	var value LocalMeasure
	err := s.pool.QueryRow(ctx, `SELECT id,document_version_id,source_id,municipality_istat,issuing_authority_id,publisher_id,platform,kind,subject,place,regional_record_id
 FROM domain_local_measures WHERE id=$1`, id).Scan(&value.ID, &value.DocumentVersionID, &value.SourceID, &value.MunicipalityISTAT, &value.IssuingAuthorityID, &value.PublisherID, &value.Platform, &value.Kind, &value.Subject, &value.Place, &value.RegionalRecordID)
	if errors.Is(err, pgx.ErrNoRows) {
		return LocalMeasure{}, false, nil
	}
	return value, err == nil, err
}

func (s *Store) PutOperationalPhase(ctx context.Context, value OperationalPhase) error {
	if !validPhase(value) {
		return ErrInvalid
	}
	identity, err := source(ctx, s.pool, value.SourceID)
	if err != nil {
		return err
	}
	matches, err := versionBelongsToSource(ctx, s.pool, value.DocumentVersionID, value.SourceID)
	if err != nil {
		return err
	}
	if identity.platform == "cittadino-informato" || identity.product != "municipal" || identity.territory != value.MunicipalityISTAT {
		return ErrInvalid
	}
	if !matches {
		return ErrInvalid
	}
	_, err = s.pool.Exec(ctx, `INSERT INTO domain_operational_phases(id,document_version_id,source_id,municipality_istat,authority_id,publisher_id,platform,phase,regional_record_id)
 VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9)`, value.ID, value.DocumentVersionID, value.SourceID, value.MunicipalityISTAT, identity.authorityID, identity.publisherID, identity.platform, value.Phase, value.RegionalRecordID)
	return err
}

func (s *Store) OperationalPhase(ctx context.Context, id string) (OperationalPhase, bool, error) {
	var value OperationalPhase
	err := s.pool.QueryRow(ctx, `SELECT id,document_version_id,source_id,municipality_istat,authority_id,publisher_id,platform,phase,regional_record_id
 FROM domain_operational_phases WHERE id=$1`, id).Scan(&value.ID, &value.DocumentVersionID, &value.SourceID, &value.MunicipalityISTAT, &value.AuthorityID, &value.PublisherID, &value.Platform, &value.Phase, &value.RegionalRecordID)
	if errors.Is(err, pgx.ErrNoRows) {
		return OperationalPhase{}, false, nil
	}
	return value, err == nil, err
}
