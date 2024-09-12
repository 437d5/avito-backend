package postgres

import (
	"context"
	"fmt"
	"log"
	"zadanie-6105/internal/config"
	"zadanie-6105/internal/storage/models"

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

func (s *Storage) GetUsername(ctx context.Context, userID string) (string, error) {
	query := `
		SELECT username FROM public.employee
		WHERE id=$1;
	`

	row := s.Pool.QueryRow(ctx, query, userID)
	var username string
	err := row.Scan(&username)
	if err != nil {
		if err == pgx.ErrNoRows {
			return "", nil
		}

		return "", fmt.Errorf("cannot get username: %w", err)
	}

	return username, nil
}

func (s *Storage) TenderExists(ctx context.Context, tenderID string) (bool, error) {
	query := `
		SELECT id FROM tenders
		WHERE id=$1;
	`
	var id string
	row := s.Pool.QueryRow(ctx, query, tenderID)
	err := row.Scan(&id)
	if err != nil {
		if err == pgx.ErrNoRows && id == "" {
			return false, nil
		}

		return false, fmt.Errorf("cannot get tender: %w", err)
	}

	return true, nil
}

func (s *Storage) UserExists(ctx context.Context, userID string) (bool, error) {
	query := `
		SELECT id FROM public.employee
		WHERE id=$1;
	`

	row := s.Pool.QueryRow(ctx, query, userID)
	var username string
	err := row.Scan(&username)
	log.Println(err)
	if err != nil {
		if err == pgx.ErrNoRows {
			return false, nil
		}

		return false, fmt.Errorf("cannot get username: %w", err)
	}

	return true, nil
}

func (s *Storage) GetBidOrganizationID(ctx context.Context, bidID string) (string, error) {
	query := `
		SELECT organization_id 
		FROM bids
		WHERE id=$1
	`

	row := s.Pool.QueryRow(ctx, query, bidID)
	var organizationID string
	err := row.Scan(&organizationID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return "", nil
		}

		return "", err
	}

	return organizationID, nil
}

func (s *Storage) GetBidByID(ctx context.Context, bidID string) (models.Bid, error) {
	query := `
		SELECT id, name, status, author_type, author_id, version, created_at
		FROM bids
		WHERE id=$1;
	`

	row := s.Pool.QueryRow(ctx, query, bidID)
    var b models.Bid
    err := row.Scan(
        &b.ID, &b.Name, &b.Status, &b.AuthorType,
        &b.AuthorID, &b.Version, &b.CreatedAt,
    )
    log.Println(err)
    if err != nil {
        return models.Bid{}, fmt.Errorf("cannot edit row: %w", err)
    }

    return b, nil
}