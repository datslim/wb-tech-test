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

// восстановление кэша (например, при первичном заполнении)
func (c *OrderCache) Restore(orders []model.Order) {
	c.mu.Lock()
	defer c.mu.Unlock()
	for _, order := range orders {
		c.cache[order.OrderUID] = order
	}
}

// получение заказа из кеша
// возвращаемое значение экземпляр типа Order и флаг указывающий на то, существует ли заказ в кеше или нет
func (c *OrderCache) Get(orderUID string) (model.Order, bool) {
	c.mu.RLock()                   // блокируем кеш для чтения
	defer c.mu.RUnlock()           // разблокируем кеш после чтения
	order, ok := c.cache[orderUID] // получаем заказ из кеша по ключу orderUID
	return order, ok               // возвращаем заказ и флаг, указывающий, существует ли заказ в кеше (1 - заказ существует, 0 - заказ не существует)
}

// получение всех заказов из кеша
// возвращаемое значение: слайс типа Order
func (c *OrderCache) GetAll() []model.Order {
	c.mu.RLock()
	defer c.mu.RUnlock()
	orders := make([]model.Order, 0, len(c.cache))
	for _, order := range c.cache {
		orders = append(orders, order)
	}

	return orders
}
