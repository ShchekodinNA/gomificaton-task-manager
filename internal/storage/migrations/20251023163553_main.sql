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

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS bought_items;
DROP TABLE IF EXISTS shopping_list_items;
DROP TABLE IF EXISTS timers;
-- +goose StatementEnd
