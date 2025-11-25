package database

import (
	"context"
	"database/sql"
)

// Tx wrapper untuk transaction dengan cleanup otomatis
type Tx struct {
	*sql.Tx
	rolled bool
}

// Commit commits transaction
func (t *Tx) Commit() error {
	if t.rolled {
		return nil
	}
	return t.Tx.Commit()
}

// Rollback rollbacks transaction
func (t *Tx) Rollback() error {
	t.rolled = true
	return t.Tx.Rollback()
}

// ExecWithTx execute function dengan automatic transaction handling
func ExecWithTx(ctx context.Context, db *sql.DB, fn func(tx *sql.Tx) error) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	if err := fn(tx); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

// ExecWithTxIsolation execute dengan isolation level tertentu
func ExecWithTxIsolation(ctx context.Context, db *sql.DB, level sql.IsolationLevel, fn func(tx *sql.Tx) error) error {
	tx, err := db.BeginTx(ctx, &sql.TxOptions{Isolation: level})
	if err != nil {
		return err
	}

	if err := fn(tx); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}
