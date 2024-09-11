package postgres

import (
	"context"
	"fmt"
	"zadanie-6105/internal/config"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// TODO add logger
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

// GetUserID return string with userID from table employee or error
// if error occurs
func (s *Storage) GetUserID(ctx context.Context, username string) (string, error) {
	var userID string
	query := `
		SELECT id FROM public.employee
		WHERE username=$1; 
	`
	row := s.Pool.QueryRow(ctx, query, username)
	err := row.Scan(&userID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return "", nil
		}

		return "", fmt.Errorf("cannot select id from employee: %w", err)
	}
	return userID, nil
}

func (s *Storage) GetOrganizationID(ctx context.Context, userID string) (string, error) {
	var organizationID string
	query := `
		SELECT organization_id
		FROM public.organization_responsible
		WHERE user_id = $1;
	`

	row := s.Pool.QueryRow(ctx, query, userID)
	err := row.Scan(&organizationID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return "", nil
		}

		return "", fmt.Errorf("cannot select id from organization_responsible: %w", err)
	}

	return organizationID, nil
}

func (s *Storage) GetOrganizationIDByTender(ctx context.Context, tenderID string) (string, error) {
	query := `
		SELECT organization_id
		FROM public.tenders
		WHERE id = $1;
	`

	var organizationID string
	row := s.Pool.QueryRow(ctx, query, tenderID)
	err := row.Scan(&organizationID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return "", nil
		}
		return "", fmt.Errorf("cannot select organization id from tenders: %w", err)
	}

	return organizationID, nil
}