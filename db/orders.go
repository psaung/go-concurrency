package db

import (
	"fmt"
	"sync"

	"github.com/psaung/go-concurrency/models"
)

type OrderDB struct {
	placeOrders sync.Map
}

func NewOrders() *OrderDB {
	return &OrderDB{}
}

func (o *OrderDB) Find(id string) (models.Order, error) {
	po, ok := o.placeOrders.Load(id)
	if !ok {
		return models.Order{}, fmt.Errorf("no order found for %s order id", id)
	}

	return toOrder(po), nil
}

func (o *OrderDB) Upsert(order models.Order) {
	o.placeOrders.Store(order.ID, order)
}

func toOrder(po interface{}) models.Order {
	order, ok := po.(models.Order)
	if !ok {
		panic(fmt.Errorf("error casting %v to order", po))
	}
	return order
}
