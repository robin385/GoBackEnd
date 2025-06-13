package models

import (
	"database/sql"
	"time"
)

// User represents a user in the system
type User struct {
	ID        int       `json:"id" db:"id"`
	Name      string    `json:"name" db:"name"`
	Email     string    `json:"email" db:"email"`
	IsAdmin   bool      `json:"is_admin" db:"is_admin"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// Post represents a post in the system
type Post struct {
	ID        int       `json:"id" db:"id"`
	Title     string    `json:"title" db:"title"`
	Content   string    `json:"content" db:"content"`
	UserID    int       `json:"user_id" db:"user_id"`
	User      *User     `json:"user,omitempty"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// UserRepository handles user database operations
type UserRepository struct {
	db *sql.DB
}

// NewUserRepository creates a new user repository
func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

// Create creates a new user
func (r *UserRepository) Create(user *User) error {
	query := `
                INSERT INTO users (name, email, is_admin)
                VALUES (?, ?, ?)
                RETURNING id, created_at, updated_at`

	err := r.db.QueryRow(query, user.Name, user.Email, user.IsAdmin).Scan(
		&user.ID, &user.CreatedAt, &user.UpdatedAt)
	return err
}

// GetByID retrieves a user by ID
func (r *UserRepository) GetByID(id int) (*User, error) {
	user := &User{}
	query := `SELECT id, name, email, is_admin, created_at, updated_at FROM users WHERE id = ?`

	err := r.db.QueryRow(query, id).Scan(
		&user.ID, &user.Name, &user.Email, &user.IsAdmin, &user.CreatedAt, &user.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	return user, err
}

// GetAll retrieves all users
func (r *UserRepository) GetAll() ([]*User, error) {
	query := `SELECT id, name, email, is_admin, created_at, updated_at FROM users ORDER BY created_at DESC`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*User
	for rows.Next() {
		user := &User{}
		err := rows.Scan(&user.ID, &user.Name, &user.Email, &user.IsAdmin, &user.CreatedAt, &user.UpdatedAt)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	return users, nil
}

// Update updates a user
func (r *UserRepository) Update(user *User) error {
	query := `
                UPDATE users
                SET name = ?, email = ?, is_admin = ?, updated_at = CURRENT_TIMESTAMP
                WHERE id = ?`

	_, err := r.db.Exec(query, user.Name, user.Email, user.IsAdmin, user.ID)
	return err
}

// Delete deletes a user
func (r *UserRepository) Delete(id int) error {
	query := `DELETE FROM users WHERE id = ?`
	_, err := r.db.Exec(query, id)
	return err
}
