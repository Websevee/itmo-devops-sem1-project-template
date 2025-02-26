#!/bin/bash

# Установка Go-зависимостей
go mod tidy

# Создание таблицы
PGPASSWORD=val1dat0r psql -h localhost -p 5432 -U validator -d project-sem-1 -c "
CREATE TABLE IF NOT EXISTS prices (
    id SERIAL PRIMARY KEY,
    product_id INTEGER,
    created_at DATE,
    name TEXT,
    category TEXT,
    price NUMERIC
);"
