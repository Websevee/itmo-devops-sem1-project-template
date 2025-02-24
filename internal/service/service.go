package service

import (
	"archive/zip"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"itmo-devops-fp1/internal/repository"
	"itmo-devops-fp1/internal/types"
	"net/http"
	"os"
	"strconv"
)

// обрабатывает загрузку данных из zip-архива
func ProcessUpload(r *http.Request) (types.GetPricesResponse, error) {
	file, _, err := r.FormFile("file")
	if err != nil {
		return types.GetPricesResponse{}, errors.New("unable to read file")
	}
	defer file.Close()

	// Сохранение временного файла
	tempFile, err := os.CreateTemp("", "upload-*.zip")
	if err != nil {
		return types.GetPricesResponse{}, errors.New("unable to create temp file")
	}
	defer os.Remove(tempFile.Name())

	_, err = io.Copy(tempFile, file)
	if err != nil {
		return types.GetPricesResponse{}, errors.New("unable to save file")
	}

	// Обработка zip-архива
	return repository.ProcessZip(tempFile.Name())
}

func ProcessDownload(w http.ResponseWriter, r *http.Request) error {
	// Получение данных
	products, err := fetchProducts()
	if err != nil {
		return err
	}

	// Создание временного CSV-файла
	tempCsvFile, err := createTempCSV(products)
	if err != nil {
		return err
	}
	defer os.Remove(tempCsvFile.Name())
	defer tempCsvFile.Close()

	// Создание zip-архива
	zipFile, err := createZipFromCSV(tempCsvFile)
	if err != nil {
		return err
	}
	defer os.Remove(zipFile.Name())

	// Возврат zip-архива клиенту
	return serveZipFile(w, r, zipFile)
}

// получает данные из репозитория
func fetchProducts() ([]types.Product, error) {
	products, err := repository.FetchData()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch products: %w", err)
	}
	return products, nil
}

// создает временный CSV-файл и записывает в него данные
func createTempCSV(products []types.Product) (*os.File, error) {
	tempCsvFile, err := os.CreateTemp("", "data-*.csv")
	if err != nil {
		return nil, fmt.Errorf("unable to create temp file: %w", err)
	}

	csvWriter := csv.NewWriter(tempCsvFile)
	defer csvWriter.Flush()

	for _, product := range products {
		err := csvWriter.Write([]string{
			strconv.Itoa(product.Id),
			product.CreatedAt,
			product.Name,
			product.Category,
			strconv.FormatFloat(product.Price, 'f', 2, 64),
		})
		if err != nil {
			return nil, fmt.Errorf("unable to write to CSV: %w", err)
		}
	}

	return tempCsvFile, nil
}

// создает zip-архив и добавляет в него CSV-файл
func createZipFromCSV(tempCsvFile *os.File) (*os.File, error) {
	zipFile, err := os.Create("data.zip")
	if err != nil {
		return nil, fmt.Errorf("unable to create zip file: %w", err)
	}

	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	csvFile, err := os.Open(tempCsvFile.Name())
	if err != nil {
		return nil, fmt.Errorf("unable to open CSV file: %w", err)
	}
	defer csvFile.Close()

	csvInZip, err := zipWriter.Create("data.csv")
	if err != nil {
		return nil, fmt.Errorf("unable to create file in zip: %w", err)
	}

	_, err = io.Copy(csvInZip, csvFile)
	if err != nil {
		return nil, fmt.Errorf("unable to write CSV to zip: %w", err)
	}

	return zipFile, nil
}

// отправляет zip-архив клиенту
func serveZipFile(w http.ResponseWriter, r *http.Request, zipFile *os.File) error {
	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", "attachment; filename=data.zip")
	http.ServeFile(w, r, zipFile.Name())
	return nil
}
