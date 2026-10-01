package registry

import (
	"context"
	"github.com/Balestrino/italian-weather-alert/internal/backend/territory"
	"strings"
)

// CreateSourceWithIdentity is the guided setup path. Identity and source draft
// either commit together or leave the database unchanged.
func (s *Store) CreateSourceWithIdentity(ctx context.Context, src Source, c Configuration, actor string, a Authority, ch Channel) error {
	c.defaults()
	if src.ID == "" || src.Territory == "" || strings.TrimSpace(actor) == "" || !c.valid() || a.ID != src.AuthorityID || ch.ID != src.ChannelID || ch.PublisherID != a.ID || strings.TrimSpace(a.Name) == "" || !validURL(a.OfficialURL) || !validURL(ch.URL) || ch.External {
		return ErrInvalid
	}
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	if _, err = tx.Exec(ctx, `INSERT INTO registry_authorities(id,name,official_url) VALUES($1,$2,$3) ON CONFLICT DO NOTHING`, a.ID, a.Name, a.OfficialURL); err != nil {
		return err
	}
	var name, officialURL string
	if err = tx.QueryRow(ctx, `SELECT name,official_url FROM registry_authorities WHERE id=$1`, a.ID).Scan(&name, &officialURL); err != nil {
		return err
	}
	if name != a.Name || officialURL != a.OfficialURL {
		return ErrConflict
	}
	if _, err = tx.Exec(ctx, `INSERT INTO registry_channels(id,publisher_id,platform,url,external) VALUES($1,$2,$3,$4,false)`, ch.ID, ch.PublisherID, ch.Platform, ch.URL); err != nil {
		return err
	}
	if _, err = tx.Exec(ctx, `INSERT INTO registry_sources(id,authority_id,channel_id,product_id,territory) VALUES($1,$2,$3,$4,$5)`, src.ID, src.AuthorityID, src.ChannelID, src.ProductID, src.Territory); err != nil {
		return err
	}
	if err = insertConfig(ctx, tx, src.ID, 1, c, actor); err != nil {
		return err
	}
	if err = territory.Associate(ctx, tx, src.ID, src.ProductID, src.Territory, c.ProcessingProfile, actor); err != nil {
		return err
	}
	return tx.Commit(ctx)
}
func (s *Store) SourceTerritory(ctx context.Context, id string) (string, string, error) {
	var region, istat string
	err := s.pool.QueryRow(ctx, `SELECT region_code,COALESCE(municipality_istat,'') FROM territorial_current_sources WHERE source_id=$1`, id).Scan(&region, &istat)
	return region, istat, err
}
