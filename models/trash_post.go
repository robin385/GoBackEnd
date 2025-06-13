package models

import (
	"database/sql"
	"time"
)

// TrashPost represents a trash spot reported by a user
type TrashPost struct {
	ID          int       `json:"id" db:"id"`
	UserID      int       `json:"user_id" db:"user_id"`
	User        *User     `json:"user,omitempty"`
	Latitude    float64   `json:"latitude" db:"latitude"`
	Longitude   float64   `json:"longitude" db:"longitude"`
	ImagePath   string    `json:"image_path,omitempty" db:"image_path"`
	Description string    `json:"description" db:"description"`
	Trail       string    `json:"trail,omitempty" db:"trail"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

// TrashPostRepository handles trash post database operations
type TrashPostRepository struct {
	db *sql.DB
}

// NewTrashPostRepository creates a new repository
func NewTrashPostRepository(db *sql.DB) *TrashPostRepository {
	return &TrashPostRepository{db: db}
}

// Create inserts a new trash post
func (r *TrashPostRepository) Create(post *TrashPost) error {
	query := `
       INSERT INTO trash_posts (user_id, latitude, longitude, image_path, description, trail)
       VALUES (?, ?, ?, ?, ?, ?)
       RETURNING id, created_at`
	return r.db.QueryRow(query, post.UserID, post.Latitude, post.Longitude, post.ImagePath, post.Description, post.Trail).Scan(&post.ID, &post.CreatedAt)
}

// GetByDateRange returns all trash posts between start and end dates inclusive
func (r *TrashPostRepository) GetByDateRange(start, end time.Time) ([]*TrashPost, error) {
	query := `
       SELECT tp.id, tp.user_id, tp.latitude, tp.longitude, COALESCE(tp.image_path, ''), tp.description, COALESCE(tp.trail, ''), tp.created_at,
              u.id, u.name, u.email, u.is_admin, u.created_at, u.updated_at
       FROM trash_posts tp
       JOIN users u ON tp.user_id = u.id
       WHERE tp.created_at BETWEEN ? AND ?
       ORDER BY tp.created_at DESC`

	rows, err := r.db.Query(query, start, end)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []*TrashPost
	for rows.Next() {
		p := &TrashPost{}
		u := &User{}
		if err := rows.Scan(&p.ID, &p.UserID, &p.Latitude, &p.Longitude, &p.ImagePath, &p.Description, &p.Trail, &p.CreatedAt, &u.ID, &u.Name, &u.Email, &u.IsAdmin, &u.CreatedAt, &u.UpdatedAt); err != nil {
			return nil, err
		}
		p.User = u
		posts = append(posts, p)
	}
	return posts, nil
}

// Delete removes a trash post by id
func (r *TrashPostRepository) Delete(id int) error {
	_, err := r.db.Exec(`DELETE FROM trash_posts WHERE id = ?`, id)
	return err
}
