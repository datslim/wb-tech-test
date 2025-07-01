package main

import (
	"log"
	"os"
	"wb-tech-test/internal/webserver"

	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load() // загрузка переменных окружения из файла .env
	if err != nil {
		log.Fatalf("Ошибка при загрузке файла .env для webserver: %v", err)
	}

	port := os.Getenv("STATIC_PORT")
	if port == "" {
		port = "3000"
	}

	webserver.Start(port)
}
