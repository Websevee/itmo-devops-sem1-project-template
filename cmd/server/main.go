package main

import (
	"itmo-devops-fp1/internal/handler"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func main() {
	// Создаем новый роутер
	r := chi.NewRouter()

	// Добавляем middleware (опционально)
	r.Use(middleware.Logger) // Логирование всех запросов

	// Регистрируем маршруты
	r.Route("/api/v0", func(r chi.Router) {
		r.Post("/prices", handler.UploadHandler)  // POST /api/v0/prices
		r.Get("/prices", handler.DownloadHandler) // GET /api/v0/prices
	})

	log.Println("Server started on :8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}
