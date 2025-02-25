#!/bin/bash

# Настройка PostgreSQL
until pg_isready -h localhost -p 5432 -U validator; do
  sleep 1
done
echo "PostgreSQL доступен."

# Создание пользователя и базы данных
sudo -u postgres psql -h localhost -p 5432 -c "CREATE USER validator WITH PASSWORD 'val1dat0r';"
sudo -u postgres psql -h localhost -p 5432 -c "CREATE DATABASE \"project-sem-1\" OWNER validator;"

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
