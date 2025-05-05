-- init.sql: This script initializes the database schema

-- Create database if it doesn't exist
CREATE DATABASE IF NOT EXISTS blogdb;

-- Connect to the newly created database
\c blogdb;

-- Create tables
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(50) UNIQUE NOT NULL,
    email VARCHAR(100) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE posts (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
    title VARCHAR(255) NOT NULL,
    content TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE comments (
    id SERIAL PRIMARY KEY,
    post_id INTEGER REFERENCES posts(id) ON DELETE CASCADE,
    user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
    content TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for better performance
CREATE INDEX idx_posts_user_id ON posts(user_id);
CREATE INDEX idx_comments_post_id ON comments(post_id);
CREATE INDEX idx_comments_user_id ON comments(user_id);

-- Insert some sample data
INSERT INTO users (username, email, password_hash) VALUES 
('john_doe', 'john@example.com', '$2a$10$1qAz2wSx3eDc4rFv5tGb5edDmJnZczZJHlfKcHKxZ.sU9IMFkxmLK'), -- password: password123
('jane_smith', 'jane@example.com', '$2a$10$1qAz2wSx3eDc4rFv5tGb5edDmJnZczZJHlfKcHKxZ.sU9IMFkxmLK');

INSERT INTO posts (user_id, title, content) VALUES 
(1, 'First Post', 'This is my first blog post. Welcome to my blog!'),
(2, 'Hello World', 'Hello everyone! This is my introduction post.');

INSERT INTO comments (post_id, user_id, content) VALUES 
(1, 2, 'Great first post!'),
(2, 1, 'Welcome to the blogging world!');