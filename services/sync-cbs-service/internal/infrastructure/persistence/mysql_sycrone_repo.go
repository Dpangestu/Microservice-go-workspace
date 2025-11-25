package persistence

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	"bkc_microservice/services/sync-cbs-service/internal/domain/entities"

	"github.com/google/uuid"
)

type MySQLSycroneCoreRepository struct {
	db *sql.DB
}

func NewMySQLSycroneCoreRepository(db *sql.DB) *MySQLSycroneCoreRepository {
	return &MySQLSycroneCoreRepository{db: db}
}

func (r *MySQLSycroneCoreRepository) Create(ctx context.Context, sc *entities.SycroneCore) error {
	if sc.ID == "" {
		sc.ID = uuid.New().String()
	}

	query := `
		INSERT INTO sycrone_core 
		(id, user_id, user_core, kode_group_1, kode_perkiraan, kode_cabang, status, sync_status, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, NOW(), NOW())
	`

	_, err := r.db.ExecContext(ctx, query,
		sc.ID, sc.UserID, sc.UserCore, sc.KodeGroup1, sc.KodePerkiraan,
		sc.KodeCabang, sc.Status, sc.SyncStatus,
	)

	if err != nil {
		log.Printf("[MySQLSycroneCoreRepo] Create error: %v", err)
		return fmt.Errorf("failed to create sycrone core: %w", err)
	}

	return nil
}

func (r *MySQLSycroneCoreRepository) GetByUserID(ctx context.Context, userID string) (*entities.SycroneCore, error) {
	query := `
		SELECT id, user_id, user_core, kode_group_1, kode_perkiraan, kode_cabang,
		       status, sync_status, last_sync_at, created_at, updated_at
		FROM sycrone_core
		WHERE user_id = ?
		LIMIT 1
	`

	var sc entities.SycroneCore
	err := r.db.QueryRowContext(ctx, query, userID).Scan(
		&sc.ID, &sc.UserID, &sc.UserCore, &sc.KodeGroup1, &sc.KodePerkiraan,
		&sc.KodeCabang, &sc.Status, &sc.SyncStatus, &sc.LastSyncAt,
		&sc.CreatedAt, &sc.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("sycrone_core not found for user_id: %s", userID)
	}

	if err != nil {
		log.Printf("[MySQLSycroneCoreRepo] GetByUserID error: %v", err)
		return nil, fmt.Errorf("failed to get sycrone core: %w", err)
	}

	return &sc, nil
}

func (r *MySQLSycroneCoreRepository) GetByUserCore(ctx context.Context, userCore string) (*entities.SycroneCore, error) {
	query := `
		SELECT id, user_id, user_core, kode_group_1, kode_perkiraan, kode_cabang,
		       status, sync_status, last_sync_at, created_at, updated_at
		FROM sycrone_core
		WHERE user_core = ?
		LIMIT 1
	`

	var sc entities.SycroneCore
	err := r.db.QueryRowContext(ctx, query, userCore).Scan(
		&sc.ID, &sc.UserID, &sc.UserCore, &sc.KodeGroup1, &sc.KodePerkiraan,
		&sc.KodeCabang, &sc.Status, &sc.SyncStatus, &sc.LastSyncAt,
		&sc.CreatedAt, &sc.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("sycrone_core not found for user_core: %s", userCore)
	}

	if err != nil {
		log.Printf("[MySQLSycroneCoreRepo] GetByUserCore error: %v", err)
		return nil, fmt.Errorf("failed to get sycrone core: %w", err)
	}

	return &sc, nil
}

func (r *MySQLSycroneCoreRepository) Update(ctx context.Context, sc *entities.SycroneCore) error {
	query := `
		UPDATE sycrone_core
		SET user_core = ?, kode_group_1 = ?, kode_perkiraan = ?, kode_cabang = ?,
		    status = ?, sync_status = ?, last_sync_at = NOW(), updated_at = NOW()
		WHERE id = ?
	`

	result, err := r.db.ExecContext(ctx, query,
		sc.UserCore, sc.KodeGroup1, sc.KodePerkiraan, sc.KodeCabang,
		sc.Status, sc.SyncStatus, sc.ID,
	)

	if err != nil {
		log.Printf("[MySQLSycroneCoreRepo] Update error: %v", err)
		return fmt.Errorf("failed to update sycrone core: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil || rows == 0 {
		return fmt.Errorf("no rows updated for id: %s", sc.ID)
	}

	return nil
}

func (r *MySQLSycroneCoreRepository) ListByStatus(ctx context.Context, status string, page, size int) ([]*entities.SycroneCore, int, error) {
	countQuery := `SELECT COUNT(*) FROM sycrone_core WHERE sync_status = ?`
	var total int
	err := r.db.QueryRowContext(ctx, countQuery, status).Scan(&total)
	if err != nil {
		log.Printf("[MySQLSycroneCoreRepo] Count error: %v", err)
		return nil, 0, fmt.Errorf("failed to count rows: %w", err)
	}

	if page < 1 {
		page = 1
	}
	if size < 1 {
		size = 10
	}
	offset := (page - 1) * size

	query := `
		SELECT id, user_id, user_core, kode_group_1, kode_perkiraan, kode_cabang,
		       status, sync_status, last_sync_at, created_at, updated_at
		FROM sycrone_core
		WHERE sync_status = ?
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`

	rows, err := r.db.QueryContext(ctx, query, status, size, offset)
	if err != nil {
		log.Printf("[MySQLSycroneCoreRepo] Query error: %v", err)
		return nil, 0, fmt.Errorf("failed to query sycrone core: %w", err)
	}
	defer rows.Close()

	var items []*entities.SycroneCore
	for rows.Next() {
		sc := &entities.SycroneCore{}
		err := rows.Scan(
			&sc.ID, &sc.UserID, &sc.UserCore, &sc.KodeGroup1, &sc.KodePerkiraan,
			&sc.KodeCabang, &sc.Status, &sc.SyncStatus, &sc.LastSyncAt,
			&sc.CreatedAt, &sc.UpdatedAt,
		)

		if err != nil {
			log.Printf("[MySQLSycroneCoreRepo] Scan error: %v", err)
			return nil, 0, fmt.Errorf("failed to scan row: %w", err)
		}

		items = append(items, sc)
	}

	if err := rows.Err(); err != nil {
		log.Printf("[MySQLSycroneCoreRepo] Rows error: %v", err)
		return nil, 0, fmt.Errorf("rows error: %w", err)
	}

	return items, total, nil
}

func (r *MySQLSycroneCoreRepository) Delete(ctx context.Context, userID string) error {
	query := "DELETE FROM sycrone_core WHERE user_id = ?"
	result, err := r.db.ExecContext(ctx, query, userID)
	if err != nil {
		log.Printf("[MySQLSycroneCoreRepo] Delete error: %v", err)
		return fmt.Errorf("failed to delete: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil || rows == 0 {
		return fmt.Errorf("no rows deleted for user_id: %s", userID)
	}

	return nil
}
