package database

import (
	"fmt"
	"log"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func MustRunMigrations(dsn, dir string) {
	// dsn format: root:@tcp(127.0.0.1:3306)/mydb
	// migrate butuh format DSN khusus: "mysql://user:pass@tcp(host:port)/dbname"
	url := fmt.Sprintf("mysql://%s", dsn)

	m, err := migrate.New(
		fmt.Sprintf("file://%s", dir),
		url,
	)
	if err != nil {
		log.Fatalf("Gagal inisialisasi migrasi: %v", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		log.Fatalf("Migrasi gagal: %v", err)
	}

	log.Println("Migrasi database berhasil dijalankan")
}
