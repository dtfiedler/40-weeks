INSERT INTO users (username, password, email) 
SELECT 'admin', '$2a$10$N9qo8uLOickgx2ZMRZoMye.q8Qr4NHw0/tLnJOYJr5xJgOGdZY4o6', 'admin@example.com'
WHERE NOT EXISTS (SELECT 1 FROM users LIMIT 1);