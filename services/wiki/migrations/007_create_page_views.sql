-- Page views for analytics
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
