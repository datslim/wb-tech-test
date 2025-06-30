package cache

import (
	"sync"

	"wb-tech-test/internal/model"
)

// структура для кеша заказов
type OrderCache struct {
	mu    sync.RWMutex           // для безопасного доступа к кешу
	cache map[string]model.Order // ключ - orderUID, значение - экземпляр структуры заказа
}

// конструктор для создания нового кеша
func NewOrderCache() *OrderCache {
	return &OrderCache{
		cache: make(map[string]model.Order),
	}
}

// добавление заказа в кеш
func (c *OrderCache) Set(order model.Order) {
	c.mu.Lock()                     // блокируем кеш для записи
	defer c.mu.Unlock()             // разблокируем кеш после записи
	c.cache[order.OrderUID] = order // добавляем заказ в кеш по ключу orderUID
}

// получение заказа из кеша
func (c *OrderCache) Get(orderUID string) (model.Order, bool) {
	c.mu.RLock()                   // блокируем кеш для чтения
	defer c.mu.RUnlock()           // разблокируем кеш после чтения
	order, ok := c.cache[orderUID] // получаем заказ из кеша по ключу orderUID
	return order, ok               // возвращаем заказ и флаг, указывающий, существует ли заказ в кеше (1 - заказ существует, 0 - заказ не существует)
}
