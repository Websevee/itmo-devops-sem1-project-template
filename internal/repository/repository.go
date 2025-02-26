package repository

import (
	"archive/tar"
	"archive/zip"
	"database/sql"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"itmo-devops-fp1/internal/types"
	"itmo-devops-fp1/pkg/utils"
	"os"
	"strconv"
	"strings"
)

var db *sql.DB

func init() {
	db = utils.ConnectDB()
}

// Обрабатывает ZIP-архив
func ProcessZip(filename string) (types.GetPricesResponse, error) {
	reader, err := zip.OpenReader(filename)
	if err != nil {
		return types.GetPricesResponse{}, fmt.Errorf("ошибка открытия ZIP: %w", err)
	}
	defer reader.Close()

	var csvFile *zip.File
	for _, file := range reader.File {
		if strings.HasSuffix(file.Name, ".csv") {
			csvFile = file
			break
		}
	}

	if csvFile == nil {
		return types.GetPricesResponse{}, errors.New("CSV файл не найден в архиве")
	}

	// Создаем временный файл для CSV
	tempFile, err := os.CreateTemp("", "*.csv")
	if err != nil {
		return types.GetPricesResponse{}, fmt.Errorf("ошибка создания временного файла: %w", err)
	}
	defer os.Remove(tempFile.Name())
	defer tempFile.Close()

	// Копируем содержимое из архива во временный файл
	rc, err := csvFile.Open()
	if err != nil {
		return types.GetPricesResponse{}, fmt.Errorf("ошибка открытия CSV: %w", err)
	}
	defer rc.Close()

	if _, err := io.Copy(tempFile, rc); err != nil {
		return types.GetPricesResponse{}, fmt.Errorf("ошибка копирования данных: %w", err)
	}

	// Используем локальную функцию вместо service.ProcessCSVFile
	return ProcessCSVFile(tempFile.Name())
}

// Обрабатывает tar-архив
func ProcessTar(filename string) (types.GetPricesResponse, error) {
	file, err := os.Open(filename)
	if err != nil {
		return types.GetPricesResponse{}, fmt.Errorf("ошибка открытия TAR: %w", err)
	}
	defer file.Close()

	tr := tar.NewReader(file)
	var csvFound bool

	// Создаем временный файл для CSV
	tempFile, err := os.CreateTemp("", "*.csv")
	if err != nil {
		return types.GetPricesResponse{}, fmt.Errorf("ошибка создания временного файла: %w", err)
	}
	defer os.Remove(tempFile.Name())
	defer tempFile.Close()

	// Ищем CSV файл в архиве
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return types.GetPricesResponse{}, fmt.Errorf("ошибка чтения TAR: %w", err)
		}

		if strings.HasSuffix(header.Name, ".csv") {
			if _, err := io.Copy(tempFile, tr); err != nil {
				return types.GetPricesResponse{}, fmt.Errorf("ошибка копирования данных: %w", err)
			}
			csvFound = true
			break
		}
	}

	if !csvFound {
		return types.GetPricesResponse{}, errors.New("CSV файл не найден в архиве")
	}

	// Используем общую логику обработки CSV
	return ProcessCSVFile(tempFile.Name())
}

// Извлекает данные из базы данных
func FetchData() ([]types.Product, error) {
	rows, err := db.Query("SELECT id, product_id, created_at, name, category, price FROM prices")
	if err != nil {
		return nil, errors.New("не удалось выполнить запрос к базе данных")
	}
	defer rows.Close()

	var products []types.Product
	for rows.Next() {
		var product types.Product
		if err := rows.Scan(&product.Id, &product.ProductId, &product.CreatedAt, &product.Name, &product.Category, &product.Price); err != nil {
			return nil, errors.New("ошибка чтения данных")
		}
		products = append(products, product)
	}

	return products, nil
}

// Возвращает статистику по загруженным данным
func GetStatistics(records [][]string) (types.GetPricesResponse, error) {
	var response types.GetPricesResponse

	// Общее количество строк (исключая заголовок)
	response.TotalCount = len(records) - 1

	// Подсчет дубликатов во входных данных
	inputDuplicates := make(map[string]int)
	for i := 1; i < len(records); i++ { // Пропускаем заголовок
		key := records[i][0] // ProductId как ключ
		inputDuplicates[key]++
	}

	// Подсчитываем дубликаты во входных данных
	inputDupsCount := 0
	for _, count := range inputDuplicates {
		if count > 1 {
			inputDupsCount += count - 1
		}
	}

	// Подсчет дубликатов в БД
	var dbDupsCount int
	err := db.QueryRow(`
		SELECT COUNT(*) - COUNT(DISTINCT product_id) 
		FROM prices
	`).Scan(&dbDupsCount)
	if err != nil {
		return response, fmt.Errorf("ошибка подсчета дубликатов в БД: %w", err)
	}

	// Общее количество дубликатов
	response.DuplicatesCount = inputDupsCount + dbDupsCount

	// Остальные подсчеты
	err = db.QueryRow(`
		SELECT 
			COUNT(*) as total_items,
			COUNT(DISTINCT category) as total_categories,
			COALESCE(SUM(price), 0) as total_price
		FROM prices
	`).Scan(&response.TotalItems, &response.TotalCategories, &response.TotalPrice)
	if err != nil {
		return response, fmt.Errorf("ошибка подсчета статистики: %w", err)
	}

	return response, nil
}

// Преобразует CSV-строку в структуру Product
func MapRecordToProduct(record []string) (types.Product, error) {
	productId, err := strconv.Atoi(record[0])
	if err != nil {
		return types.Product{}, errors.New("неверный формат ProductId")
	}

	price, err := strconv.ParseFloat(record[3], 64)
	if err != nil {
		return types.Product{}, errors.New("неверный формат цены")
	}

	return types.Product{
		ProductId: productId,
		CreatedAt: record[4],
		Name:      record[1],
		Category:  record[2],
		Price:     price,
	}, nil
}

// Вставляет данные о продукте в базу данных
func InsertProductIntoDB(product types.Product) error {
	_, err := db.Exec("INSERT INTO prices (product_id, created_at, name, category, price) VALUES ($1, $2, $3, $4, $5)",
		product.ProductId, product.CreatedAt, product.Name, product.Category, product.Price)
	return err
}

// Получает отфильтрованные данные из БД
func FetchFilteredData(start, end string, min, max float64) ([]types.Product, error) {
	query := `
		SELECT id, product_id, created_at, name, category, price 
		FROM prices 
		WHERE created_at >= $1 
		AND created_at <= $2 
		AND price >= $3 
		AND price <= $4
	`

	rows, err := db.Query(query, start, end, min, max)
	if err != nil {
		return nil, fmt.Errorf("ошибка выполнения запроса: %w", err)
	}
	defer rows.Close()

	var products []types.Product
	for rows.Next() {
		var product types.Product
		if err := rows.Scan(
			&product.Id,
			&product.ProductId,
			&product.CreatedAt,
			&product.Name,
			&product.Category,
			&product.Price,
		); err != nil {
			return nil, fmt.Errorf("ошибка сканирования данных: %w", err)
		}
		products = append(products, product)
	}

	return products, nil
}

// Обрабатывает CSV файл и возвращает статистику
func ProcessCSVFile(filename string) (types.GetPricesResponse, error) {
	file, err := os.Open(filename)
	if err != nil {
		return types.GetPricesResponse{}, fmt.Errorf("не удалось открыть файл: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return types.GetPricesResponse{}, fmt.Errorf("ошибка чтения CSV: %w", err)
	}

	// Пропускаем заголовки при обработке
	for i := 1; i < len(records); i++ {
		product, err := MapRecordToProduct(records[i])
		if err != nil {
			return types.GetPricesResponse{}, fmt.Errorf("ошибка обработки записи %d: %w", i, err)
		}

		if err := InsertProductIntoDB(product); err != nil {
			return types.GetPricesResponse{}, fmt.Errorf("ошибка вставки в БД: %w", err)
		}
	}

	return GetStatistics(records)
}

// Возвращает количество дубликатов в БД
func GetDuplicatesCount() (int, error) {
	var count int
	err := db.QueryRow(`
		SELECT COUNT(*) - COUNT(DISTINCT product_id) 
		FROM prices
	`).Scan(&count)
	return count, err
}

// Возвращает статистику из БД
func GetDBStats() (int, int, float64, error) {
	var totalItems, totalCategories int
	var totalPrice float64

	err := db.QueryRow(`
		SELECT 
			COUNT(*) as total_items,
			COUNT(DISTINCT category) as total_categories,
			COALESCE(SUM(price), 0) as total_price
		FROM prices
	`).Scan(&totalItems, &totalCategories, &totalPrice)

	return totalItems, totalCategories, totalPrice, err
}
