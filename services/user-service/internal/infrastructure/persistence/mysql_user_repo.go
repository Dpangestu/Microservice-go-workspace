package persistence

import (
	"bkc_microservice/services/user-service/internal/domain/entities"
	"database/sql"
	"log"
)

type MySQLUserRepository struct {
	DB *sql.DB
}

func NewMySQLUserRepository(db *sql.DB) *MySQLUserRepository {
	return &MySQLUserRepository{DB: db}
}

func (r *MySQLUserRepository) ListAll() ([]*entities.User, error) {
	// PERBAIKAN: WHERE deleted_at IS NULL (sebelumnya OR deleted_at IS NULL redundant)
	rows, err := r.DB.Query(`
		SELECT id, email, username, created_at 
		FROM users 
		WHERE deleted_at IS NULL
	`)
	if err != nil {
		log.Printf("[MySQLUserRepository] Error in ListAll: %v", err)
		return nil, err
	}
	defer rows.Close()

	var out []*entities.User
	for rows.Next() {
		u := &entities.User{}
		if err := rows.Scan(&u.ID, &u.Email, &u.Username, &u.CreatedAt); err != nil {
			log.Printf("[MySQLUserRepository] Error scanning row: %v", err)
			return nil, err
		}
		out = append(out, u)
	}

	if err = rows.Err(); err != nil {
		log.Printf("[MySQLUserRepository] Error in ListAll rows: %v", err)
		return nil, err
	}

	return out, nil
}

func (r *MySQLUserRepository) FindByID(id string) (*entities.User, error) {
	row := r.DB.QueryRow(`
		SELECT 
			id, username, email, role_id, is_active, 
			is_locked, failed_login_attempts, last_login, created_at, updated_at
		FROM users 
		WHERE id = ?
	`, id)

	u := &entities.User{}
	err := row.Scan(
		&u.ID, &u.Username, &u.Email, &u.RoleID, &u.IsActive,
		&u.IsLocked, &u.FailedLoginAttempts, &u.LastLogin, &u.CreatedAt, &u.UpdatedAt)

	if err != nil {
		log.Printf("[MySQLUserRepository] Error in FindByID: %v", err)
		return nil, err
	}

	return u, nil
}

func (r *MySQLUserRepository) FindByEmail(email string) (*entities.User, error) {
	row := r.DB.QueryRow(`
		SELECT id, username, email, is_active, role_id, created_at, updated_at 
		FROM users 
		WHERE email = ?
	`, email)

	u := &entities.User{}
	err := row.Scan(&u.ID, &u.Username, &u.Email, &u.IsActive, &u.RoleID, &u.CreatedAt, &u.UpdatedAt)

	if err != nil {
		log.Printf("[MySQLUserRepository] Error in FindByEmail: %v", err)
		return nil, err
	}

	return u, nil
}

func (r *MySQLUserRepository) Create(u *entities.User) error {
	// PERBAIKAN: Sinkronkan kolom dan values
	_, err := r.DB.Exec(`
		INSERT INTO users 
		(id, username, email, is_active, role_id, created_at)
		VALUES (?, ?, ?, ?, ?, NOW())
	`, u.ID, u.Username, u.Email, u.IsActive, u.RoleID)

	if err != nil {
		log.Printf("[MySQLUserRepository] Error in Create: %v", err)
		return err
	}

	return nil
}

func (r *MySQLUserRepository) Update(u *entities.User) error {
	_, err := r.DB.Exec(`
		UPDATE users 
		SET username = ?, email = ?, is_active = ?, role_id = ?, updated_at = NOW()
		WHERE id = ?
	`, u.Username, u.Email, u.IsActive, u.RoleID, u.ID)

	if err != nil {
		log.Printf("[MySQLUserRepository] Error in Update: %v", err)
		return err
	}

	return nil
}

func (r *MySQLUserRepository) Delete(id string) error {
	_, err := r.DB.Exec(`
		UPDATE users 
		SET deleted_at = NOW(), is_active = FALSE 
		WHERE id = ?
	`, id)

	if err != nil {
		log.Printf("[MySQLUserRepository] Error in Delete: %v", err)
		return err
	}

	return nil
}
