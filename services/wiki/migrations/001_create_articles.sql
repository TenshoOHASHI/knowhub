CREATE TABLE articles (
    id VARCHAR(36)  PRIMARY KEY, -- UUID
    title VARCHAR(200) NOT NULL,
    content TEXT NOT NULL,
    updated_at DATETIME NOT NULL
);
