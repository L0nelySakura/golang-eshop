package cache

import (
	"go-postgres-docker/internal/model"
	"sync"
)

type OrderCache struct {
	mu     sync.RWMutex
	orders map[string]*model.Order
}

func NewOrderCache() *OrderCache {
	return &OrderCache{
		orders: make(map[string]*model.Order),
	}
}

func (c *OrderCache) Get(uid string) (*model.Order, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	order, exists := c.orders[uid]
	return order, exists
}

func (c *OrderCache) Set(order *model.Order) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.orders[order.OrderUID] = order
}
