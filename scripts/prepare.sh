#!/bin/bash

# Настройка PostgreSQL
until pg_isready -h localhost -p 5432 -U validator; do
    sleep 1
done
echo "PostgreSQL доступен."

# Создание пользователя и базы данных
echo "1"

psql -U validator -c "CREATE DATABASE \"project-sem-1\" OWNER validator;"

echo "2"

psql "postgresql://validator:val1dat0r@localhost:5432/project-sem-1" -c "
CREATE TABLE IF NOT EXISTS prices (
    id SERIAL PRIMARY KEY,
    created_at DATE,
    name TEXT,
    category TEXT,
    price NUMERIC
);"

# Установка Go-зависимостей
go mod tidy
