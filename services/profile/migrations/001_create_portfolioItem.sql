CREATE TABLE portfolio_items (
    id VARCHAR(36) PRIMARY KEY,
    title VARCHAR(200) NOT NULL,
    description TEXT NOT NULL,
    url TEXT NOT NULL,
    status TEXT NOT NULL,
    created_at DATETIME NOT NULL
);
