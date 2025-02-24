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

	tempCsvFile, err := os.CreateTemp("", "data-*.csv")
	if err != nil {
		http.Error(w, "Unable to create CSV file", http.StatusInternalServerError)
		return
	}
	defer os.Remove(tempCsvFile.Name())
	defer tempCsvFile.Close()

	csvWriter := csv.NewWriter(tempCsvFile)

	// Write the CSV header
	header := []string{"id", "create_date", "name", "category", "price"}
	if err := csvWriter.Write(header); err != nil {
		http.Error(w, "Unable to write header to CSV", http.StatusInternalServerError)
		return
	}

	// Write data to CSV
	for rows.Next() {
		var id int
		var createDate, name, category string
		var price float64
		if err := rows.Scan(&id, &createDate, &name, &category, &price); err != nil {
			http.Error(w, "Unable to scan row", http.StatusInternalServerError)
			return
		}
		if err := csvWriter.Write([]string{strconv.Itoa(id), createDate, name, category, fmt.Sprintf("%.2f", price)}); err != nil {
			http.Error(w, "Unable to write row to CSV", http.StatusInternalServerError)
			return
		}
	}

	// Check for errors during writing
	if err := csvWriter.Error(); err != nil {
		http.Error(w, "Error writing CSV data", http.StatusInternalServerError)
		return
	}

	// Flush the CSV writer to ensure all data is written to the file
	csvWriter.Flush()
	if err := csvWriter.Error(); err != nil {
		http.Error(w, "Error flushing CSV data", http.StatusInternalServerError)
		return
	}

	// Reset the file pointer to the beginning
	if _, err := tempCsvFile.Seek(0, 0); err != nil {
		http.Error(w, "Unable to reset file pointer", http.StatusInternalServerError)
		return
	}

	// Create a temporary ZIP file
	zipFile, err := os.CreateTemp("", "data-*.zip")
	if err != nil {
		http.Error(w, "Unable to create zip file", http.StatusInternalServerError)
		return
	}
	defer os.Remove(zipFile.Name()) // Remove the temporary file after use
	defer zipFile.Close()

	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	// Create a file inside the ZIP archive
	dataFile, err := zipWriter.Create("data.csv")
	if err != nil {
		http.Error(w, "Unable to create file in zip archive", http.StatusInternalServerError)
		return
	}

	// Copy data from the CSV file to the ZIP archive
	if _, err := io.Copy(dataFile, tempCsvFile); err != nil {
		http.Error(w, "Unable to copy data to zip file", http.StatusInternalServerError)
		return
	}

	// Ensure all data is written to the zip file
	if err := zipWriter.Close(); err != nil {
		http.Error(w, "Unable to close zip writer", http.StatusInternalServerError)
		return
	}

	// Send the ZIP file to the client
	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", "attachment; filename=data.zip")
	http.ServeFile(w, r, zipFile.Name())
}
