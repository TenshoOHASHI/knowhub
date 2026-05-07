-- +goose Up

CREATE TABLE IF NOT EXISTS users (
  id VARCHAR(36) PRIMARY KEY,
  username VARCHAR(100) NOT NULL,
  email VARCHAR(200) NOT NULL UNIQUE,
  password_hash VARCHAR(200) NOT NULL,
  created_at DATETIME NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS categories (
  id VARCHAR(36) PRIMARY KEY,
  name VARCHAR(100) NOT NULL,
  parent_id VARCHAR(36) NOT NULL DEFAULT ''
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS articles (
  id VARCHAR(36) PRIMARY KEY,
  title VARCHAR(200) NOT NULL,
  content TEXT NOT NULL,
  category_id VARCHAR(36) NOT NULL DEFAULT '',
  visibility VARCHAR(20) NOT NULL DEFAULT 'public',
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  INDEX idx_articles_created_at (created_at),
  INDEX idx_articles_visibility (visibility)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS article_likes (
  id VARCHAR(36) PRIMARY KEY,
  article_id VARCHAR(36) NOT NULL,
  fingerprint VARCHAR(64) NOT NULL,
  created_at DATETIME NOT NULL,
  UNIQUE KEY uk_article_fingerprint (article_id, fingerprint),
  INDEX idx_article_id (article_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS saved_articles (
  id VARCHAR(36) PRIMARY KEY,
  article_id VARCHAR(36) NOT NULL,
  fingerprint VARCHAR(64) NOT NULL,
  created_at DATETIME NOT NULL,
  UNIQUE KEY uk_saved_article_fingerprint (article_id, fingerprint),
  INDEX idx_fingerprint (fingerprint)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS page_views (
  id VARCHAR(36) PRIMARY KEY,
  path VARCHAR(500) NOT NULL,
  ip_hash VARCHAR(64) NOT NULL,
  user_agent VARCHAR(500),
  referrer VARCHAR(500),
  visited_at DATETIME NOT NULL,
  INDEX idx_visited_at (visited_at),
  INDEX idx_path (path)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS profiles (
  id VARCHAR(36) PRIMARY KEY,
  title VARCHAR(200) NOT NULL,
  bio TEXT NOT NULL,
  github_url TEXT NOT NULL,
  avatar_url VARCHAR(2000) NOT NULL DEFAULT '',
  twitter_url VARCHAR(2000) NOT NULL DEFAULT '',
  linkedin_url VARCHAR(2000) NOT NULL DEFAULT '',
  wantedly_url VARCHAR(2000) NOT NULL DEFAULT '',
  skills VARCHAR(2000) NOT NULL DEFAULT '',
  languages VARCHAR(2000) NOT NULL DEFAULT '',
  created_at DATETIME NOT NULL,
  updated_at DATETIME NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS portfolio_items (
  id VARCHAR(36) PRIMARY KEY,
  title VARCHAR(200) NOT NULL,
  description TEXT NOT NULL,
  url TEXT NOT NULL,
  status TEXT NOT NULL,
  category VARCHAR(100) NOT NULL DEFAULT 'project',
  tech_stack VARCHAR(2000) NOT NULL DEFAULT '',
  created_at DATETIME NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- +goose Down

DROP TABLE IF EXISTS portfolio_items;
DROP TABLE IF EXISTS profiles;
DROP TABLE IF EXISTS page_views;
DROP TABLE IF EXISTS saved_articles;
DROP TABLE IF EXISTS article_likes;
DROP TABLE IF EXISTS articles;
DROP TABLE IF EXISTS categories;
DROP TABLE IF EXISTS users;
