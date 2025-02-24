package main

import (
	"archive/zip"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
)

func uploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	file, _, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Unable to read file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	reader, err := zip.NewReader(file, r.ContentLength)
	if err != nil {
		http.Error(w, "Unable to read zip file", http.StatusInternalServerError)
		return
	}

	var totalItems int
	var totalCategories = make(map[string]bool)
	var totalPrice float64

	for _, f := range reader.File {
		nameParts := strings.Split(f.Name, "/")
		fileName := nameParts[len(nameParts)-1]

		if fileName == "data.csv" {
			rc, err := f.Open()
			if err != nil {
				http.Error(w, "Unable to open csv file", http.StatusInternalServerError)
				return
			}
			defer rc.Close()

			csvReader := csv.NewReader(rc)
			isHeader := true
			for {
				record, err := csvReader.Read()

				if isHeader {
					isHeader = false
					continue
				}

				if err == io.EOF {
					break
				}
				if err != nil {
					http.Error(w, "Unable to read csv file", http.StatusInternalServerError)
					return
				}

				price, err := strconv.ParseFloat(record[3], 64)
				if err != nil {
					http.Error(w, "Invalid price format", http.StatusInternalServerError)
					return
				}

				_, err = db.Exec("INSERT INTO prices (id, create_date, name, category, price) VALUES ($1, $2, $3, $4, $5)",
					record[0], record[4], record[1], record[2], price)
				if err != nil {
					http.Error(w, "Unable to insert data into database: "+err.Error(), http.StatusInternalServerError)
					return
				}

				totalItems++
				totalCategories[record[2]] = true
				totalPrice += price
			}
		}
	}

	response := Response{
		TotalItems:      totalItems,
		TotalCategories: len(totalCategories),
		TotalPrice:      totalPrice,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func downloadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	rows, err := db.Query("SELECT id, create_date, name, category, price FROM prices")
	if err != nil {
		http.Error(w, "Unable to fetch data from database", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	// Создаем временный CSV-файл
	file, err := os.CreateTemp("", "data-*.csv")
	if err != nil {
		http.Error(w, "Unable to create csv file", http.StatusInternalServerError)
		return
	}
	defer os.Remove(file.Name()) // Удаляем временный файл после использования
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Записываем заголовок
	header := []string{"id", "create_date", "name", "category", "price"}
	if err := writer.Write(header); err != nil {
		http.Error(w, "Unable to write header to CSV", http.StatusInternalServerError)
		return
	}

	for rows.Next() {
		var id int
		var create_date, name, category string
		var price float64
		err = rows.Scan(&id, &create_date, &name, &category, &price)
		if err != nil {
			http.Error(w, "Unable to scan row", http.StatusInternalServerError)
			return
		}
		writer.Write([]string{strconv.Itoa(id), create_date, name, category, fmt.Sprintf("%.2f", price)})
	}

	if err := writer.Error(); err != nil {
		http.Error(w, "Error writing CSV data", http.StatusInternalServerError)
		return
	}

	// Создаем временный ZIP-файл
	zipFile, err := os.CreateTemp("", "data-*.zip")
	if err != nil {
		http.Error(w, "Unable to create zip file", http.StatusInternalServerError)
		return
	}
	defer os.Remove(zipFile.Name()) // Удаляем временный файл после использования
	defer zipFile.Close()

	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	dataFile, err := zipWriter.Create("data.csv")
	if err != nil {
		http.Error(w, "Unable to create file in zip archive", http.StatusInternalServerError)
		return
	}

	// Перемещаем указатель файла CSV в начало
	file.Seek(0, 0)
	if _, err := io.Copy(dataFile, file); err != nil {
		http.Error(w, "Unable to copy data to zip file", http.StatusInternalServerError)
		return
	}

	// Отправляем ZIP-файл клиенту
	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", "attachment; filename=data.zip")
	http.ServeFile(w, r, zipFile.Name())
}
