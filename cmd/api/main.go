package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"wb-tech-test/internal/api"
	"wb-tech-test/internal/cache"
	"wb-tech-test/internal/db"

	"github.com/joho/godotenv"
)

func main() {
	database := db.NewDB()      // создаем новый пул соединений с БД
	defer database.Pool.Close() // закрываем пул соединений с БД

	ctx := context.Background() // создаем новый контекст

	orders, err := database.GetAllOrders(ctx) // получаем все заказы из БД
	if err != nil {
		fmt.Println("Ошибка при получении всех заказов:", err)
		return
	}
	orderCache := cache.NewOrderCache() // создаем новый кэш
	orderCache.Restore(orders)          // записываем в кэш данные, полученные из БД

	// Создаём и запускаем HTTP-сервер
	Server := api.NewServer(database, orderCache)

	err = godotenv.Load() // загрузка переменных окружения из файла .env
	if err != nil {
		log.Fatalf("Ошибка при загрузке файла .env для API: %v", err)
	}

	port := os.Getenv("API_PORT")
	if port == "" {
		port = "8081"
	}
	log.Printf("API server запущен на порту :%s\n", port)
	log.Println(Server.Start(":" + port))
}
