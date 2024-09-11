package postgres

import (
	"context"
	"fmt"
	"zadanie-6105/internal/config"
	"zadanie-6105/internal/storage/models"

	"github.com/jackc/pgx/v5"
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

// FIXME: rewrite this query to return tender slice
// GetTenderList takes limit, offset and service_type to return
// list of tenders with provided params
func (s *Storage) GetTenderList(ctx context.Context, limit, offset int, service_type string) ([]models.Tender, error) {
    var query string
    if service_type != "" {
        query = `
            SELECT id, name, description, status, service_type, version, created_at
            FROM tenders
            WHERE service_type = $1
            LIMIT $2 OFFSET $3
		`
        
    } else {
        query = `
            SELECT id, name, description, status, service_type, version, created_at
            FROM tenders
            LIMIT $1 OFFSET $2
		`  
    }

    var rows pgx.Rows
    var err error
    if service_type != "" {
        rows, err = s.Pool.Query(ctx, query, service_type, limit, offset)
    } else {
        rows, err = s.Pool.Query(ctx, query, limit, offset)
    }
    
    if err != nil {
        return nil, fmt.Errorf("cannot get tender list: %w", err)
    }
    defer rows.Close()

    tenders := make([]models.Tender, 0, limit)
    for rows.Next() {
        var t models.Tender
        err := rows.Scan(
            &t.ID, &t.Name, &t.Description, &t.Status,
            &t.ServiceType, &t.Version, &t.CreatedAt,
        )
        if err != nil {
            return nil, fmt.Errorf("cannot scan row: %w", err)
        }
        tenders = append(tenders, t)
    }

    if err := rows.Err(); err != nil {
        return nil, fmt.Errorf("error while reading rows: %w", err)
    }

    return tenders, nil
}
