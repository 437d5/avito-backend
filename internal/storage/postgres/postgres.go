package postgres

import (
	"context"
	"fmt"
	"log"
	"zadanie-6105/internal/storage/models"

	"github.com/jackc/pgx/v5"
)

const (
    StatusCreated = "Created"
    StatusPublished = "Published"
    StatusClosed = "Closed"
    StatusCanceled = "Canceled"
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

func (s *Storage) InsertBid(ctx context.Context, bid models.BidRequest, organizationID string) (models.Bid, error) {
    query := `
        INSERT INTO public.bids (tender_id, organization_id, name, description, status, author_type, author_id, version)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
        RETURNING id, name, status, author_type, author_id, version, created_at;
    `

    row := s.Pool.QueryRow(ctx, query, bid.TenderID, organizationID, bid.Name, bid.Description, 
        StatusCreated, bid.AuthorType, bid.AuthorID, 1)
    var b models.Bid
    err := row.Scan(
        &b.ID, &b.Name, &b.Status, &b.AuthorType,
        &b.AuthorID, &b.Version, &b.CreatedAt,
    )
    if err != nil {
        return models.Bid{}, fmt.Errorf("cannot insert new bid: %w", err)
    }

    return b, nil
}

func (s *Storage) GetMyBidsList(ctx context.Context, limit, offset int, userID string) ([]models.Bid, error) {
    query := `
        SELECT id, name, status, author_type, author_id, version, created_at
        FROM bids
        WHERE author_id=$1;
    `

    rows, err := s.Pool.Query(ctx, query, userID)
    if err != nil {
        return nil, fmt.Errorf("cannot get bids list: %w", err)
    }
    defer rows.Close()

    bids := make([]models.Bid, 0, limit)
    for rows.Next() {
        var b models.Bid
        err := rows.Scan(
            &b.ID, &b.Name, &b.Status, &b.AuthorType,
            &b.AuthorID, &b.Version, &b.CreatedAt,
        )
        if err != nil {
            return nil, fmt.Errorf("cannot scan row: %w", err)
        }
        bids = append(bids, b)
    }

    return bids, nil
}

func (s *Storage) GetTenderBids(ctx context.Context, limit, offset int, tenderID string) ([]models.Bid, error) {
    query := `
        SELECT id, name, status, author_type, author_id, version, created_at
        FROM bids
        WHERE tender_id=$1
        LIMIT $2 offset $3;
    `

    rows, err := s.Pool.Query(ctx, query, tenderID, limit, offset)
    if err != nil {
        return nil, fmt.Errorf("cannot get bids list: %w", err)
    }
    defer rows.Close()

    bids := make([]models.Bid, 0, limit)
    for rows.Next() {
        var b models.Bid
        err := rows.Scan(
            &b.ID, &b.Name, &b.Status, &b.AuthorType,
            &b.AuthorID, &b.Version, &b.CreatedAt,
        )
        if err != nil {            
            return nil, fmt.Errorf("cannot scan row: %w", err)
        }
        bids = append(bids, b)
    }

    return bids, nil
}

func (s *Storage) GetBidStatus(ctx context.Context, bidID string) (string, error) {
    query := `
        SELECT status
        FROM bids
        WHERE id=$1;
    `

    row := s.Pool.QueryRow(ctx, query, bidID)
    var status string
    err := row.Scan(&status)
    if err != nil {
        if err == pgx.ErrNoRows {
            return "", nil
        }

        return "", err
    }

    switch status {
    case StatusCreated:
        status = "Created"
    case StatusPublished:
        status = "Published"
    case StatusCanceled:
        status = "Canceled"
    }

    return status, nil
}

func (s *Storage) ChangeBitStatus(ctx context.Context, bidID, status string) (models.Bid, error) {
    query := `
        UPDATE bids SET status=$1
        WHERE id=$2
        RETURNING id, name, status, author_type, author_id, version, created_at;
    `

    row := s.Pool.QueryRow(ctx, query, status, bidID)
    var b models.Bid
    err := row.Scan(
        &b.ID, &b.Name, &b.Status, &b.AuthorType,
        &b.AuthorID, &b.Version, &b.CreatedAt,
    )
    if err != nil {
        return models.Bid{}, fmt.Errorf("cannot update row: %w", err)
    }

    return b, nil
}

func (s *Storage) EditBid(ctx context.Context, bidID, changeStr string) (models.Bid, error) {
    query := fmt.Sprintf("UPDATE bids SET %s WHERE id=$1 RETURNING id, name, status, author_type, author_id, version, created_at;", changeStr)
    log.Println(query)

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

    log.Println(b)
    return b, nil
}

func (s *Storage) BidDecision(ctx context.Context, bidID, decision string) (models.Bid, error) {
    var query string
    if decision == "Approved" {
        query = `
            UPDATE submissions s
            SET accept_rate = accept_rate + 1
            FROM bids b
            WHERE s.bid_id = b.id AND b.id = $1
            RETURNING b.id, b.name, b.status, b.author_type, b.author_id, b.version, b.created_at;
        `     
    } else {
        query = `
            UPDATE submissions s
            SET rejected = true
            FROM bids b
            WHERE s.bid_id = b.id AND s.bid_id = $1
            RETURNING b.id, b.name, b.status, b.author_type, b.author_id, b.version, b.created_at;
        `
    }

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

func (s *Storage) SendFeedback(ctx context.Context, bidID, userID, feedback string)  error {
    query := `
        INSERT INTO feedback (bid_id, feedback, creator_id)
        VALUES ($1, $2, $3);
    `
    _, err := s.Pool.Exec(ctx, query, bidID, feedback, userID)
    if err != nil {
        return err
    }

    return nil
}

func (s *Storage) RollbackBid(ctx context.Context, bidID string, version int) error {
    conn, err := s.Pool.Acquire(ctx)
    if err != nil {
        log.Println("Error acquiring connection:", err)
        return err
    }
    defer conn.Release()

    transaction, err := conn.BeginTx(ctx, pgx.TxOptions{})
    if err != nil {
        return err
    }

    _, err = transaction.Exec(ctx, "ALTER TABLE bids DISABLE TRIGGER bid_version_trigger")
    if err != nil {
        return transaction.Rollback(ctx)
    }

    updateQuery := `
        UPDATE bids
        SET tender_id = bh.tender_id,
            organization_id = bh.organization_id,
            name = bh.name,
            description = bh.description,
            status = bh.status,
            author_type = bh.author_type,
            author_id = bh.author_id,
            version = $1,
            created_at = bh.created_at
        FROM bids_history bh
        WHERE bh.bid_id = $2 AND bh.version = $1 AND bids.id = bh.bid_id
    `

    _, err = transaction.Exec(ctx, updateQuery, version, bidID)
    if err != nil {
        log.Println("Error updating bid:", err)
        return transaction.Rollback(ctx)
    }

    deleteQuery := `
        DELETE FROM bids_history
        WHERE bid_id = $1 AND version >= $2;
    `

    _, err = transaction.Exec(ctx, deleteQuery, bidID, version)
    if err != nil {
        log.Println("Error deleting history:", err)
        return transaction.Rollback(ctx)
    }

    _, err = transaction.Exec(ctx, "ALTER TABLE bids ENABLE TRIGGER bid_version_trigger")
    if err != nil {
        return transaction.Rollback(ctx)
    }

    if err := transaction.Commit(ctx); err != nil {
        log.Println("Error committing transaction:", err)
        return err
    }

    return nil
}

// FIXME не работаю
func (s *Storage) GetFeedback(ctx context.Context, authorUserID string, limit, offset int) ([]models.Feedback, error) {    
    query := `
        SELECT f.id, f.feedback, f.created_at        
        FROM bids b
        JOIN feedback f ON b.id = f.bid_id        
        WHERE b.author_id = $1
        LIMIT $2 OFFSET $3;  
      `
    rows, err := s.Pool.Query(ctx, query, authorUserID, limit, offset)
    if err != nil {        
        return nil, fmt.Errorf("cannot get feedback list: %w", err)
    }    
    defer rows.Close()

    feedbackList := make([]models.Feedback, 0, limit)
    for rows.Next() {
        var f models.Feedback        
        err := rows.Scan(
            &f.ID, &f.Description, &f.CreatedAt,        
        )
        if err != nil {            
            return nil, fmt.Errorf("cannot scan row: %w", err)
        }        
        feedbackList = append(feedbackList, f)
    }

    if err := rows.Err(); err != nil {        
        return nil, fmt.Errorf("error while reading rows: %w", err)
    }

    return feedbackList, nil
}
