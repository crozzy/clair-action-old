package libvuln

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/jackc/pgx/v4/stdlib"
	"github.com/quay/claircore/datastore"
	"github.com/quay/claircore/datastore/postgres"
	"github.com/quay/claircore/datastore/postgres/migrations"
	"github.com/remind101/migrate"
)

// InitPostgresStore initialize a indexer.Store given libindex.Opts
func InitPostgresStore(_ context.Context, pool *pgxpool.Pool, doMigration bool) (datastore.MatcherStore, error) {
	cfg, err := pgx.ParseConfig(pool.Config().ConnConfig.ConnString())
	if err != nil {
		return nil, err
	}
	db, err := sql.Open("pgx", stdlib.RegisterConnConfig(cfg))
	if err != nil {
		return nil, err
	}
	defer db.Close()

	// do migrations if requested
	if doMigration {
		migrator := migrate.NewPostgresMigrator(db)
		migrator.Table = migrations.MatcherMigrationTable
		err := migrator.Exec(migrate.Up, migrations.MatcherMigrations...)
		if err != nil {
			return nil, fmt.Errorf("failed to perform migrations: %w", err)
		}
	}

	store := postgres.NewMatcherStore(pool)
	return store, nil
}
