// Post Service (main.go)
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

// Post represents a blog post
type Post struct {
	ID        int       `json:"id"`
	UserID    int       `json:"user_id"`
	Title     string    `json:"title"`
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
	r.HandleFunc("/posts", createPost).Methods("POST")
	r.HandleFunc("/posts", getPosts).Methods("GET")
	r.HandleFunc("/posts/{id:[0-9]+}", getPost).Methods("GET")
	r.HandleFunc("/posts/{id:[0-9]+}", updatePost).Methods("PUT")
	r.HandleFunc("/posts/{id:[0-9]+}", deletePost).Methods("DELETE")

	// Start server
	port := getEnv("PORT", "8082")
	log.Printf("Post service starting on port %s...", port)
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

// createPost handles creating a new blog post
func createPost(w http.ResponseWriter, r *http.Request) {
	var post Post
	if err := json.NewDecoder(r.Body).Decode(&post); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Simple validation
	if post.Title == "" || post.Content == "" || post.UserID == 0 {
		http.Error(w, "Title, content, and user_id are required", http.StatusBadRequest)
		return
	}

	// Insert the new post
	err := db.QueryRow(
		"INSERT INTO posts (user_id, title, content) VALUES ($1, $2, $3) RETURNING id, created_at, updated_at",
		post.UserID, post.Title, post.Content,
	).Scan(&post.ID, &post.CreatedAt, &post.UpdatedAt)
	if err != nil {
		http.Error(w, "Error creating post: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(post)
}

// getPosts returns all blog posts, with optional pagination
func getPosts(w http.ResponseWriter, r *http.Request) {
	// In a real app, you'd implement pagination here
	rows, err := db.Query("SELECT id, user_id, title, content, created_at, updated_at FROM posts ORDER BY created_at DESC LIMIT 100")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	posts := []Post{}
	for rows.Next() {
		var post Post
		if err := rows.Scan(&post.ID, &post.UserID, &post.Title, &post.Content, &post.CreatedAt, &post.UpdatedAt); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		posts = append(posts, post)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(posts)
}

// getPost returns a specific blog post by ID
func getPost(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	var post Post
	err := db.QueryRow(
		"SELECT id, user_id, title, content, created_at, updated_at FROM posts WHERE id = $1",
		id,
	).Scan(&post.ID, &post.UserID, &post.Title, &post.Content, &post.CreatedAt, &post.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Post not found", http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(post)
}

// updatePost updates an existing blog post
func updatePost(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	var post Post
	if err := json.NewDecoder(r.Body).Decode(&post); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Simple validation
	if post.Title == "" || post.Content == "" {
		http.Error(w, "Title and content are required", http.StatusBadRequest)
		return
	}

	// Update the post
	_, err := db.Exec(
		"UPDATE posts SET title = $1, content = $2, updated_at = CURRENT_TIMESTAMP WHERE id = $3",
		post.Title, post.Content, id,
	)
	if err != nil {
		http.Error(w, "Error updating post: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Get the updated post
	err = db.QueryRow(
		"SELECT id, user_id, title, content, created_at, updated_at FROM posts WHERE id = $1",
		id,
	).Scan(&post.ID, &post.UserID, &post.Title, &post.Content, &post.CreatedAt, &post.UpdatedAt)
	if err != nil {
		http.Error(w, "Post not found after update", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(post)
}

// deletePost removes a blog post
func deletePost(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	// Delete the post
	result, err := db.Exec("DELETE FROM posts WHERE id = $1", id)
	if err != nil {
		http.Error(w, "Error deleting post: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Check if a row was actually deleted
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if rowsAffected == 0 {
		http.Error(w, "Post not found", http.StatusNotFound)
		return
	}

	// Return a success message
	w.WriteHeader(http.StatusNoContent)
}
