package config

import (
	"database/sql"
	"log"

	_ "github.com/lib/pq"
)

func ConnectDB() *sql.DB {
	connStr := "postgres://postgres:Ana4aBada$$@localhost:5432/nanny_db?sslmode=disable"

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal("Ошибка подключения к БД:", err)
	}

	if err := db.Ping(); err != nil {
		log.Fatal("БД недоступна:", err)
	}

	log.Println("✅ Подключено к PostgreSQL")
	return db
}
