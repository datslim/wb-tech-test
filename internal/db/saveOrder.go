package db

import (
	"context"
	"log"
	"wb-tech-test/internal/model"
)

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
	_, err := db.Pool.Exec(ctx, `
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
