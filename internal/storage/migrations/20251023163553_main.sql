-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS timers (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    external_id TEXT UNIQUE,
    fixed_at DATE not null, 
    seconds_spent int not null,
    "name" TEXT NOT NULL DEFAULT '',
    "description" TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

create table if not EXISTS shopping_list_items (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    description TEXT,
    medal_type TEXT, 
    medal_count INT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

create table if not EXISTS bought_items (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    shopping_list_item_id INTEGER NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (shopping_list_item_id) REFERENCES shopping_list_items(id)
);

create table if not EXISTS wallet (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    count int,
    medal_type TEXT UNIQUE
);

CREATE TABLE IF NOT EXISTS rewards_daily (
    day DATE NOT NULL,
    medal_type TEXT NOT NULL,
    count INT NOT NULL,
    PRIMARY KEY(day, medal_type)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS bought_items;
DROP TABLE IF EXISTS shopping_list_items;
DROP TABLE IF EXISTS timers;
DROP TABLE IF EXISTS wallet;
DROP TABLE IF EXISTS rewards_daily;
-- +goose StatementEnd
