package utils

import (
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
	"sync"

	"github.com/psaung/go-concurrency/models"
)

var productInputPath string = "./input/products.csv"

func ImportProducts(products *sync.Map) error {
	input, err := readCsv(productInputPath)

	if err != nil {
		return err
	}

	for _, line := range input {
		if len(line) != 5 {
			continue
		}

		id := line[0]
		stock, err := strconv.Atoi(line[2])

		if err != nil {
			continue
		}

		price, err := strconv.ParseFloat(line[4], 64)
		if err != nil {
			continue
		}

		products.Store(id, models.Product{
			ID:    id,
			Name:  fmt.Sprintf("%s(%s)", line[1], line[3]),
			Stock: stock,
			Price: price,
		})

	}
	return nil
}

func readCsv(filename string) ([][]string, error) {
	f, err := os.Open(filename)
	if err != nil {
		return [][]string{}, err
	}

	defer f.Close()
	lines, err := csv.NewReader(f).ReadAll()
	if err != nil {
		return [][]string{}, err
	}

	return lines, err
}
