// go.mod (use the same for all three services)
module blog-service

go 1.24

require (
	github.com/gorilla/mux v1.8.0
	github.com/lib/pq v1.10.9
	golang.org/x/crypto v0.12.0
)

// go.sum will be generated when you run go mod download
