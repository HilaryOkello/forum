-- USERS Table
CREATE TABLE IF NOT EXISTS users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,      
    email TEXT UNIQUE NOT NULL,                
    username TEXT UNIQUE NOT NULL,             
    password TEXT NOT NULL,              
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP 
);

-- POSTS Table
CREATE TABLE IF NOT EXISTS posts (
    id INTEGER PRIMARY KEY AUTOINCREMENT,     
    user_id INTEGER NOT NULL,                 
    title TEXT NOT NULL,                      
    content TEXT NOT NULL,                    
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) 
);

-- COMMENTS Table
CREATE TABLE IF NOT EXISTS comments (
    id INTEGER PRIMARY KEY AUTOINCREMENT,     
    post_id INTEGER NOT NULL,                 
    user_id INTEGER NOT NULL, 
    parent_id INTEGER DEFAULT NULL,                
    content TEXT NOT NULL,                    
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (post_id) REFERENCES posts(id),
    FOREIGN KEY (user_id) REFERENCES users(id)
    FOREIGN KEY (parent_id) REFERENCES comments(id)
);

-- CATEGORIES Table
CREATE TABLE IF NOT EXISTS categories (
    id INTEGER PRIMARY KEY AUTOINCREMENT,     
    name TEXT UNIQUE NOT NULL,                
    description TEXT                          
);

-- POST_CATEGORIES Table
CREATE TABLE IF NOT EXISTS post_categories (
    post_id INTEGER NOT NULL,                  
    category_id INTEGER NOT NULL,              
    PRIMARY KEY (post_id, category_id),
    FOREIGN KEY (post_id) REFERENCES posts(id),
    FOREIGN KEY (category_id) REFERENCES categories(id)
);

-- POST_REACTIONS Table
CREATE TABLE IF NOT EXISTS post_reactions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,     
    post_id INTEGER NOT NULL,                 
    user_id INTEGER NOT NULL,                 
    reaction_type TEXT CHECK (reaction_type IN ('LIKE', 'DISLIKE')) NOT NULL,
    FOREIGN KEY (post_id) REFERENCES posts(id),
    FOREIGN KEY (user_id) REFERENCES users(id) 
);

-- COMMENT_REACTIONS Table
CREATE TABLE IF NOT EXISTS comment_reactions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,     
    comment_id INTEGER NOT NULL,              
    user_id INTEGER NOT NULL,                 
    reaction_type TEXT CHECK (reaction_type IN ('LIKE', 'DISLIKE')) NOT NULL,
    FOREIGN KEY (comment_id) REFERENCES comments(id),
    FOREIGN KEY (user_id) REFERENCES users(id) 
    UNIQUE (comment_id, user_id)
);

-- SESSIONS Table
CREATE TABLE IF NOT EXISTS sessions (
    uuid TEXT PRIMARY KEY,                    
    user_id INTEGER NOT NULL,                 
    expires_at DATETIME NOT NULL,             
    FOREIGN KEY (user_id) REFERENCES users(id)
);
