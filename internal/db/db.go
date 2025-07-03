package db

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

// структура для хранения пула соединений с базой данных
type DB struct {
	Pool *pgxpool.Pool // функция из библиотеки pgx для создания пула соединений
}

// конструктор для создания нового пула соединений
// возвращаемое значение: указатель на структуру DB
func NewDB() *DB {
	err := godotenv.Load() // загрузка переменных окружения из файла .env
	if err != nil {
		log.Fatalf("[DB] Ошибка при загрузке файла .env: %v", err)
	}

	// формирование строки подключения к БД с использованием переменных окружения
	dbURL := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", os.Getenv("PG_USER"), os.Getenv("PG_PASS"), os.Getenv("PG_HOST"), os.Getenv("PG_PORT"), os.Getenv("PG_DB"))
	pool, err := pgxpool.New(context.Background(), dbURL) // создание пула соединений с БД
	if err != nil {
		log.Fatalf("[DB] Ошибка при создании пула: %v", err)
	}

	return &DB{
		Pool: pool,
	}
}

func (db *DB) IsTransactionExists(ctx context.Context, transaction string) (bool, error) {
	var exists bool
	err := db.Pool.QueryRow(ctx, `
        SELECT EXISTS(SELECT 1 FROM payment WHERE transaction = $1)
    `, transaction).Scan(&exists)
	return exists, err
}
