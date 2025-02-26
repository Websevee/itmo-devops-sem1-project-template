#!/bin/bash

go run cmd/server/main.go &

# Сохраняем PID процесса сервера
SERVER_PID=$!

while ! nc -z localhost 8080; do
    sleep 1
done

# Выводим PID для информации
echo "Сервер запущен с PID: $SERVER_PID"
