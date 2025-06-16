package models

import (
	"database/sql"
	"time"
)

// Comment represents a comment on a trash post
type Comment struct {
	ID        int       `json:"id" db:"id"`
	PostID    int       `json:"post_id" db:"post_id"`
	UserID    int       `json:"user_id" db:"user_id"`
	User      *User     `json:"user,omitempty"`
	Content   string    `json:"content" db:"content"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// CommentRepository handles comment database operations
type CommentRepository struct {
	db *sql.DB
}

// NewCommentRepository creates a new repository
func NewCommentRepository(db *sql.DB) *CommentRepository {
	return &CommentRepository{db: db}
}

// Create inserts a new comment into the database
func (r *CommentRepository) Create(c *Comment) error {
	query := `
        INSERT INTO comments (post_id, user_id, content)
        VALUES (?, ?, ?)
        RETURNING id, created_at`
	return r.db.QueryRow(query, c.PostID, c.UserID, c.Content).Scan(&c.ID, &c.CreatedAt)
}

// GetByPostID retrieves all comments for a given post
func (r *CommentRepository) GetByPostID(postID int) ([]*Comment, error) {
	query := `
        SELECT c.id, c.post_id, c.user_id, c.content, c.created_at,
               u.id, u.name, u.email, u.is_admin, u.created_at, u.updated_at
        FROM comments c
        JOIN users u ON c.user_id = u.id
        WHERE c.post_id = ?
        ORDER BY c.created_at ASC`
	rows, err := r.db.Query(query, postID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var comments []*Comment
	for rows.Next() {
		c := &Comment{}
		u := &User{}
		if err := rows.Scan(&c.ID, &c.PostID, &c.UserID, &c.Content, &c.CreatedAt,
			&u.ID, &u.Name, &u.Email, &u.IsAdmin, &u.CreatedAt, &u.UpdatedAt); err != nil {
			return nil, err
		}
		c.User = u
		comments = append(comments, c)
	}
	return comments, nil
}
