CREATE TABLE users (
   id BIGSERIAL PRIMARY KEY
);

CREATE TABLE models (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT now(),

    UNIQUE (user_id, name)
);
