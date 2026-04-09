-- +goose Up
CREATE TABLE stephanie-ce (
    id INTEGER PRIMARY KEY,
    test VARCHAR(100)
);

-- +goose Down
DELETE FROM stephanie-ce;
