package handler

import (
	"encoding/json"
	"itmo-devops-fp1/internal/service"
	"itmo-devops-fp1/internal/types"
	"net/http"
)

// POST-запрос для загрузки данных
func UploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	// Получаем тип архива из параметра запроса
	archiveType := r.URL.Query().Get("type")
	if archiveType == "" {
		archiveType = "zip" // По умолчанию zip
	}

	response, err := service.ProcessUpload(r, types.ArchiveType(archiveType))

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GET-запрос для скачивания данных
func DownloadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	// Проверяем наличие параметров фильтрации
	if r.URL.Query().Has("start") {
		err := service.ProcessFilteredDownload(w, r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		return
	}

	// Если параметров нет, возвращаем все данные
	err := service.ProcessDownload(w, r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
