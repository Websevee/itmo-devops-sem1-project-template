package service

import (
	"archive/zip"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"itmo-devops-fp1/internal/repository"
	"itmo-devops-fp1/internal/types"
	"mime/multipart"
	"net/http"
	"os"
	"strconv"
)

// Обрабатывает загрузку данных из архива
func ProcessUpload(r *http.Request, archiveType types.ArchiveType) (types.GetPricesResponse, error) {
	file, err := getUploadedFile(r)
	if err != nil {
		return types.GetPricesResponse{}, err
	}
	defer file.Close()

	tempFile, err := saveTempFile(file, archiveType)
	if err != nil {
		return types.GetPricesResponse{}, err
	}
	defer os.Remove(tempFile.Name())

	return processArchive(tempFile.Name(), archiveType)
}

// Получает загруженный файл из запроса
func getUploadedFile(r *http.Request) (multipart.File, error) {
	file, _, err := r.FormFile("file")
	if err != nil {
		return nil, errors.New("не удалось прочитать файл")
	}
	return file, nil
}

// Сохраняет временный файл
func saveTempFile(file multipart.File, archiveType types.ArchiveType) (*os.File, error) {
	ext := ".zip"
	if archiveType == types.Tar {
		ext = ".tar"
	}

	tempFile, err := os.CreateTemp("", "upload-*"+ext)
	if err != nil {
		return nil, errors.New("не удалось создать временный файл")
	}

	if _, err := io.Copy(tempFile, file); err != nil {
		return nil, errors.New("не удалось сохранить файл")
	}

	return tempFile, nil
}

// Обрабатывает архив в зависимости от типа
func processArchive(filename string, archiveType types.ArchiveType) (types.GetPricesResponse, error) {
	if archiveType == types.Tar {
		return repository.ProcessTar(filename)
	}
	return repository.ProcessZip(filename)
}

// Обрабатывает скачивание данных
func ProcessDownload(w http.ResponseWriter, r *http.Request) error {
	products, err := fetchProducts()
	if err != nil {
		return err
	}

	tempCsvFile, err := createTempCSV(products)
	if err != nil {
		return err
	}
	defer cleanupFile(tempCsvFile)

	zipFile, err := createZipFromCSV(tempCsvFile)
	if err != nil {
		return err
	}
	defer cleanupFile(zipFile)

	return serveZipFile(w, r, zipFile)
}

// Очищает временный файл
func cleanupFile(file *os.File) {
	file.Close()
	os.Remove(file.Name())
}

// Получает данные из репозитория
func fetchProducts() ([]types.Product, error) {
	products, err := repository.FetchData()
	if err != nil {
		return nil, fmt.Errorf("не удалось получить продукты: %w", err)
	}
	return products, nil
}

// Создает временный CSV-файл с данными
func createTempCSV(products []types.Product) (*os.File, error) {
	tempCsvFile, err := os.CreateTemp("", "data-*.csv")
	if err != nil {
		return nil, fmt.Errorf("не удалось создать временный файл: %w", err)
	}

	if err := writeProductsToCSV(tempCsvFile, products); err != nil {
		return nil, err
	}

	return tempCsvFile, nil
}

// Записывает продукты в CSV
func writeProductsToCSV(file *os.File, products []types.Product) error {
	writer := csv.NewWriter(file)
	defer writer.Flush()

	for _, product := range products {
		record := []string{
			strconv.Itoa(product.ProductId),
			product.Name,
			product.Category,
			strconv.FormatFloat(product.Price, 'f', 2, 64),
			product.CreatedAt,
		}
		if err := writer.Write(record); err != nil {
			return fmt.Errorf("не удалось записать в CSV: %w", err)
		}
	}
	return nil
}

// Создает ZIP-архив из CSV-файла
func createZipFromCSV(csvFile *os.File) (*os.File, error) {
	zipFile, err := os.Create("data.zip")
	if err != nil {
		return nil, fmt.Errorf("не удалось создать ZIP файл: %w", err)
	}

	if err := addFileToZip(zipFile, csvFile); err != nil {
		return nil, err
	}

	return zipFile, nil
}

// Добавляет файл в ZIP-архив
func addFileToZip(zipFile *os.File, fileToAdd *os.File) error {
	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	fileReader, err := os.Open(fileToAdd.Name())
	if err != nil {
		return fmt.Errorf("не удалось открыть файл: %w", err)
	}
	defer fileReader.Close()

	zipEntry, err := zipWriter.Create("data.csv")
	if err != nil {
		return fmt.Errorf("не удалось создать запись в ZIP: %w", err)
	}

	if _, err := io.Copy(zipEntry, fileReader); err != nil {
		return fmt.Errorf("не удалось записать в ZIP: %w", err)
	}

	return nil
}

// Отправляет ZIP-архив клиенту
func serveZipFile(w http.ResponseWriter, r *http.Request, zipFile *os.File) error {
	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", "attachment; filename=data.zip")
	http.ServeFile(w, r, zipFile.Name())
	return nil
}
