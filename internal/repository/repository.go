package repository

import (
	"database/sql"
	"go-postgres-docker/internal/model"
    "github.com/google/uuid"
)

type OrderRepository struct {
	db *sql.DB
}

func NewProductRepository(db *sql.DB) *OrderRepository {
	return &OrderRepository{db: db}
}

// Выводит все заказы
func (d *OrderRepository) GetOrders() ([]model.Order, error) {
	rows, err := d.db.Query("SELECT order_uid, track_number, entry FROM orders")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []model.Order
	for rows.Next() {
		var p model.Order
		if err := rows.Scan(&p.OrderUID, &p.TrackNumber, &p.Entry); err != nil {
			return nil, err
		}
		orders = append(orders, p)
	}

	return orders, nil
}

// Выводит заказ с определённым UID
func (d *OrderRepository) GetOrderByID(uid string) (*model.Order, error) {
	tx, err := d.db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	order := &model.Order{}
	err = tx.QueryRow(`
		SELECT order_uid, track_number, entry, locale, internal_signature, 
			customer_id, delivery_service, shardkey, sm_id, date_created, oof_shard
		FROM orders WHERE order_uid = $1`, uid).Scan(
		&order.OrderUID, &order.TrackNumber, &order.Entry, &order.Locale, 
		&order.InternalSignature, &order.CustomerID, &order.DeliveryService, 
		&order.Shardkey, &order.SmID, &order.DateCreated, &order.OofShard,
	)
	if err != nil {
		return nil, err
	}

	delivery := &model.Delivery{}
	err = tx.QueryRow(`
		SELECT name, phone, zip, city, address, region, email
		FROM deliveries WHERE order_uid = $1`, uid).Scan(
		&delivery.Name, &delivery.Phone, &delivery.Zip, &delivery.City,
		&delivery.Address, &delivery.Region, &delivery.Email,
	)
	if err != nil {
		return nil, err
	}
	order.Delivery = *delivery


	payment := &model.Payment{}
	err = tx.QueryRow(`
		SELECT transaction, request_id, currency, provider, amount, 
			payment_dt, bank, delivery_cost, goods_total, custom_fee
		FROM payments WHERE order_uid = $1`, uid).Scan(
		&payment.Transaction, &payment.RequestID, &payment.Currency, 
		&payment.Provider, &payment.Amount, &payment.PaymentDt, &payment.Bank,
		&payment.DeliveryCost, &payment.GoodsTotal, &payment.CustomFee,
	)
	if err != nil {
		return nil, err
	}
	order.Payment = *payment

	rows, err := tx.Query(`
		SELECT i.chrt_id, i.track_number, i.price, i.rid, i.name, i.sale, 
			i.size, i.total_price, i.nm_id, i.brand, i.status
		FROM items i
		JOIN order_items oi ON i.item_uid  = oi.item_id
		JOIN orders o 		ON oi.order_id = o.order_uid
		WHERE o.order_uid = $1`, uid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []model.Item
	for rows.Next() {
		item := model.Item{}
		err := rows.Scan(
			&item.ChrtID, &item.TrackNumber, &item.Price, &item.Rid, 
			&item.Name, &item.Sale, &item.Size, &item.TotalPrice, 
			&item.NmID, &item.Brand, &item.Status,
		)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	order.Items = items

	if err = tx.Commit(); err != nil {
		return nil, err
	}

	return order, nil
}

func (d *OrderRepository) SaveOrder(order *model.Order) error {
	// Начало транзакции
	tx, err := d.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Вставки в соответствующие таблицы
	order.OrderUID = uuid.New().String()
	_, err = tx.Exec(`
		INSERT INTO orders (
			order_uid, track_number, entry, locale, internal_signature, 
			customer_id, delivery_service, shardkey, sm_id, date_created, oof_shard
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`,
		order.OrderUID, order.TrackNumber, order.Entry, order.Locale, 
		order.InternalSignature, order.CustomerID, order.DeliveryService, 
		order.Shardkey, order.SmID, order.DateCreated, order.OofShard,
	)
	if err != nil {
		return err
	}

	delivery_uid := uuid.New().String()
	_, err = tx.Exec(`
		INSERT INTO deliveries (
			delivery_uid, order_uid, name, phone, zip, city, address, region, email
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
		delivery_uid, order.OrderUID, order.Delivery.Name, order.Delivery.Phone, 
		order.Delivery.Zip, order.Delivery.City, order.Delivery.Address, 
		order.Delivery.Region, order.Delivery.Email,
	)
	if err != nil {
		return err
	}

	payment_uid := uuid.New().String()
	_, err = tx.Exec(`
		INSERT INTO payments (
			payment_uid, order_uid, transaction, request_id, currency, provider, 
			amount, payment_dt, bank, delivery_cost, goods_total, custom_fee
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`,
		payment_uid, order.OrderUID, order.Payment.Transaction, order.Payment.RequestID, 
		order.Payment.Currency, order.Payment.Provider, order.Payment.Amount, 
		order.Payment.PaymentDt, order.Payment.Bank, order.Payment.DeliveryCost, 
		order.Payment.GoodsTotal, order.Payment.CustomFee,
	)
	if err != nil {
		return err
	}

	for _, item := range order.Items {
		item_id := uuid.New().String()
		_, err := tx.Exec(`
			INSERT INTO items (
				item_uid, chrt_id, track_number, price, rid, name, sale, size, 
				total_price, nm_id, brand, status
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`,
			item_id, item.ChrtID, item.TrackNumber, item.Price, item.Rid, item.Name, 
			item.Sale, item.Size, item.TotalPrice, item.NmID, item.Brand, item.Status,
		)
		if err != nil {
			return err
		}

		_, err = tx.Exec(`
			INSERT INTO order_items (order_id, item_id) VALUES ($1, $2)`,
			order.OrderUID, item_id,
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}