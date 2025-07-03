package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"time"

	"wb-tech-test/internal/cache"
	"wb-tech-test/internal/db"
	"wb-tech-test/internal/model"

	"github.com/segmentio/kafka-go"
)

// структура для консьюмера
type Consumer struct {
	Reader *kafka.Reader     // ридер сообщений из Kafka
	DB     *db.DB            // БД
	Cache  *cache.OrderCache // кеш
}

const (
	groupID = "order-consumer-group" // группа консьюмеров
)

// функция для создания нового консьюмера
func NewConsumer(brokers []string, topic string, db *db.DB, cache *cache.OrderCache) *Consumer {
	return &Consumer{
		Reader: kafka.NewReader(kafka.ReaderConfig{
			Brokers: brokers,
			Topic:   topic,
			GroupID: groupID,
		}),
		DB:    db,
		Cache: cache,
	}
}

// функция для чтения сообщений из Kafka
func (c *Consumer) Consume() {
	var order model.Order // создаем экземпляр структуры Order

	for {
		msg, err := c.Reader.ReadMessage(context.Background()) // читаем сообщение из Kafka
		if err != nil {
			log.Printf("[KAFKA] Ошибка чтения сообщения: %v", err)
			continue
		}

		// десериализуем сообщение в структуру Order
		if err := json.Unmarshal(msg.Value, &order); err != nil {
			log.Printf("[KAFKA] Ошибка десериализации сообщения: %v", err)
			continue
		}

		// выводим информацию о полученном заказе
		log.Printf("[KAFKA] Получен заказ %s с сообщением: %s", order.OrderUID, string(msg.Value))

		// сохраняем заказ в БД и кеш
		if err := c.ProcessOrder(order); err != nil {
			log.Printf("[KAFKA] Ошибка сохранения заказа: %v", err)
			continue
		}

	}

}

// функция для обработки заказа (сохранение в БД и кеш)
func (c *Consumer) ProcessOrder(order model.Order) error {
	if err := c.DB.SaveOrder(context.Background(), order); err != nil {
		log.Printf("[KAFKA] Ошибка при сохранении заказа %s: %v", order.OrderUID, err)
		return err
	}
	c.Cache.Set(order)
	return nil
}

// функция для проверки существования топика в Kafka
func EnsureTopicExists(broker, topic string, partitions int) error {
	conn, err := kafka.Dial("tcp", broker) // создаем соединение с брокером Kafka
	if err != nil {
		return err
	}

	defer conn.Close() // закрываем соединение с брокером Kafka

	controller, err := conn.Controller() // получаем контроллер топика
	if err != nil {
		return err
	}

	controllerConn, err := kafka.Dial("tcp", controller.Host+":"+strconv.Itoa(controller.Port)) // создаем соединение с контроллером топика

	if err != nil {
		return err
	}

	defer controllerConn.Close() // закрываем соединение с контроллером топика

	err = controllerConn.CreateTopics(kafka.TopicConfig{ // создаем топик
		Topic:             topic,
		NumPartitions:     partitions,
		ReplicationFactor: 1,
	})

	if err != nil && err != kafka.TopicAlreadyExists { // если топик уже существует, то возвращаем ошибку
		return err
	}

	return nil
}

// функция для ожидания доступности Kafka
// возвращаемое значение: ошибка, если Kafka не доступна
func WaitForKafka(brokers []string, topic string, maxRetries int, delay time.Duration) error {
	for i := range maxRetries {
		conn, err := kafka.DialLeader(context.Background(), "tcp", brokers[0], topic, 0)
		if err == nil {
			conn.Close()
			log.Println("[MAIN] Kafka доступна!")
			return nil
		}
		log.Printf("[MAIN] Ожидание Kafka (%d/%d): %v", i+1, maxRetries, err)
		time.Sleep(delay)
	}
	return fmt.Errorf("kafka не доступна после %d попыток", maxRetries)
}
