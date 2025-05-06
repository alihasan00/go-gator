-- +goose Up 
CREATE TABLE feeds(
    ID UUID PRIMARY KEY,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    name TEXT,
    url Text UNIQUE,
    user_id UUID,
    FOREIGN KEY (user_id) REFERENCES users(ID) ON DELETE CASCADE
);

-- +goose Down
DROP TABLE feeds;