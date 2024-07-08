package cache

import (
	"encoding/json"
	"log"
	"sync"
	"test/internal/entities"
	"test/internal/storage/postgres"
)

// Структура кэша
type OrdersCache struct {
	Orders map[string]entities.Order
	mu     sync.RWMutex
}

// возвращаем кэшированные заказы
func New() *OrdersCache {
	return &OrdersCache{
		Orders: make(map[string]entities.Order),
	}
}

// добавление нового заказа в кэш
func (c *OrdersCache) Add(value entities.Order) bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	key := value.OrderUID
	if _, exists := c.Orders[key]; exists {
		log.Println("Error adding new order to cache: Order with such id already exists")
		return false
	}
	c.Orders[key] = value
	return true
}

// возвращает заказ, если он существует
func (c *OrdersCache) Get(key string) (entities.Order, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	value, ok := c.Orders[key]
	if !ok {
		return entities.Order{}, false
	}
	return value, ok
}

// восстанавливаем данные из постгреса в кэш
func (c *OrdersCache) RecoverFromPostgres(db *postgres.Pgdb) bool {
	var (
		delivery []byte
		payment  []byte
		items    []byte
	)
	rows, err := db.RecoverData()
	if err != nil {
		log.Println(err)
		return false
	}
	for rows.Next() {
		itemsParseStruct := entities.ParseItemsStruct{}
		order := entities.Order{}
		err = rows.Scan(
			&order.OrderUID,
			&order.TrackNumber,
			&order.Entry,
			&delivery,
			&payment,
			&items,
			&order.Locale,
			&order.InternalSignature,
			&order.CustomerID,
			&order.DeliveryService,
			&order.Shardkey,
			&order.SmID,
			&order.DateCreated,
			&order.OofShard)
		if err != nil {
			log.Println(err)
			return false
		}
		err = json.Unmarshal(delivery, &order.Delivery)
		if err != nil {
			log.Println(err)
			return false
		}
		err = json.Unmarshal(payment, &order.Payment)
		if err != nil {
			log.Println(err)
			return false
		}
		err = json.Unmarshal(items, &itemsParseStruct)
		if err != nil {
			log.Println(err)
			return false
		}
		for _, item := range itemsParseStruct.Items {
			order.Items = append(order.Items, item[0])
		}
		c.Add(order)
	}

	log.Println("Recovering finished successfully")
	return true
}
