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
    payment_dt INT,
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
    'b563feb7b2b84b6test', 'WBILMTESTTRACK', 'WBIL', 'en', '', 'test',
    'meest', '9', 99, '2021-11-26T06:22:19Z', '1'
);
INSERT INTO deliveries (
    delivery_uid, order_uid, name, phone, zip, city, address, region, email
) VALUES (
    'delivery_uid_1','b563feb7b2b84b6test', 'Test Testov', '+9720000000', '2639809', 'Kiryat Mozkin',
    'Ploshad Mira 15', 'Kraiot', 'test@gmail.com'
);
INSERT INTO payments (
    payment_uid, order_uid, transaction, request_id, currency, provider, amount,
    payment_dt, bank, delivery_cost, goods_total, custom_fee
) VALUES (
    'payment_uid_1', 'b563feb7b2b84b6test', 'b563feb7b2b84b6test', '', 'USD', 'wbpay', 1817,
    1637907727, 'Сбербанк', 1500, 317, 0
);
INSERT INTO items (
    item_uid, chrt_id, track_number, price, rid, name, sale,
    size, total_price, nm_id, brand, status
) VALUES (
    'item_uid_1', 9934930, 'WBILMTESTTRACK', 453, 'ab4219087a764ae0btest', 'Mascaras', 30,
    '0', 317, 2389212, 'Vivienne Sabo', 202
);
INSERT INTO order_items (order_id, item_id)
VALUES ('b563feb7b2b84b6test', 'item_uid_1');


-- +goose Down 
DROP TABLE orders;
DROP TABLE  deliveries;
DROP TABLE  payments;
DROP TABLE  order_items;