package utils

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
)

// Конфигурация подключения к базе данных
type DBConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
}

// Получает конфигурацию из переменных окружения
func getDBConfig() DBConfig {
	return DBConfig{
		Host:     getEnvOrDefault("POSTGRES_HOST", "localhost"),
		Port:     getEnvOrDefault("POSTGRES_PORT", "5432"),
		User:     getEnvOrDefault("POSTGRES_USER", "validator"),
		Password: getEnvOrDefault("POSTGRES_PASSWORD", "val1dat0r"),
		DBName:   getEnvOrDefault("POSTGRES_DB", "project-sem-1"),
	}
}

// Получает значение переменной окружения или возвращает значение по умолчанию
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// Создает строку подключения из конфигурации
func buildConnectionString(config DBConfig) string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		config.Host, config.Port, config.User, config.Password, config.DBName,
	)
}

// Подключается к базе данных
func ConnectDB() *sql.DB {
	config := getDBConfig()
	connStr := buildConnectionString(config)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("Не удалось подключиться к базе данных: %v", err)
	}

	if err = db.Ping(); err != nil {
		db.Close()
		log.Fatalf("Не удалось проверить подключение к базе данных: %v", err)
	}

	log.Println("Успешное подключение к базе данных")
	return db
}

// Закрывает соединение с базой данных
func CloseDB(db *sql.DB) {
	if db != nil {
		db.Close()
	}
}
