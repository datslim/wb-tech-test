package main

import (
	"context"
	"log"
	"os"
	"time"
	"wb-tech-test/internal/api"
	"wb-tech-test/internal/cache"
	"wb-tech-test/internal/db"
	"wb-tech-test/internal/kafka"

	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load() // загрузка переменных окружения из файла .env
	if err != nil {
		log.Fatalf("[MAIN] Ошибка при загрузке файла .env для API: %v", err)
	}

	ctx := context.Background() // создаем новый контекст

	database := db.NewDB()              // создаем новый пул соединений с БД
	defer database.Pool.Close()         // закрываем пул соединений с БД
	orderCache := cache.NewOrderCache() // создаем новый кэш

	// восстанавливаем кэш из БД
	if err := restoreCache(ctx, database, orderCache); err != nil {
		log.Printf("[MAIN] Ошибка при восстановлении кэша: %v", err)
	}

	// проверяем, что топик в Kafka существует
	err = kafka.EnsureTopicExists("wb-kafka:9092", "orders", 1)
	if err != nil {
		log.Fatalf("[MAIN] Ошибка при создании топика: %v", err)
	}

	// ожидаем доступности Kafka
	if err := kafka.WaitForKafka([]string{"wb-kafka:9092"}, "orders", 10, 5*time.Second); err != nil {
		log.Fatalf("[MAIN] Kafka недоступна: %v", err)
	}

	// создаем и запускаем нового консьюмера
	consumer := kafka.NewConsumer([]string{"wb-kafka:9092"}, "orders", database, orderCache)
	go consumer.Consume()

	// Создаём и запускаем HTTP-сервер
	Server := api.NewServer(database, orderCache)

	port := getPort() // получаем порт из переменных окружения
	if err := Server.Start(":" + port); err != nil {
		log.Printf("[MAIN] Ошибка при запуске HTTP-сервера: %v", err)
	}
	log.Printf("[MAIN] API server запущен на порту :%s\n", port)
}

// функция для получения порта из переменных окружения
// возвращаемое значение: порт для запуска HTTP-сервера
func getPort() string {
	if port := os.Getenv("API_PORT"); port != "" {
		return port
	}
	return "8081"
}

// функция для восстановления кэша из БД
// возвращаемое значение: ошибка, если кэш не восстановлен
func restoreCache(ctx context.Context, database *db.DB, orderCache *cache.OrderCache) error {
	orders, err := database.GetAllOrders(ctx) // получаем все заказы из БД
	if err != nil {
		log.Printf("[MAIN] Ошибка при получении всех заказов для загрузки в кэш: %v", err)
		return err
	}
	orderCache.Restore(orders) // записываем в кэш данные, полученные из БД
	log.Printf("[MAIN] Кэш восстановлен, количество заказов: %d", len(orders))
	return nil
}
