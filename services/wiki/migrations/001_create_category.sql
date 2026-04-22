CREATE TABLE categories (
      id VARCHAR(36) PRIMARY KEY,
      name VARCHAR(100) NOT NULL,
      parent_id VARCHAR(36) NOT NULL DEFAULT ''
  );

--
