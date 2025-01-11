-- USERS Table
CREATE TABLE IF NOT EXISTS users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,      
    email TEXT UNIQUE NOT NULL,                
    username TEXT UNIQUE NOT NULL,             
    password_hash TEXT NOT NULL,              
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP 
);


