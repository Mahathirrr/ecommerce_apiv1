package db

import (
	"fmt"

	"github.com/jmoiron/sqlx"
)

func GetConnection() (*sqlx.DB, error) {
	db, err := sqlx.Open("mysql", "dev:dev@tcp(localhost:3306)/go_kasus_4?parseTime=true")
	if err != nil {
		return nil, fmt.Errorf("Error opening database: %w", err)
	}
	return db, nil
}
