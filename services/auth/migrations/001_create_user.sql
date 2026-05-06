CREATE TABLE users (
      id VARCHAR(36) PRIMARY KEY,
      username VARCHAR(100) NOT NULL,
      email VARCHAR(200) NOT NULL UNIQUE,
      password_hash VARCHAR(200) NOT NULL,
      created_at DATETIME NOT NULL
  );
