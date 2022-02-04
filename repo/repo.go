package repo

import (
	"context"
	"fmt"
	"math"

	"github.com/psaung/go-concurrency/db"
	"github.com/psaung/go-concurrency/models"
	"github.com/psaung/go-concurrency/stats"
)

type repo struct {
	products  *db.ProductDB
	orders    *db.OrderDB
	stats     stats.StatsService
	incoming  chan models.Order
	done      chan struct{}
	processed chan models.Order
}

type Repo interface {
	CreateOrder(item models.Item) (*models.Order, error)
	GetAllProducts() []models.Product
	GetOrder(id string) (models.Order, error)
	Close()
	GetOrderStats(ctx context.Context) (models.Statistics, error)
	RequestReversal(orderId string) (*models.Order, error)
}

func New() (Repo, error) {
	processed := make(chan models.Order, stats.WorkerCount)
	done := make(chan struct{})
	p, err := db.NewProducts()
	if err != nil {
		return nil, err
	}
	statsService := stats.New(processed, done)
	o := repo{
		products:  p,
		orders:    db.NewOrders(),
		stats:     statsService,
		incoming:  make(chan models.Order),
		done:      done,
		processed: processed,
	}

	go o.processOrders()
	return &o, nil
}

func (r *repo) GetAllProducts() []models.Product {
	return r.products.FindAll()
}

func (r *repo) GetOrder(id string) (models.Order, error) {
	return r.orders.Find(id)
}

func (r *repo) CreateOrder(item models.Item) (*models.Order, error) {
	if err := r.validateItem(item); err != nil {
		return nil, err
	}

	order := models.NewOrder(item)
	select {
	case r.incoming <- order:
		r.orders.Upsert(order)
		return &order, nil
	case <-r.done:
		return nil, fmt.Errorf("orders app is closed, try again later")
	}
}

func (r *repo) validateItem(item models.Item) error {
	if item.Amount < 1 {
		return fmt.Errorf("order amount must be at leaset 1:got %d", item.Amount)
	}
	if err := r.products.Exists(item.ProductID); err != nil {
		return fmt.Errorf("product %s does not exist", item.ProductID)
	}
	return nil
}

func (r *repo) processOrders() {
	fmt.Println("Order processing started!")
	for {
		select {
		case order := <-r.incoming:
			r.processOrder(&order)
			r.orders.Upsert(order)
			fmt.Printf("Processing order %s complted\n", order.ID)
			r.processed <- order
		case <-r.done:
			fmt.Println("Order processing stopped!")
			return
		}
	}
}

func (r *repo) processOrder(order *models.Order) {
	fetchedOrder, err := r.orders.Find(order.ID)
	if err != nil || fetchedOrder.Status != models.OrderStatus_Completed {
		fmt.Println("duplicate reveral on order", order.ID)
	}

	item := order.Item
	if order.Status == models.OrderStatus_ReversalRequested {
		item.Amount = -item.Amount
	}

	product, err := r.products.Find(item.ProductID)
	if err != nil {
		order.Status = models.OrderStatus_Rejected
		order.Error = err.Error()
		return
	}

	if product.Stock < item.Amount {
		order.Status = models.OrderStatus_Rejected
		order.Error = fmt.Sprintf("not enough stock for prodcut %s:got %d, want %d", item.ProductID, product.Stock, item.Amount)
		return
	}

	remainingStock := product.Stock - item.Amount
	product.Stock = remainingStock
	r.products.Upsert(product)

	total := math.Round(float64(order.Item.Amount)*product.Price*100) / 100
	order.Total = &total
	order.Complete()
}

func (r *repo) Close() {
	close(r.done)
}

func (r repo) GetOrderStats(ctx context.Context) (models.Statistics, error) {
	select {
	case s := <-r.stats.GetStats(ctx):
		return s, nil
	case <-ctx.Done():
		return models.Statistics{}, ctx.Err()
	}
}

func (r repo) RequestReversal(orderId string) (*models.Order, error) {
	order, err := r.orders.Find(orderId)
	if err != nil {
		return nil, err
	}

	if order.Status != models.OrderStatus_Completed {
		return nil, fmt.Errorf("order status is %s, only completed orders can be requested for reversal", order.Status)
	}

	order.Status = models.OrderStatus_ReversalRequested
	select {
	case r.incoming <- order:
		r.orders.Upsert(order)
		return &order, nil
	case <-r.done:
		return nil, fmt.Errorf("sorry, the orders app is closed")
	}
}
