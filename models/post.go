package models

import (
	"database/sql"
)

// PostRepository handles post database operations
type PostRepository struct {
	db *sql.DB
}

// NewPostRepository creates a new post repository
func NewPostRepository(db *sql.DB) *PostRepository {
	return &PostRepository{db: db}
}

// Create creates a new post
func (r *PostRepository) Create(post *Post) error {
	query := `
		INSERT INTO posts (title, content, user_id) 
		VALUES (?, ?, ?) 
		RETURNING id, created_at, updated_at`
	
	err := r.db.QueryRow(query, post.Title, post.Content, post.UserID).Scan(
		&post.ID, &post.CreatedAt, &post.UpdatedAt)
	return err
}

// GetByID retrieves a post by ID with user information
func (r *PostRepository) GetByID(id int) (*Post, error) {
	post := &Post{}
	user := &User{}
	
	query := `
		SELECT p.id, p.title, p.content, p.user_id, p.created_at, p.updated_at,
		       u.id, u.name, u.email, u.created_at, u.updated_at
		FROM posts p
		JOIN users u ON p.user_id = u.id
		WHERE p.id = ?`
	
	err := r.db.QueryRow(query, id).Scan(
		&post.ID, &post.Title, &post.Content, &post.UserID, &post.CreatedAt, &post.UpdatedAt,
		&user.ID, &user.Name, &user.Email, &user.CreatedAt, &user.UpdatedAt)
	
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	
	post.User = user
	return post, nil
}

// GetAll retrieves all posts with user information
func (r *PostRepository) GetAll() ([]*Post, error) {
	query := `
		SELECT p.id, p.title, p.content, p.user_id, p.created_at, p.updated_at,
		       u.id, u.name, u.email, u.created_at, u.updated_at
		FROM posts p
		JOIN users u ON p.user_id = u.id
		ORDER BY p.created_at DESC`
	
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []*Post
	for rows.Next() {
		post := &Post{}
		user := &User{}
		
		err := rows.Scan(
			&post.ID, &post.Title, &post.Content, &post.UserID, &post.CreatedAt, &post.UpdatedAt,
			&user.ID, &user.Name, &user.Email, &user.CreatedAt, &user.UpdatedAt)
		if err != nil {
			return nil, err
		}
		
		post.User = user
		posts = append(posts, post)
	}
	return posts, nil
}

// GetByUserID retrieves all posts by a specific user
func (r *PostRepository) GetByUserID(userID int) ([]*Post, error) {
	query := `
		SELECT p.id, p.title, p.content, p.user_id, p.created_at, p.updated_at,
		       u.id, u.name, u.email, u.created_at, u.updated_at
		FROM posts p
		JOIN users u ON p.user_id = u.id
		WHERE p.user_id = ?
		ORDER BY p.created_at DESC`
	
	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []*Post
	for rows.Next() {
		post := &Post{}
		user := &User{}
		
		err := rows.Scan(
			&post.ID, &post.Title, &post.Content, &post.UserID, &post.CreatedAt, &post.UpdatedAt,
			&user.ID, &user.Name, &user.Email, &user.CreatedAt, &user.UpdatedAt)
		if err != nil {
			return nil, err
		}
		
		post.User = user
		posts = append(posts, post)
	}
	return posts, nil
}

// Update updates a post
func (r *PostRepository) Update(post *Post) error {
	query := `
		UPDATE posts 
		SET title = ?, content = ?, updated_at = CURRENT_TIMESTAMP 
		WHERE id = ?`
	
	_, err := r.db.Exec(query, post.Title, post.Content, post.ID)
	return err
}

// Delete deletes a post
func (r *PostRepository) Delete(id int) error {
	query := `DELETE FROM posts WHERE id = ?`
	_, err := r.db.Exec(query, id)
	return err
}
