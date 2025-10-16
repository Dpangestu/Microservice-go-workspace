package persistence

import (
	"bkc_microservice/services/user-service/internal/domain/entities"
	"database/sql"
)

type MySQLUserRepository struct {
	DB *sql.DB
}

func NewMySQLUserRepository(db *sql.DB) *MySQLUserRepository {
	return &MySQLUserRepository{DB: db}
}

func (r *MySQLUserRepository) FindByID(id string) (*entities.User, error) {
	row := r.DB.QueryRow("SELECT id, company_id, username, email, first_name, last_name, avatar_url, phone, is_active, role_id, created_at, updated_at FROM users WHERE id = ?", id)
	u := &entities.User{}
	err := row.Scan(&u.ID, &u.CompanyID, &u.Username, &u.Email, &u.FirstName, &u.LastName, &u.AvatarURL, &u.Phone, &u.IsActive, &u.RoleID, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return u, nil
}

func (r *MySQLUserRepository) FindByEmail(email string) (*entities.User, error) {
	row := r.DB.QueryRow("SELECT id, company_id, username, email, first_name, last_name, avatar_url, phone, is_active, role_id, created_at, updated_at FROM users WHERE email = ?", email)
	u := &entities.User{}
	err := row.Scan(&u.ID, &u.CompanyID, &u.Username, &u.Email, &u.FirstName, &u.LastName, &u.AvatarURL, &u.Phone, &u.IsActive, &u.RoleID, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return u, nil
}

func (r *MySQLUserRepository) Create(u *entities.User) error {
	_, err := r.DB.Exec(`INSERT INTO users (id, company_id, username, email, first_name, last_name, avatar_url, phone, is_active, role_id, created_at) 
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, NOW())`,
		u.ID, u.CompanyID, u.Username, u.Email, u.FirstName, u.LastName, u.AvatarURL, u.Phone, u.IsActive, u.RoleID)
	return err
}

func (r *MySQLUserRepository) Update(u *entities.User) error {
	_, err := r.DB.Exec(`UPDATE users SET username=?, email=?, first_name=?, last_name=?, avatar_url=?, phone=?, is_active=?, role_id=?, updated_at=NOW() WHERE id=?`,
		u.Username, u.Email, u.FirstName, u.LastName, u.AvatarURL, u.Phone, u.IsActive, u.RoleID, u.ID)
	return err
}

func (r *MySQLUserRepository) Delete(id string) error {
	_, err := r.DB.Exec(`UPDATE users SET deleted_at=NOW(), is_active=FALSE WHERE id=?`, id)
	return err
}
