CREATE TABLE profiles (
    id VARCHAR(36)  PRIMARY KEY, -- UUID
    title VARCHAR(200) NOT NULL,
    bio TEXT NOT NULL,
    github_url TEXT NOT NULL,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL
);
