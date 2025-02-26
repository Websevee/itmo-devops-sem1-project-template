package repository

import (
	"archive/tar"
	"archive/zip"
	"database/sql"
	"encoding/csv"
	"errors"
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

// Общая структура для хранения статистики обработки
type processStats struct {
	totalItems int
	categories map[string]bool
	totalPrice float64
}

// Создает новый экземпляр статистики
func newProcessStats() *processStats {
	return &processStats{
		categories: make(map[string]bool),
	}
}

// Преобразует статистику в ответ API
func (ps *processStats) toResponse() types.GetPricesResponse {
	return types.GetPricesResponse{
		TotalItems:      ps.totalItems,
		TotalCategories: len(ps.categories),
		TotalPrice:      ps.totalPrice,
	}
}

// Обрабатывает zip-архив и сохраняет данные в базу данных
func ProcessZip(zipPath string) (types.GetPricesResponse, error) {
	zipReader, err := zip.OpenReader(zipPath)
	if err != nil {
		return types.GetPricesResponse{}, errors.New("не удалось открыть zip-файл")
	}
	defer zipReader.Close()

	stats := newProcessStats()

	for _, f := range zipReader.File {
		if !strings.HasSuffix(f.Name, "data.csv") {
			continue
		}

		if err := processZipFile(f, stats); err != nil {
			return types.GetPricesResponse{}, err
		}
	}

	return stats.toResponse(), nil
}

// Обрабатывает один файл из zip-архива
func processZipFile(f *zip.File, stats *processStats) error {
	csvFile, err := f.Open()
	if err != nil {
		return errors.New("не удалось открыть CSV файл")
	}
	defer csvFile.Close()

	return processCSVReader(csv.NewReader(csvFile), stats)
}

// Обрабатывает tar-архив
func ProcessTar(filename string) (types.GetPricesResponse, error) {
	file, err := os.Open(filename)
	if err != nil {
		return types.GetPricesResponse{}, errors.New("не удалось открыть tar-файл")
	}
	defer file.Close()

	stats := newProcessStats()
	tr := tar.NewReader(file)

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return types.GetPricesResponse{}, errors.New("ошибка чтения tar-архива")
		}

		if header.Typeflag == tar.TypeReg && strings.HasSuffix(header.Name, ".csv") {
			if err := processCSVReader(csv.NewReader(tr), stats); err != nil {
				return types.GetPricesResponse{}, err
			}
		}
	}

	return stats.toResponse(), nil
}

// Обрабатывает CSV-данные из reader
func processCSVReader(reader *csv.Reader, stats *processStats) error {
	records, err := reader.ReadAll()
	if err != nil {
		return errors.New("ошибка чтения CSV файла")
	}

	// Пропускаем заголовок
	for _, record := range records[1:] {
		if err := processRecord(record, stats); err != nil {
			return err
		}
	}

	return nil
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

// Обрабатывает одну запись из CSV и обновляет статистику
func processRecord(record []string, stats *processStats) error {
	product, err := mapRecordToProduct(record)
	if err != nil {
		return err
	}

	if err := insertProductIntoDB(db, product); err != nil {
		return errors.New("ошибка вставки в базу данных: " + err.Error())
	}

	stats.totalItems++
	stats.categories[product.Category] = true
	stats.totalPrice += product.Price

	return nil
}

// Преобразует CSV-строку в структуру Product
func mapRecordToProduct(record []string) (types.Product, error) {
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
func insertProductIntoDB(db *sql.DB, product types.Product) error {
	_, err := db.Exec("INSERT INTO prices (id, created_at, name, category, price) VALUES ($1, $2, $3, $4, $5)",
		product.Id, product.CreatedAt, product.Name, product.Category, product.Price)
	return err
}
