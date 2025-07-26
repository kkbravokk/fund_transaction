CREATE TABLE `transaction` (
    id INTEGER PRIMARY KEY NOT NULL,
    buy_id INTEGER NOT NULL,
    unit REAL NOT NULL,
    amount INTEGER NOT NULL,
    price REAL NOT NULL,
    load REAL NOT NULL,
    left_amount INTEGER NOT NULL,
    profit REAL NOT NULL,
    profit_margin REAL NOT NULL,
    net_profit REAL NOT NULL,
    created_at DATETIME NOT NULL
);