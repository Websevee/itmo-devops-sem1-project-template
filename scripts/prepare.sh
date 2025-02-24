#!/bin/bash

# Установка PostgreSQL
sudo apt-get update
sudo apt-get install -y postgresql postgresql-contrib

# Запуск PostgreSQL
sudo service postgresql start

# Создание пользователя и базы данных
sudo -u postgres psql -c "CREATE USER validator WITH PASSWORD 'val1dat0r';"
sudo -u postgres psql -c "CREATE DATABASE \"project-sem-1\" WITH OWNER validator;"

# Применение миграций
sudo -u postgres psql -d project-sem-1 -c "
CREATE TABLE prices (
    id SERIAL PRIMARY KEY,
    created_at DATE,
    name TEXT,
    category TEXT,
    price NUMERIC
);"