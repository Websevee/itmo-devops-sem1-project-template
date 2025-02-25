#!/bin/bash

# # Установка зависимостей
# sudo apt-get update
# sudo apt-get install -y postgresql postgresql-contrib
# sudo apt-get install -y golang

# Настройка PostgreSQL
sudo -u postgres psql -c "CREATE USER validator WITH PASSWORD 'val1dat0r';"
sudo -u postgres psql -c "CREATE DATABASE \"project-sem-1\" OWNER validator;"

# sudo -u postgres psql -c "CREATE USER validator WITH PASSWORD 'val1dat0r';"
# sudo -u postgres psql -c "CREATE DATABASE \"project-sem-1\" OWNER validator;"
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




# #!/bin/bash

# sudo apt-get update

# # Установка Golang
# echo "Установка Golang..."
# sudo rm go1.24.0.linux-amd64.tar.gz
# wget https://go.dev/dl/go1.24.0.linux-amd64.tar.gz
# sudo rm -rf /usr/local/go && sudo tar -C /usr/local -xzf go1.24.0.linux-amd64.tar.gz
# sudo rm go1.24.0.linux-amd64.tar.gz
# export PATH=$PATH:/usr/local/go/bin
# go version

# # Установка PostgreSQL
# echo "Установка PostgreSQL..."
# sudo apt-get install -y postgresql postgresql-contrib

# # Запуск PostgreSQL
# sudo service postgresql start

# # Создание пользователя и базы данных
# sudo -u postgres psql -c "CREATE USER validator WITH PASSWORD 'val1dat0r';"
# sudo -u postgres psql -c "CREATE DATABASE \"project-sem-1\" WITH OWNER validator;"

# # Применение миграций
# sudo -u postgres psql -d project-sem-1 -c "
# CREATE TABLE prices (
#     id SERIAL PRIMARY KEY,
#     created_at DATE,
#     name TEXT,
#     category TEXT,
#     price NUMERIC
# );"