package repository

import (
	"encoding/csv"
	"fmt"
	"log"
	"market-engine-go/internal/models"
	"os"
	"path/filepath"
	"time"
)

type CsvStockRepository struct {
	Dir string
}

func NewCsvStockRepository(dir string) *CsvStockRepository {
	_ = os.MkdirAll(dir, os.ModePerm)
	return &CsvStockRepository{Dir: dir}
}

func (r *CsvStockRepository) SaveAll(stocks []models.Stock) error {
	filePath := filepath.Join(r.Dir, fmt.Sprintf("stocks_idx_%s.csv", time.Now().Format("2_01_2006")))
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	writer.Write([]string{"code", "name", "high", "low", "close", "change", "volume", "value", "frequency"})

	for _, s := range stocks {
		writer.Write([]string{
			s.Code,
			s.Name,
			s.High,
			s.Low,
			s.Close,
			s.Change,
			s.Volume,
			s.Value,
			s.Frequency,
		})
	}
	fmt.Printf("CSV saved successfully to: %s\n", filePath)
	return nil
}

func (r *CsvStockRepository) ReadStockSnapshotCsv(filename string) ([]models.Stock, error) {
	file, err := os.Open(fmt.Sprintf("%s/%s", r.Dir, filename))

	if err != nil {
		log.Printf("Error while reading file: %v", err)
		return nil, err
	}

	defer file.Close()

	reader := csv.NewReader(file)

	records, err := reader.ReadAll()

	if err != nil {
		log.Printf("Error reading records: %v", err)
		return nil, err
	}

	var stocks []models.Stock
	for _, record := range records {
		stocks = append(stocks, models.Stock{
			Code:      record[0],
			Name:      record[1],
			High:      record[2],
			Low:       record[3],
			Close:     record[4],
			Change:    record[5],
			Volume:    record[6],
			Value:     record[7],
			Frequency: record[8],
		})
	}

	return stocks, nil
}
