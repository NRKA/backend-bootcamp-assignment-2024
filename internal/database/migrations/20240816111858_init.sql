-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS "user" (
    id BIGSERIAL PRIMARY KEY NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    user_type VARCHAR(10) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS house (
    id BIGSERIAL PRIMARY KEY NOT NULL,
    address VARCHAR(255) NOT NULL,
    year INT NOT NULL,
    developer VARCHAR(255),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS flat (
    id BIGSERIAL PRIMARY KEY NOT NULL,
    number INT NOT NULL,
    house_id INT NOT NULL,
    price INT NOT NULL,
    rooms INT NOT NULL,
    status VARCHAR(255) DEFAULT 'created',
    UNIQUE (number, house_id),
    CONSTRAINT fk_house
        FOREIGN KEY (house_id)
        REFERENCES house(id)
        ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS subscriber (
    id BIGSERIAL PRIMARY KEY NOT NULL,
    email VARCHAR(255) NOT NULL,
    house_id INT NOT NULL,
    UNIQUE (house_id),
    CONSTRAINT fk_house
        FOREIGN KEY (house_id)
        REFERENCES house(id)
        ON DELETE CASCADE
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS subscriber;
DROP TABLE IF EXISTS flat;
DROP TABLE IF EXISTS house;
DROP TABLE IF EXISTS "user";
-- +goose StatementEnd
