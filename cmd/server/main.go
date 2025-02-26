package main

import (
	"fmt"
	"itmo-devops-fp1/internal/handler"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func main() {
	fmt.Println("Server is starting...")
	log.Println("Server is starting...")

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
	if err := http.ListenAndServe(":8080", r); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}

	// go func() {
	// 	log.Println("Server started on :8080")
	// 	if err := http.ListenAndServe(":8080", r); err != nil {
	// 		log.Fatalf("Server failed to start: %v", err)
	// 	}
	// }()

	// Основной поток завершается, но сервер продолжает работать
	fmt.Println("Server is running...")
	log.Println("Server is running...")
}
