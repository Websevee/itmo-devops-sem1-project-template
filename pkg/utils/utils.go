package utils

import (
	"database/sql"

	_ "github.com/lib/pq"
)

func ConnectDB() *sql.DB {
	// connStr := "user=postgres password=pass dbname=project-sem-1 sslmode=disable"
	connStr := "user=validator password=val1dat0r dbname=project-sem-1 sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		panic("Unable to connect to database: " + err.Error())
	}
	return db
}
