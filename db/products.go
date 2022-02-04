package db

import (
	"fmt"
	"sort"
	"sync"

	"github.com/psaung/go-concurrency/models"
	"github.com/psaung/go-concurrency/utils"
)

type ProductDB struct {
	products sync.Map
}

func NewProducts() (*ProductDB, error) {
	p := &ProductDB{}
	if err := utils.ImportProducts(&p.products); err != nil {
		return nil, err
	}
	return p, nil
}

func (p *ProductDB) Exists(id string) error {
	if _, ok := p.products.Load(id); !ok {
		return fmt.Errorf("no product fond for id %s", id)
	}

	return nil
}

func (p *ProductDB) Find(id string) (models.Product, error) {
	pp, ok := p.products.Load(id)
	if !ok {
		return models.Product{}, fmt.Errorf("no product found for id %s", id)
	}

	return toProduct(pp), nil
}

func (p *ProductDB) Upsert(prod models.Product) {
	p.products.Store(prod.ID, prod)
}

func (p *ProductDB) FindAll() []models.Product {
	var allProducts []models.Product
	p.products.Range(func(_, value interface{}) bool {
		allProducts = append(allProducts, toProduct(value))
		return true
	})

	sort.Slice(allProducts, func(i, j int) bool {
		return allProducts[i].ID < allProducts[j].ID
	})
	return allProducts
}

func toProduct(pp interface{}) models.Product {
	prod, ok := pp.(models.Product)
	if !ok {
		panic(fmt.Errorf("error casting %w to product", pp))
	}
	return prod
}
