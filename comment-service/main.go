// Comment Service (main.go)
package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)

// Comment represents a comment on a blog post
type Comment struct {
	ID        int       `json:"id"`
	PostID    int       `json:"post_id"`
	UserID    int       `json:"user_id"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

var db *sql.DB

func main() {
	// Set up database connection using environment variables
	dbUser := getEnv("DB_USER", "")
	dbPassword := getEnv("DB_PASSWORD", "")
	dbHost := getEnv("DB_HOST", "")
	dbPort := getEnv("DB_PORT", "")
	dbName := getEnv("DB_NAME", "")

	connectionString := fmt.Sprintf("postgresql://%s:%s@%s:%s/%s?sslmode=require",
		dbUser, dbPassword, dbHost, dbPort, dbName)

	var err error
	db, err = sql.Open("postgres", connectionString)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Connected to database successfully!")

	// Set up API routes
	r := mux.NewRouter()

	r.Use(corsMiddleware)

	r.PathPrefix("/").Methods("OPTIONS").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	r.HandleFunc("/health", healthCheck).Methods("GET")
	r.HandleFunc("/posts/{post_id:[0-9]+}/comments", createComment).Methods("POST")
	r.HandleFunc("/posts/{post_id:[0-9]+}/comments", getComments).Methods("GET")
	r.HandleFunc("/comments/{id:[0-9]+}", updateComment).Methods("PUT")
	r.HandleFunc("/comments/{id:[0-9]+}", deleteComment).Methods("DELETE")

	// Start server
	port := getEnv("PORT", "8083")
	log.Printf("Comment service starting on port %s...", port)
	log.Fatal(http.ListenAndServe(":"+port, r))
}

// corsMiddleware handles CORS for all routes
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set CORS headers
		w.Header().Set("Access-Control-Allow-Origin", "*") // Allow all origins (change this in production)
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		// Handle preflight OPTIONS requests
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		// Pass to the next handler
		next.ServeHTTP(w, r)
	})
}

// getEnv gets an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// Simple healthcheck handler for K8s probes
func healthCheck(w http.ResponseWriter, r *http.Request) {
	// Basic health check that always returns 200 OK for liveness probe
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

// createComment adds a new comment to a post
func createComment(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	postID := vars["post_id"]

	var comment Comment
	if err := json.NewDecoder(r.Body).Decode(&comment); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Set the post ID from the URL
	fmt.Sscanf(postID, "%d", &comment.PostID)

	// Simple validation
	if comment.Content == "" || comment.UserID == 0 {
		http.Error(w, "Content and user_id are required", http.StatusBadRequest)
		return
	}

	// Check if the post exists
	var exists bool
	err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM posts WHERE id = $1)", comment.PostID).Scan(&exists)
	if err != nil {
		http.Error(w, "Error checking post existence: "+err.Error(), http.StatusInternalServerError)
		return
	}
	if !exists {
		http.Error(w, "Post not found", http.StatusNotFound)
		return
	}

	// Insert the new comment
	err = db.QueryRow(
		"INSERT INTO comments (post_id, user_id, content) VALUES ($1, $2, $3) RETURNING id, created_at, updated_at",
		comment.PostID, comment.UserID, comment.Content,
	).Scan(&comment.ID, &comment.CreatedAt, &comment.UpdatedAt)
	if err != nil {
		http.Error(w, "Error creating comment: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(comment)
}

// getComments returns all comments for a specific post
func getComments(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	postID := vars["post_id"]

	// Check if the post exists
	var exists bool
	err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM posts WHERE id = $1)", postID).Scan(&exists)
	if err != nil {
		http.Error(w, "Error checking post existence: "+err.Error(), http.StatusInternalServerError)
		return
	}
	if !exists {
		http.Error(w, "Post not found", http.StatusNotFound)
		return
	}

	rows, err := db.Query(
		"SELECT id, post_id, user_id, content, created_at, updated_at FROM comments WHERE post_id = $1 ORDER BY created_at ASC",
		postID,
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	comments := []Comment{}
	for rows.Next() {
		var comment Comment
		if err := rows.Scan(&comment.ID, &comment.PostID, &comment.UserID, &comment.Content, &comment.CreatedAt, &comment.UpdatedAt); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		comments = append(comments, comment)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(comments)
}

// updateComment updates an existing comment
func updateComment(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	var comment Comment
	if err := json.NewDecoder(r.Body).Decode(&comment); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Simple validation
	if comment.Content == "" {
		http.Error(w, "Content is required", http.StatusBadRequest)
		return
	}

	// Update the comment
	_, err := db.Exec(
		"UPDATE comments SET content = $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2",
		comment.Content, id,
	)
	if err != nil {
		http.Error(w, "Error updating comment: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Get the updated comment
	err = db.QueryRow(
		"SELECT id, post_id, user_id, content, created_at, updated_at FROM comments WHERE id = $1",
		id,
	).Scan(&comment.ID, &comment.PostID, &comment.UserID, &comment.Content, &comment.CreatedAt, &comment.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Comment not found", http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(comment)
}

// deleteComment removes a comment
func deleteComment(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	// Delete the comment
	result, err := db.Exec("DELETE FROM comments WHERE id = $1", id)
	if err != nil {
		http.Error(w, "Error deleting comment: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Check if a row was actually deleted
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if rowsAffected == 0 {
		http.Error(w, "Comment not found", http.StatusNotFound)
		return
	}

	// Return a success message
	w.WriteHeader(http.StatusNoContent)
}
