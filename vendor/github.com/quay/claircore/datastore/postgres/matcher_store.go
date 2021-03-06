package postgres

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/quay/zlog"

	"github.com/quay/claircore"
	"github.com/quay/claircore/datastore"
	"github.com/quay/claircore/libvuln/driver"
)

// MatcherStore implements all interfaces in the vulnstore package
type MatcherStore struct {
	pool *pgxpool.Pool
	// Initialized is used as an atomic bool for tracking initialization.
	initialized uint32
}

func NewMatcherStore(pool *pgxpool.Pool) *MatcherStore {
	return &MatcherStore{
		pool: pool,
	}
}

var (
	_ datastore.Updater       = (*MatcherStore)(nil)
	_ datastore.Vulnerability = (*MatcherStore)(nil)
)

// UpdateVulnerabilities implements vulnstore.Updater.
func (s *MatcherStore) UpdateVulnerabilities(ctx context.Context, updater string, fingerprint driver.Fingerprint, vulns []*claircore.Vulnerability) (uuid.UUID, error) {
	return updateVulnerabilites(ctx, s.pool, updater, fingerprint, vulns)
}

// DeleteUpdateOperations implements vulnstore.Updater.
func (s *MatcherStore) DeleteUpdateOperations(ctx context.Context, id ...uuid.UUID) (int64, error) {
	const query = `DELETE FROM update_operation WHERE ref = ANY($1::uuid[]);`
	ctx = zlog.ContextWithValues(ctx, "component", "internal/vulnstore/postgres/deleteUpdateOperations")
	if len(id) == 0 {
		return 0, nil
	}

	// Pgx seems unwilling to do the []uuid.UUID → uuid[] conversion, so we're
	// forced to make some garbage here.
	refStr := make([]string, len(id))
	for i := range id {
		refStr[i] = id[i].String()
	}
	tag, err := s.pool.Exec(ctx, query, refStr)
	if err != nil {
		return 0, fmt.Errorf("failed to delete: %w", err)
	}
	return tag.RowsAffected(), nil
}
