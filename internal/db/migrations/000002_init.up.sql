-- +goose Up
-- Таблица заказов
CREATE TABLE IF NOT EXISTS orders (
    order_uid VARCHAR PRIMARY KEY,
    track_number TEXT,
    entry TEXT,
    locale TEXT,
    internal_signature TEXT,
    customer_id TEXT,
    delivery_service TEXT,
    shardkey TEXT,
    sm_id INT,
    date_created TIMESTAMP,
    oof_shard TEXT
);

-- Таблица доставок
CREATE TABLE IF NOT EXISTS deliveries (
    delivery_uid VARCHAR PRIMARY KEY,
    order_uid VARCHAR REFERENCES orders(order_uid) ON DELETE CASCADE,
    name TEXT,
    phone TEXT,
    zip TEXT,
    city TEXT,
    address TEXT,
    region TEXT,
    email TEXT
);

-- Таблица платежей
CREATE TABLE IF NOT EXISTS payments (
    payment_uid VARCHAR PRIMARY KEY,
    order_uid VARCHAR REFERENCES orders(order_uid) ON DELETE CASCADE,
    transaction TEXT,
    request_id TEXT,
    currency TEXT,
    provider TEXT,
    amount INT,
    payment_dt TIMESTAMP,
    bank TEXT,
    delivery_cost INT,
    goods_total INT,
    custom_fee INT
);

-- Таблица предметов
CREATE TABLE IF NOT EXISTS items (
    item_uid VARCHAR PRIMARY KEY,
    chrt_id INT,
    track_number TEXT,
    price INT,
    rid TEXT,
    name TEXT,
    sale INT,
    size TEXT,
    total_price INT,
    nm_id INT,
    brand TEXT,
    status INT
);

-- Промежуточная таблица связей
CREATE TABLE IF NOT EXISTS order_items (
    connection_uid SERIAL PRIMARY KEY,
    order_id VARCHAR REFERENCES orders(order_uid) ON DELETE CASCADE,
    item_id  VARCHAR REFERENCES items(item_uid) ON DELETE CASCADE
);

-- Тестовое значение
INSERT INTO orders (
    order_uid, track_number, entry, locale, internal_signature, customer_id,
    delivery_service, shardkey, sm_id, date_created, oof_shard
) VALUES (
    'b563feb7b2b84b6test', 'WBILMTESTTRACK', 'WBIL', 'en', '', 'test_customer',
    'meest', 'ab', 99, now(), '1'
);
INSERT INTO deliveries (
    delivery_uid, order_uid, name, phone, zip, city, address, region, email
) VALUES (
    'delivery1','b563feb7b2b84b6test', 'Иван Иванов', '+79991234567', '123456', 'Москва',
    'ул. Ленина, 1', 'Московская область', 'ivan@example.com'
);
INSERT INTO payments (
    payment_uid, order_uid, transaction, request_id, currency, provider, amount,
    payment_dt, bank, delivery_cost, goods_total, custom_fee
) VALUES (
    'payment1', 'b563feb7b2b84b6test', 'trans123', '', 'RUB', 'visa', 1000,
    now(), 'Сбербанк', 100, 900, 0
);
INSERT INTO items (
    item_uid, chrt_id, track_number, price, rid, name, sale,
    size, total_price, nm_id, brand, status
) VALUES (
    'item1', 123456, 'WBILMTESTTRACK', 500, 'rid123', 'Куртка', 10,
    'M', 450, 987654, 'Nike', 202
);
INSERT INTO order_items (order_id, item_id)
VALUES ('b563feb7b2b84b6test', 'item1');


-- +goose Down 
DROP TABLE orders;
DROP TABLE  deliveries;
DROP TABLE  payments;
DROP TABLE  order_items;