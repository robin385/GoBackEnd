package models

import (
	"database/sql"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// User represents a user in the system
type User struct {
	ID           int       `json:"id" db:"id"`
	Name         string    `json:"name" db:"name"`
	Email        string    `json:"email" db:"email"`
	PasswordHash string    `json:"-" db:"password"`
	IsAdmin      bool      `json:"is_admin" db:"is_admin"`
	Exp          int       `json:"exp" db:"exp"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

// SetPassword hashes and sets the password for the user
func (u *User) SetPassword(pw string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(pw), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.PasswordHash = string(hash)
	return nil
}

// CheckPassword verifies the given password against the stored hash
func (u *User) CheckPassword(pw string) bool {
	return bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(pw)) == nil
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
               INSERT INTO users (name, email, password, is_admin)
               VALUES (?, ?, ?, ?)
               RETURNING id, exp, created_at, updated_at`

	err := r.db.QueryRow(query, user.Name, user.Email, user.PasswordHash, user.IsAdmin).Scan(
		&user.ID, &user.Exp, &user.CreatedAt, &user.UpdatedAt)
	return err
}

// GetByID retrieves a user by ID
func (r *UserRepository) GetByID(id int) (*User, error) {
	user := &User{}
	query := `SELECT id, name, email, password, is_admin, exp, created_at, updated_at FROM users WHERE id = ?`

	err := r.db.QueryRow(query, id).Scan(
		&user.ID, &user.Name, &user.Email, &user.PasswordHash, &user.IsAdmin, &user.Exp, &user.CreatedAt, &user.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	return user, err
}

// GetByEmail retrieves a user by email
func (r *UserRepository) GetByEmail(email string) (*User, error) {
	user := &User{}
	query := `SELECT id, name, email, password, is_admin, exp, created_at, updated_at FROM users WHERE email = ?`

	err := r.db.QueryRow(query, email).Scan(
		&user.ID, &user.Name, &user.Email, &user.PasswordHash, &user.IsAdmin, &user.Exp, &user.CreatedAt, &user.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	return user, err
}

// GetAll retrieves all users
func (r *UserRepository) GetAll() ([]*User, error) {
	query := `SELECT id, name, email, password, is_admin, exp, created_at, updated_at FROM users ORDER BY created_at DESC`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*User
	for rows.Next() {
		user := &User{}
		err := rows.Scan(&user.ID, &user.Name, &user.Email, &user.PasswordHash, &user.IsAdmin, &user.Exp, &user.CreatedAt, &user.UpdatedAt)
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
               SET name = ?, email = ?, password = ?, is_admin = ?, updated_at = CURRENT_TIMESTAMP
               WHERE id = ?`

	_, err := r.db.Exec(query, user.Name, user.Email, user.PasswordHash, user.IsAdmin, user.ID)
	return err
}

// Delete deletes a user
func (r *UserRepository) Delete(id int) error {
	query := `DELETE FROM users WHERE id = ?`
	_, err := r.db.Exec(query, id)
	return err
}

// AddExp increments a user's experience points
func (r *UserRepository) AddExp(userID, amount int) error {
	_, err := r.db.Exec(`UPDATE users SET exp = exp + ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`, amount, userID)
	return err
}

// GetTopByExp returns users ordered by experience descending limited by count
func (r *UserRepository) GetTopByExp(limit int) ([]*User, error) {
	query := `SELECT id, name, email, password, is_admin, exp, created_at, updated_at FROM users ORDER BY exp DESC LIMIT ?`
	rows, err := r.db.Query(query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*User
	for rows.Next() {
		u := &User{}
		if err := rows.Scan(&u.ID, &u.Name, &u.Email, &u.PasswordHash, &u.IsAdmin, &u.Exp, &u.CreatedAt, &u.UpdatedAt); err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, nil
}

// GetRank returns the ranking (1-based) and exp for a user by id
func (r *UserRepository) GetRank(userID int) (int, int, error) {
	var exp int
	if err := r.db.QueryRow(`SELECT exp FROM users WHERE id = ?`, userID).Scan(&exp); err != nil {
		if err == sql.ErrNoRows {
			return 0, 0, nil
		}
		return 0, 0, err
	}
	var rank int
	if err := r.db.QueryRow(`SELECT COUNT(*) + 1 FROM users WHERE exp > ?`, exp).Scan(&rank); err != nil {
		return 0, 0, err
	}
	return rank, exp, nil
}
