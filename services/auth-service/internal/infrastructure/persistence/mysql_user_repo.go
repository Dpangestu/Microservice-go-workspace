package persistence

import (
	"context"
	"database/sql"
	"errors"
	"log"

	"bkc_microservice/services/auth-service/internal/domain/entities"
	"bkc_microservice/services/auth-service/internal/domain/repositories"

	"golang.org/x/crypto/bcrypt"
)

type MySQLUserRepo struct{ db *sql.DB }

func NewMySQLUserRepo(db *sql.DB) repositories.UserRepository {
	return &MySQLUserRepo{db: db}
}

func (r *MySQLUserRepo) FindByEmail(ctx context.Context, email string) (*entities.User, error) {
	row := r.db.QueryRowContext(ctx, `SELECT id, email, password_hash, created_at FROM users WHERE email = ?`, email)
	var u entities.User
	if err := row.Scan(&u.ID, &u.Email, &u.PasswordHash, &u.CreatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) { // Jika tidak ada baris ditemukan
			return nil, nil // Mengembalikan pointer nil dan error nil. INI PENTING!
		}
		return nil, err // Mengembalikan pointer nil dan error database
	}
	return &u, nil
}

func (r *MySQLUserRepo) FindByID(ctx context.Context, id string) (*entities.User, error) {
	row := r.db.QueryRowContext(ctx, `SELECT id, email, password_hash, created_at FROM users WHERE id = ?`, id)
	var u entities.User
	if err := row.Scan(&u.ID, &u.Email, &u.PasswordHash, &u.CreatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &u, nil
}

func (r *MySQLUserRepo) CheckPassword(ctx context.Context, userID string, password string) (bool, error) {
	var storedHash string
	log.Printf("UserID: %s", userID)
	err := r.db.QueryRowContext(ctx, `SELECT password_hash FROM users WHERE id = ?`, userID).Scan(&storedHash)
	if err != nil {
		return false, err
	}

	log.Printf("yuuhuuhuhuhbfkdkfksd: %s", userID)
	// Menggunakan bcrypt untuk membandingkan hash password
	err = bcrypt.CompareHashAndPassword([]byte(storedHash), []byte(password))
	if err != nil {
		log.Printf("Password verification failed for userID: %s, error: %v", userID, err)
		return false, err // Jika hash password tidak cocok
	}
	log.Printf("Password match for userID: %s", userID)
	return true, nil // Password cocok
}
