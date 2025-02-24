package repository

import (
	"archive/zip"
	"database/sql"
	"encoding/csv"
	"errors"
	"itmo-devops-fp1/internal/types"
	"itmo-devops-fp1/pkg/utils"
	"strconv"
	"strings"
)

var db *sql.DB

func init() {
	db = utils.ConnectDB()
}

// обрабатывает zip-архив и сохраняет данные в базу данных
func ProcessZip(zipPath string) (types.GetPricesResponse, error) {
	zipReader, err := zip.OpenReader(zipPath)
	if err != nil {
		return types.GetPricesResponse{}, errors.New("unable to open zip file")
	}
	defer zipReader.Close()

	var totalItems int
	var totalPrice float64
	categories := make(map[string]bool)

	for _, f := range zipReader.File {
		if strings.HasSuffix(f.Name, "data.csv") {
			csvFile, err := f.Open()
			if err != nil {
				return types.GetPricesResponse{}, errors.New("unable to open CSV file")
			}
			defer csvFile.Close()

			reader := csv.NewReader(csvFile)
			records, err := reader.ReadAll()
			if err != nil {
				return types.GetPricesResponse{}, errors.New("unable to read CSV file")
			}

			// убираем шапку
			records = records[1:]

			for _, record := range records {
				err := processRecord(record, db, &totalItems, categories, &totalPrice)
				if err != nil {
					return types.GetPricesResponse{}, err
				}
			}
		}
	}

	return types.GetPricesResponse{
		TotalItems:      totalItems,
		TotalCategories: len(categories),
		TotalPrice:      totalPrice,
	}, nil
}

// извлекает данные из базы данных
func FetchData() ([]types.Product, error) {
	rows, err := db.Query("SELECT id, created_at, name, category, price FROM prices")
	if err != nil {
		return nil, errors.New("unable to query database")
	}
	defer rows.Close()

	var products []types.Product
	for rows.Next() {
		var product types.Product
		err := rows.Scan(&product.Id, &product.CreatedAt, &product.Name, &product.Category, &product.Price)
		if err != nil {
			return nil, errors.New("unable to scan row")
		}
		products = append(products, product)
	}

	return products, nil
}

// обрабатывает одну запись из CSV и вставляет данные в базу данных
func processRecord(record []string, db *sql.DB, totalItems *int, categories map[string]bool, totalPrice *float64) error {
	product, err := mapRecordToProduct(record)
	if err != nil {
		return err
	}

	err = insertProductIntoDB(db, product)
	if err != nil {
		return errors.New("unable to insert data into database" + err.Error())
	}

	// обновляем статистику
	*totalItems++
	categories[product.Category] = true
	*totalPrice += product.Price

	return nil
}

// преобразует CSV-строку в структуру Product
func mapRecordToProduct(record []string) (types.Product, error) {
	id, err := strconv.Atoi(record[0])
	if err != nil {
		return types.Product{}, errors.New("invalid Id format")
	}

	price, err := strconv.ParseFloat(record[3], 64)
	if err != nil {
		return types.Product{}, errors.New("invalid price format")
	}

	return types.Product{
		Id:        id,
		CreatedAt: record[4],
		Name:      record[1],
		Category:  record[2],
		Price:     price,
	}, nil
}

// вставляет данные о продукте в базу данных
func insertProductIntoDB(db *sql.DB, product types.Product) error {
	_, err := db.Exec("INSERT INTO prices (id, created_at, name, category, price) VALUES ($1, $2, $3, $4, $5)",
		product.Id, product.CreatedAt, product.Name, product.Category, product.Price)
	return err
}
