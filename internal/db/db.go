package db


import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	"log"

	"wb-tech-test/internal/model"
)


const (
	dbURL = "postgres://wb_user:wb-tech-pass@localhost:5432/wb_orders" // константа для подключения к БД, решил не использовать переменные окружения поскольку это тестовый проект
)

// структура для хранения пула соединений с базой данных
type DB struct {
	Pool *pgxpool.Pool // функция из библиотеки pgx для создания пула соединений
}

// конструктор для создания нового пула соединений
// возвращаемое значение: указатель на структуру DB
func NewDB() *DB {
	pool, err := pgxpool.New(context.Background(), dbURL)
	if err != nil {
		log.Fatalf("Ошибка при создании пула: %v", err)
	}

	return &DB{
		Pool: pool,
	}
}

// функция для получения заказа (а также связанные delivery, payment и items) по order_uid
// возвращаемое значение: экземпляр структуры Order и ошибку, если заказ не найден
func (db *DB) GetOrder(ctx context.Context, orderUID string) (model.Order, error) {
	var order model.Order // объявляем экземпляр типа Order

	// получаем основную информацию о заказе
	row := db.Pool.QueryRow(ctx, `
	SELECT order_uid, track_number, entry, locale, internal_signature, customer_id, delivery_service, shardkey, sm_id, date_created, oof_shard
	FROM orders
	WHERE order_uid = $1
	`, orderUID)

	// заполняем наш экземпляр order полученными значениями
	err := row.Scan(
		&order.OrderUID,
		&order.TrackNumber,
		&order.Entry,
		&order.Locale,
		&order.InternalSignature,
		&order.CustomerID,
		&order.DeliveryService,
		&order.ShardKey,
		&order.SmID,
		&order.DateCreated,
		&order.OofShard,
	)

	// логгируем и возвращаем ошибку, если таковая есть
	if err != nil { 
		log.Printf("Ошибка получения заказа %s: %v", orderUID, err)
		return order, err
	}

	delivery, err := db.getDeliveryByOrderUID(ctx, orderUID)	// получаем связанные данные о доставке
	if err != nil {
		return order, nil
	}
	order.Delivery = delivery	// добавляем в order полученные данные о доставке


	payment, err := db.getPaymentByOrderUID(ctx, orderUID) // получаем связанные данные об оплате
	if err != nil {
		return order, err
	}
	order.Payment = payment // добавляем в order полученные данные об оплате

	items, err := db.getItemsByOrderUID(ctx, orderUID) // получаем связанные данные о товарах
	if err != nil {
		return order, err
	}

	order.Items = items // добавляем в order полученные данные о товарах
	
	return order, nil

}

// функция для получения delivery по order_uid
// возвращаемое значение: экземпляр структуры Delivery и ошибку, если delivery не найден
func (db *DB) getDeliveryByOrderUID(ctx context.Context, orderUID string) (model.Delivery, error) {
	var delivery model.Delivery // объявляем экземпляр структуры Delivery

	// получаем основную информацию о доставке
	row := db.Pool.QueryRow(ctx, `
	SELECT name, phone, zip, city, address, region, email
	FROM delivery
	WHERE order_uid = $1
	`, orderUID)

	// заполняем наш экземпляр delivery полученными значениями
	err := row.Scan(
		&delivery.Name,
		&delivery.Phone,
		&delivery.Zip,
		&delivery.City,
		&delivery.Address,
		&delivery.Region,
		&delivery.Email,
	)

	// логгируем и возвращаем ошибку, если таковая есть
	if err != nil {
		log.Printf("Ошибка получения delivery для заказа %s: %v", orderUID, err)
		return delivery, err
	}
	log.Printf("Delivery для заказа %s успешно получен", orderUID) // логгируем сообщение об успешном получении delivery
	return delivery, nil

}

// функция для получения payment по order_uid
// возвращаемое значение: экземпляр структуры Payment и ошибку, если payment не найден
func (db *DB) getPaymentByOrderUID(ctx context.Context, orderUID string) (model.Payment, error) {
	var payment model.Payment // объявляем экземпляр структуры Payment

	// получаем основную информацию об оплате
	row := db.Pool.QueryRow(ctx, `
	SELECT transaction, request_id, currency, provider, amount, payment_dt, bank, delivery_cost, goods_total, custom_fee
	FROM payment
	WHERE order_uid = $1
	`, orderUID)

	// заполняем наш экземпляр payment полученными значениями
	err := row.Scan(
		&payment.Transaction,
		&payment.RequestID,
		&payment.Currency,
		&payment.Provider,
		&payment.Amount,
		&payment.PaymentDT,
		&payment.Bank,
		&payment.DeliveryCost,
		&payment.GoodsTotal,
		&payment.CustomFee,
	)

	// логгируем и возвращаем ошибку, если таковая есть
	if err != nil {
		log.Printf("Ошибка получения payment для заказа %s: %v", orderUID, err)
		return payment, err
	}
	log.Printf("Payment для заказа %s успешно получен", orderUID) // логгируем сообщение об успешном получении payment
	return payment, nil
}


// функция для получения items по order_uid
// возвращаемое значение: слайс экземпляров структуры Item и ошибку, если items не найдены
func (db *DB) getItemsByOrderUID(ctx context.Context, orderUID string) ([]model.Item, error) {
	var items []model.Item // объявляем слайс экземпляров структуры Item

	// получаем основную информацию о товарах
	rows, err := db.Pool.Query(ctx, `
		SELECT chrt_id, track_number, price, rid, name, sale, size, total_price, nm_id, brand, status
		FROM items
		WHERE order_uid = $1
	`, orderUID)

	// закрываем соединение с БД
	defer rows.Close()


	// логгируем и возвращаем ошибку, если таковая есть
	if err != nil {
		log.Printf("Ошибка получения items для заказа %s: %v", orderUID, err)
	}

	// заполняем наш слайс items полученными значениями
	for rows.Next() {
		var item model.Item
		err := rows.Scan(
			&item.ChrtID,
			&item.TrackNumber,
			&item.Price,
			&item.Rid,
			&item.Name,
			&item.Sale,
			&item.Size,
			&item.TotalPrice,
			&item.NmID,
			&item.Brand,
			&item.Status,
		)

		// логгируем и возвращаем ошибку, если таковая есть
		if err != nil {
			log.Printf("Ошибка сканирования item для заказа %s: %v", orderUID, err)
			return nil, err
		}
		items = append(items, item)
	}
	log.Printf("Для заказа %s найдено %d items", orderUID, len(items)) // логгируем сообщение об успешном получении items
	return items, nil
}

// функция для сохранения заказа в БД
// возвращаемое значение: ошибка, если заказ не сохранен
func (db *DB) SaveOrder(ctx context.Context, order model.Order) error {
	// сохраняем заказ в таблицу orders
	if err := db.insertOrder(ctx, order); err != nil {
		return err
	}

	// сохраняем delivery в таблицу delivery
	if err := db.insertDelivery(ctx, order.OrderUID, order.Delivery); err != nil {
		return err
	}
	
	// сохраняем payment в таблицу payment
	if err := db.insertPayment(ctx, order.OrderUID, order.Payment); err != nil {
		return err
	}
	
	// сохраняем items в таблицу items
	if err := db.insertItems(ctx, order.OrderUID, order.Items); err != nil {
		return err
	}
	
	return nil
}

// функция для сохранения заказа в таблицу orders
// возвращаемое значение: ошибка, если заказ не сохранен
func (db *DB) insertOrder(ctx context.Context, order model.Order) error { 
	_, err := db.Pool.Exec(ctx, `
		INSERT INTO orders (order_uid, track_number, entry, locale, internal_signature, customer_id, delivery_service, shardkey, sm_id, date_created, oof_shard)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
		ON CONFLICT (order_uid) DO NOTHING
	`,

		order.OrderUID, order.TrackNumber, order.Entry, order.Locale, order.InternalSignature,
		order.CustomerID, order.DeliveryService, order.ShardKey, order.SmID, order.DateCreated, order.OofShard,
	)
	if err != nil {
		// логгируем и возвращаем ошибку, если таковая есть
		log.Printf("Ошибка сохранения заказа %s в таблицу orders: %v", order.OrderUID, err)
		return err
	}
	// при успешном сохранении заказа в таблице orders, логгируем сообщение
	log.Printf("Заказ %s успешно сохранён в таблице orders", order.OrderUID)
	
	return nil
}

// функция для сохранения заказа в таблицу delivery
// возвращаемое значение: ошибка, если заказ не сохранен
func (db *DB) insertDelivery(ctx context.Context, orderUID string, delivery model.Delivery) error { 
	_, err := db.Pool.Exec(ctx,`
		INSERT INTO delivery (order_uid, name, phone, zip, city, address, region, email)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
		ON CONFLICT (order_uid) DO NOTHING
	`,
	orderUID, delivery.Name, delivery.Phone, delivery.Zip, delivery.City, delivery.Address, delivery.Region, delivery.Email,
	)

	if err != nil {
		// логгируем и возвращаем ошибку, если таковая есть
		log.Printf("Ошибка сохранения delivery для заказа %s в таблицу delivery: %v", orderUID, err)
		return err
	}
	// при успешном сохранении delivery в таблице delivery, логгируем сообщение
	log.Printf("Delivery для заказа %s успешно сохранён", orderUID)

	return nil
}

// функция для сохранения заказа в таблицу payment
// возвращаемое значение: ошибка, если заказ не сохранен
func (db *DB) insertPayment(ctx context.Context, orderUID string, payment model.Payment) error { 
	_, err := db.Pool.Exec(ctx, `
		INSERT INTO payment (transaction, order_uid, request_id, currency, provider, amount, payment_dt, bank, delivery_cost, goods_total, custom_fee)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
		ON CONFLICT (transaction) DO NOTHING
	`,
	payment.Transaction, orderUID, payment.RequestID, payment.Currency, payment.Provider, payment.Amount, payment.PaymentDT, payment.Bank, payment.DeliveryCost, payment.GoodsTotal, payment.CustomFee,
	)
	if err != nil {
		// логгируем и возвращаем ошибку, если таковая есть
		log.Printf("Ошибка сохранения payment для заказа %s в таблицу payment: %v", orderUID, err)
		return err
	}

	// при успешном сохранении payment в таблице payment, логгируем сообщение
	log.Printf("Payment для заказа %s успешно сохранён", orderUID)
	
	return nil

}

// функция для сохранения заказа в таблицу items
// возвращаемое значение: ошибка, если заказ не сохранен
func (db *DB) insertItems(ctx context.Context, orderUID string, items []model.Item) error { 
	for _, item := range items {
		_, err := db.Pool.Exec(ctx, `
			INSERT INTO items (order_uid, chrt_id, track_number, price, rid, name, sale, size, total_price, nm_id, brand, status)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)
		`,
			orderUID, item.ChrtID, item.TrackNumber, item.Price, item.Rid, item.Name, item.Sale, item.Size, item.TotalPrice, item.NmID, item.Brand, item.Status,
		)
		if err != nil {
			// логгируем и возвращаем ошибку, если таковая есть
		log.Printf("Ошибка сохранения items для заказа %s в таблицу items: %v", orderUID, err)
			return err
		}
	}

	// при успешном сохранении items в таблице items, логгируем сообщение
	log.Printf("Items для заказа %s успешно сохранены: %d шт.", orderUID, len(items))
	return nil
}

// функция для получения всех заказов
// возвращаемое значение: слайс экземпляров структуры Order и ошибка, если заказы не найдены
func (db *DB) GetAllOrders(ctx context.Context) ([]model.Order, error) {
	var orders []model.Order // объявляем слайс экземпляров структуры Order

	// получаем все заказы из таблицы orders	
	rows, err := db.Pool.Query(ctx, `
		SELECT order_uid FROM orders
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close() // закрываем соединение с БД

	// заполняем наш слайс orders полученными значениями из таблицы orders
	for rows.Next() {
		var orderUID string
		if err := rows.Scan(&orderUID); err != nil {
			// логгируем и возвращаем ошибку, если таковая есть
			log.Printf("Ошибка сканирования order_uid: %v", err)
			continue // пропускаем ошибочные заказы
		}

		order, err := db.GetOrder(ctx, orderUID)
		if err != nil {
			// логгируем и возвращаем ошибку, если таковая есть
			log.Printf("Ошибка получения заказа %s: %v", orderUID, err)
			continue // пропускаем ошибочные заказы
		}
		orders = append(orders, order) // добавляем в наш слайс orders полученный заказ
	}

	return orders, nil

}
