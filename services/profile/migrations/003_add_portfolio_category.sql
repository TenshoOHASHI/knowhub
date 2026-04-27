ALTER TABLE portfolio_items ADD COLUMN category VARCHAR(100) NOT NULL DEFAULT 'project';
ALTER TABLE portfolio_items ADD COLUMN tech_stack VARCHAR(2000) NOT NULL DEFAULT '';
