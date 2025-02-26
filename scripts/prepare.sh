#!/bin/bash

# Установка Go-зависимостей
go mod tidy

# Создание пользователя и базы данных
echo "1"

# ! PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -U $DB_USER -d $DB_NAME

export PGPASSWORD=val1dat0r
psql -h localhost -p 5432 -U validator -c "CREATE DATABASE project-sem-1 OWNER validator;"

echo "2"

psql "postgresql://validator:val1dat0r@localhost:5432/project-sem-1" -c "
CREATE TABLE IF NOT EXISTS prices (
    id SERIAL PRIMARY KEY,
    created_at DATE,
    name TEXT,
    category TEXT,
    price NUMERIC
);"
