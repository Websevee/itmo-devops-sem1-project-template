package main

import (
	"itmo-devops-fp1/internal/handler" // Убедитесь, что путь правильный
	"log"
	"net/http"
)

func main() {
	http.HandleFunc("/api/v0/prices", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			handler.UploadHandler(w, r)
		case http.MethodGet:
			handler.DownloadHandler(w, r)
		default:
			http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		}
	})

	log.Println("Server started on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
