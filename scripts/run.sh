#!/bin/bash

go run cmd/server/main.go &

sleep 5

# Сохраняем PID процесса сервера
SERVER_PID=$!

# Выводим PID для информации
echo "Сервер запущен с PID: $SERVER_PID"

# Завершаем скрипт, но оставляем сервер работать
exit 0
