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

	// Создаем итоговый файл для CSV
	resultFile, err := os.Create("result.csv")
	if err != nil {
		return types.GetPricesResponse{}, fmt.Errorf("ошибка создания файла: %w", err)
	}
	defer os.Remove(resultFile.Name())
	defer resultFile.Close()

	// Копируем содержимое из архива в файл
	rc, err := csvFile.Open()
	if err != nil {
		return types.GetPricesResponse{}, fmt.Errorf("ошибка открытия CSV: %w", err)
	}
	defer rc.Close()

	if _, err := io.Copy(resultFile, rc); err != nil {
		return types.GetPricesResponse{}, fmt.Errorf("ошибка копирования данных: %w", err)
	}

	return ProcessCSVFile(resultFile.Name())
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

	// Создаем итоговый файл для CSV
	resultFile, err := os.Create("result.csv")
	if err != nil {
		return types.GetPricesResponse{}, fmt.Errorf("ошибка создания файла: %w", err)
	}
	defer os.Remove(resultFile.Name())
	defer resultFile.Close()

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
			if _, err := io.Copy(resultFile, tr); err != nil {
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
	return ProcessCSVFile(resultFile.Name())
}

// Извлекает данные из базы данных
func FetchData() ([]types.Product, error) {
	rows, err := db.Query("SELECT id, created_at, name, category, price FROM prices")
	if err != nil {
		return nil, errors.New("не удалось выполнить запрос к базе данных")
	}
	defer rows.Close()

	var products []types.Product
	for rows.Next() {
		var product types.Product
		if err := rows.Scan(&product.Id, &product.CreatedAt, &product.Name, &product.Category, &product.Price); err != nil {
			return nil, errors.New("ошибка чтения данных")
		}
		products = append(products, product)
	}

	return products, nil
}

// Возвращает статистику по загруженным данным
func GetStatistics(products []types.Product) (types.GetPricesResponse, error) {
	var response types.GetPricesResponse

	// Получаем все статистические данные одним запросом
	var dbDupsCount, totalCategories int
	var totalPrice float64
	err := db.QueryRow(`
		SELECT 
			COUNT(*) - COUNT(DISTINCT (created_at, name, category, price)) as duplicates,
			COUNT(DISTINCT category) as categories,
			COALESCE(SUM(price), 0) as total_price
		FROM prices
	`).Scan(&dbDupsCount, &totalCategories, &totalPrice)
	if err != nil {
		return response, fmt.Errorf("ошибка получения статистики из БД: %w", err)
	}

	response.TotalCount = len(products)
	response.DuplicatesCount = dbDupsCount
	response.TotalItems = countUniqueProducts(products)
	response.TotalCategories = totalCategories
	response.TotalPrice = totalPrice

	return response, nil
}

// Подсчитывает количество уникальных товаров по всем полям, кроме id
func countUniqueProducts(products []types.Product) int {
	uniqueCount := 0
	for i := 0; i < len(products); i++ {
		isUnique := true
		for j := 0; j < i; j++ {
			if products[i].Name == products[j].Name &&
				products[i].Category == products[j].Category &&
				products[i].Price == products[j].Price &&
				products[i].CreatedAt == products[j].CreatedAt {
				isUnique = false
				break
			}
		}
		if isUnique {
			uniqueCount++
		}
	}
	return uniqueCount
}

// Преобразует CSV-строку в структуру Product
func MapRecordToProduct(record []string) (types.Product, error) {
	id, err := strconv.Atoi(record[0])
	if err != nil {
		return types.Product{}, errors.New("неверный формат Id")
	}

	price, err := strconv.ParseFloat(record[3], 64)
	if err != nil {
		return types.Product{}, errors.New("неверный формат цены")
	}

	return types.Product{
		Id:        id,
		CreatedAt: record[4],
		Name:      record[1],
		Category:  record[2],
		Price:     price,
	}, nil
}

// Вставляет данные о продукте в базу данных
func InsertProductIntoDB(product types.Product) error {
	_, err := db.Exec(`
		INSERT INTO prices (id, created_at, name, category, price) 
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (id) DO NOTHING`,
		product.Id, product.CreatedAt, product.Name, product.Category, product.Price)
	return err
}

// Получает отфильтрованные данные из БД
func FetchFilteredData(start, end string, min, max float64) ([]types.Product, error) {
	query := `
		SELECT id, created_at, name, category, price 
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

	var products []types.Product
	// Пропускаем заголовки при обработке
	for i := 1; i < len(records); i++ {
		product, err := MapRecordToProduct(records[i])
		if err != nil {
			return types.GetPricesResponse{}, fmt.Errorf("ошибка обработки записи %d: %w", i, err)
		}

		if err := InsertProductIntoDB(product); err != nil {
			return types.GetPricesResponse{}, fmt.Errorf("ошибка вставки в БД: %w", err)
		}
		products = append(products, product)
	}

	return GetStatistics(products)
}
