#!/bin/bash

# Настройка PostgreSQL
sudo -u postgres psql -c "CREATE USER $POSTGRES_USER WITH PASSWORD '$POSTGRES_PASSWORD';"
sudo -u postgres psql -c "CREATE DATABASE \"$POSTGRES_DB\" OWNER $POSTGRES_USER;"
psql "postgresql://$POSTGRES_USER:$POSTGRES_PASSWORD@$POSTGRES_HOST:$POSTGRES_PORT/$POSTGRES_DB" -c "
CREATE TABLE IF NOT EXISTS prices (
    id SERIAL PRIMARY KEY,
    created_at DATE,
    name TEXT,
    category TEXT,
    price NUMERIC
);"

# Установка Go-зависимостей
go mod tidy
