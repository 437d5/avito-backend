package postgres

import (
	"context"
	"fmt"
	"log"
	"zadanie-6105/internal/storage/models"

	"github.com/jackc/pgx/v5"
)

const (
    StatusCreated = "CREATED"
    StatusPublished = "PUBLISHED"
    StatusClosed = "CLOSED"
)

// GetTenderList takes limit, offset and service_type to return
// list of tenders with provided params
func (s *Storage) GetTenderList(ctx context.Context, limit, offset int, service_type string) ([]models.Tender, error) {
    var query string
    if service_type != "" {
        query = `
            SELECT id, name, description, status, service_type, version, created_at
            FROM tenders
            WHERE service_type = $1
            ORDER BY name ASC
            LIMIT $2 OFFSET $3
		`
        
    } else {
        query = `
            SELECT id, name, description, status, service_type, version, created_at
            FROM tenders
            ORDER BY name ASC
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

// IsertTender takes NewTenderRequest and creatorID then insert
// new tender into database returning new tender and error if occurs 
func (s *Storage) InsertTender(ctx context.Context, newTender *models.NewTenderRequest, creatorID string) (models.Tender, error) {
    query := `
        INSERT INTO public.tenders (name, organization_id, creator_id, description, status, service_type, version)
        VALUES ($1, $2, $3, $4, $5, $6, $7)
        RETURNING id, name, description, status, service_type, version, created_at;
    `
    
    var tender models.Tender

    row := s.Pool.QueryRow(ctx, query, newTender.Name, newTender.OrganizationID, creatorID, newTender.Description, StatusCreated, newTender.ServiceType, 1)
    err := row.Scan(
        &tender.ID, &tender.Name, &tender.Description, &tender.Status,
        &tender.ServiceType, &tender.Version, &tender.CreatedAt,
    )

    if err != nil {
        return models.Tender{}, fmt.Errorf("cannot insert new tender: %w", err)
    }

    return tender, nil
}

// GetMyTendersList takes limit, offset, userID and return 
// slice of tenders where creator_id=userID and error if occurs
func (s *Storage) GetMyTendersList(ctx context.Context, limit, offset int, userID string) ([]models.Tender, error) {
    query := `
        SELECT id, name, description, status, service_type, version, created_at
        FROM tenders
        WHERE creator_id = $1
        ORDER BY name ASC
        LIMIT $2 OFFSET $3
    `

    rows, err := s.Pool.Query(ctx, query, userID, limit, offset)
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

// GetTenderStatus takes tenderID and return 
// status value or error if occurs
func (s *Storage) GetTenderStatus(ctx context.Context, tenderID string) (string, error) {
    query := `
        select status
        from public.tenders
        where id = $1
    `

    var status string
    row := s.Pool.QueryRow(ctx, query, tenderID)
    err := row.Scan(&status)
    if err != nil {
        if err == pgx.ErrNoRows {
            return "", nil
        }

        return "", fmt.Errorf("cannot select status: %w", err)
    }

    return status, nil
}

func (s *Storage) ChangeTenderStatus(ctx context.Context, tenderID, status string) (models.Tender, error) {
    switch status {
    case "Created":
        status = StatusCreated
    case "Published":
        status = StatusPublished
    case "Closed":
        status = StatusClosed
    }

    query := `
        UPDATE tenders SET status=$1
        WHERE id=$2
        RETURNING id, name, description, status, service_type, version, created_at;
    `

    var tender models.Tender
    row := s.Pool.QueryRow(ctx, query, status, tenderID)
    err := row.Scan(
        &tender.ID, &tender.Name, &tender.Description,
        &tender.Status, &tender.ServiceType, &tender.Version, &tender.CreatedAt,
    )
    if err != nil {
        if err == pgx.ErrNoRows {
            return models.Tender{}, nil
        }

        return models.Tender{}, fmt.Errorf("cannot update tender: %w", err)
    }

    return tender, nil
}

func (s *Storage) EditTender(ctx context.Context, tenderID, changeStr string) (models.Tender, error) {
    query := fmt.Sprintf("UPDATE tenders SET %s WHERE id=$1 RETURNING id, name, description, status, service_type, version, created_at;", changeStr)

    var tender models.Tender
    row := s.Pool.QueryRow(ctx, query, tenderID)
    err := row.Scan(
        &tender.ID, &tender.Name, &tender.Description,
        &tender.Status, &tender.ServiceType, &tender.Version, &tender.CreatedAt,
    )
    if err != nil {
        if err == pgx.ErrNoRows {
            return models.Tender{}, nil
        }

        return models.Tender{}, fmt.Errorf("cannot change tender: %w", err)
    }

    return tender, nil
}

func (s *Storage) RollbackTender(ctx context.Context, tenderID, version string) (models.Tender, error) {
    conn, err := s.Pool.Acquire(ctx)
    if err != nil {
        log.Println("Error acquiring connection:", err)
        return models.Tender{}, err
    }
    defer conn.Release()

    transaction, err := conn.BeginTx(ctx, pgx.TxOptions{})
    if err != nil {
        log.Println("Error beginning transaction:", err)
        return models.Tender{}, err
    }

    // Отключаем триггер
    _, err = transaction.Exec(ctx, "ALTER TABLE tenders DISABLE TRIGGER tender_version_trigger")
    if err != nil {
        log.Println("Error disabling trigger:", err)
        return models.Tender{}, transaction.Rollback(ctx)
    }

    // Запрос на обновление тендера
    updateQuery := `
        UPDATE tenders
        SET organization_id = th.organization_id,
            creator_id = th.creator_id,
            name = th.name,
            description = th.description,
            status = th.status,
            service_type = th.service_type,
            created_at = th.created_at,
            version = $2
        FROM tenders_history th
        WHERE th.tender_id = $1 AND th.version = $2 AND tenders.id = th.tender_id;
    `
    
    
    _, err = transaction.Exec(ctx, updateQuery, tenderID, version)
    if err != nil {
        log.Println("Error updating tender:", err)
        return models.Tender{}, transaction.Rollback(ctx)
    }

    // Удаляем все версии после той и ту, к которой мы делаем роллбэк
    deleteQuery := `
        DELETE FROM tenders_history
        WHERE tender_id = $1 AND version >= $2;
    `
    
    _, err = transaction.Exec(ctx, deleteQuery, tenderID, version)
    if err != nil {
        log.Println("Error deleting history:", err)
        return models.Tender{}, transaction.Rollback(ctx)
    }

    // Включаем триггер обратно
    _, err = transaction.Exec(ctx, "ALTER TABLE tenders ENABLE TRIGGER tender_version_trigger")
    if err != nil {
        log.Println("Error enabling trigger:", err)
        return models.Tender{}, transaction.Rollback(ctx)
    }

    // Коммитим транзакцию
    if err := transaction.Commit(ctx); err != nil {
        log.Println("Error committing transaction:", err)
        return models.Tender{}, err
    }

    // Получаем обновленный тендер
    var tender models.Tender
    selectQuery := `
        SELECT id, name, description, status, service_type, version, created_at
        FROM tenders
        WHERE id = $1;
    `
    
    err = conn.QueryRow(ctx, selectQuery, tenderID).Scan(
        &tender.ID, &tender.Name, &tender.Description,
        &tender.Status, &tender.ServiceType, &tender.Version, &tender.CreatedAt,
    )
    
    if err != nil {
        log.Println("Error retrieving updated tender:", err)
        return models.Tender{}, err
    }

    return tender, nil
}
