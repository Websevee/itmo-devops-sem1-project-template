package utils

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
)

func ConnectDB() *sql.DB {
	log.Println("HOST: ")
	log.Println("HOST: " + os.Getenv("POSTGRES_HOST"))
	fmt.Println("HOST: " + os.Getenv("POSTGRES_HOST"))

	db, err := sql.Open("postgres", fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		"locahost",
		"5432",
		"validator",
		"val1dat0r",
		"project-sem-1",
		// os.Getenv("POSTGRES_HOST"),
		// os.Getenv("POSTGRES_PORT"),
		// os.Getenv("POSTGRES_USER"),
		// os.Getenv("POSTGRES_PASSWORD"),
		// os.Getenv("POSTGRES_DB"),
	))
	if err != nil {
		log.Fatalf("Unable to connect to database: %v", err)
	}
	defer db.Close()

	fmt.Println("Connected to the database!")

	return db
}
