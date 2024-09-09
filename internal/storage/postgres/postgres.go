package postgres

import (
	"context"
	"fmt"
	"zadanie-6105/internal/config"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Storage helds the pointer to pool of connections to postgres
type Storage struct {
	Pool *pgxpool.Pool
}

// New creates new pool of connections to database if POSTGRES_CONN or
// username, password, host, port and dbname env variables were provided
func New(ctx context.Context, cfg config.Config) (*Storage, error) {
	if cfg.POSTGRES_CONN != "" {
		pool, err := pgxpool.New(ctx, cfg.POSTGRES_CONN)
		if err != nil {
			return nil, err
		}

		err = pool.Ping(ctx)
		if err != nil {
			return nil, err
		}

		return &Storage{Pool: pool}, nil
	} else {
		connString := fmt.Sprintf("postgres://%s:%s@%s:%s/%s",
	cfg.POSTGRES_USERNAME, cfg.POSTGRES_PASSWORD, cfg.POSTGRES_HOST, cfg.POSTGRES_PORT, cfg.POSTGRES_DATABASE)
		pool, err := pgxpool.New(ctx, connString)
		if err != nil {
			return nil, err
		}

		err = pool.Ping(ctx)
		if err != nil {
			return nil, err
		}
		return &Storage{Pool: pool}, nil
	}

}