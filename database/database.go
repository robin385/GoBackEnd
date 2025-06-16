package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
)

type DB struct {
	*sql.DB
}

// Initialize creates and returns a new database connection
func Initialize(dbPath string) (*DB, error) {
	// Create directory if it doesn't exist
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create database directory: %w", err)
	}

	// Open database connection
	db, err := sql.Open("sqlite3", dbPath+"?_foreign_keys=on")
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Test connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	database := &DB{db}

	// Run migrations
	if err := database.migrate(); err != nil {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	log.Println("Database initialized successfully")
	return database, nil
}

// migrate runs database migrations
func (db *DB) migrate() error {
	// Create users table
	createUsersTable := `
        CREATE TABLE IF NOT EXISTS users (
                id INTEGER PRIMARY KEY AUTOINCREMENT,
                name TEXT NOT NULL,
                email TEXT UNIQUE NOT NULL,
                password TEXT NOT NULL,
                is_admin BOOLEAN NOT NULL DEFAULT 0,
                exp INTEGER NOT NULL DEFAULT 0,
                created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
                updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
        );`

	if _, err := db.Exec(createUsersTable); err != nil {
		return fmt.Errorf("failed to create users table: %w", err)
	}

	// Ensure password column exists for old installations
	var col string
	err := db.QueryRow("SELECT name FROM pragma_table_info('users') WHERE name='password'").Scan(&col)
	if err == sql.ErrNoRows {
		if _, err := db.Exec("ALTER TABLE users ADD COLUMN password TEXT NOT NULL DEFAULT ''"); err != nil {
			return fmt.Errorf("failed to add password column: %w", err)
		}
	} else if err != nil {
		return fmt.Errorf("failed to check password column: %w", err)
	}

	// Ensure exp column exists for old installations
	err = db.QueryRow("SELECT name FROM pragma_table_info('users') WHERE name='exp'").Scan(&col)
	if err == sql.ErrNoRows {
		if _, err := db.Exec("ALTER TABLE users ADD COLUMN exp INTEGER NOT NULL DEFAULT 0"); err != nil {
			return fmt.Errorf("failed to add exp column: %w", err)
		}
	} else if err != nil {
		return fmt.Errorf("failed to check exp column: %w", err)
	}

	// Create trash_posts table
	createTrashTable := `
       CREATE TABLE IF NOT EXISTS trash_posts (
               id INTEGER PRIMARY KEY AUTOINCREMENT,
               user_id INTEGER NOT NULL,
               latitude REAL NOT NULL,
               longitude REAL NOT NULL,
               image_path TEXT,
               description TEXT NOT NULL,
               trail TEXT,
               created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
               FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
       );`

	if _, err := db.Exec(createTrashTable); err != nil {
		return fmt.Errorf("failed to create trash_posts table: %w", err)
	}

	// Create comments table
	createCommentsTable := `
       CREATE TABLE IF NOT EXISTS comments (
               id INTEGER PRIMARY KEY AUTOINCREMENT,
               post_id INTEGER NOT NULL,
               user_id INTEGER NOT NULL,
               content TEXT NOT NULL,
               created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
               FOREIGN KEY (post_id) REFERENCES trash_posts(id) ON DELETE CASCADE,
               FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
       );`

	if _, err := db.Exec(createCommentsTable); err != nil {
		return fmt.Errorf("failed to create comments table: %w", err)
	}

	// Create indexes
	createIndexes := []string{
		"CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);",
		"CREATE INDEX IF NOT EXISTS idx_trash_user_id ON trash_posts(user_id);",
		"CREATE INDEX IF NOT EXISTS idx_trash_created_at ON trash_posts(created_at);",
		"CREATE INDEX IF NOT EXISTS idx_comments_post_id ON comments(post_id);",
		"CREATE INDEX IF NOT EXISTS idx_comments_created_at ON comments(created_at);",
	}

	for _, query := range createIndexes {
		if _, err := db.Exec(query); err != nil {
			return fmt.Errorf("failed to create index: %w", err)
		}
	}

	log.Println("Database migrations completed successfully")
	return nil
}
