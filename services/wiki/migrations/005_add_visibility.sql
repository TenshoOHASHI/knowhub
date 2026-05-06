ALTER TABLE articles ADD COLUMN visibility VARCHAR(20) NOT NULL DEFAULT 'public' AFTER category_id;
