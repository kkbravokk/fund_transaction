CREATE TABLE `transaction` (
    id INTEGER PRIMARY KEY NOT NULL,
    original_buy_id INTEGER NOT NULL,
    fund_code TEXT NOT NULL,
    transaction_type TEXT NOT NULL,
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

CREATE TABLE `fund` (
    code TEXT PRIMARY KEY NOT NULL,
    name TEXT NOT NULL
);