# Blog Platform Microservices

A simple blog platform built with Go microservices and PostgreSQL.

## Architecture

This application consists of three microservices:

1. **User Service** - Handles user registration, authentication, and profile management
2. **Post Service** - Manages blog posts (create, read, update, delete)
3. **Comment Service** - Handles comments on blog posts

All three services share a single PostgreSQL database.

## Getting Started

### Prerequisites

- Docker and Docker Compose
- Go 1.20+ (for local development)
- PostgreSQL (for local development)

### Directory Structure

```
blog-platform/
├── docker-compose.yml
├── init.sql
├── user-service/
│   ├── Dockerfile
│   ├── go.mod
│   ├── go.sum
│   └── main.go
├── post-service/
│   ├── Dockerfile
│   ├── go.mod
│   ├── go.sum
│   └── main.go
├── comment-service/
│   ├── Dockerfile
│   ├── go.mod
│   ├── go.sum
│   └── main.go
└── README.md
```

### Running with Docker Compose

1. Clone the repository
2. Navigate to the project root directory
3. Run docker-compose:

```bash
docker-compose up -d
```

This will start all three services and the PostgreSQL database.

### Service Endpoints

#### User Service (Port 8081)
- `POST /users` - Create a new user
- `POST /users/login` - User login
- `GET /users/:id` - Get user profile
- `PUT /users/:id` - Update user profile

#### Post Service (Port 8082)
- `POST /posts` - Create a new post
- `GET /posts` - List all posts
- `GET /posts/:id` - Get a specific post
- `PUT /posts/:id` - Update a post
- `DELETE /posts/:id` - Delete a post

#### Comment Service (Port 8083)
- `POST /posts/:id/comments` - Add a comment to a post
- `GET /posts/:id/comments` - Get all comments for a post
- `PUT /comments/:id` - Update a comment
- `DELETE /comments/:id` - Delete a comment

## Development

### Local Setup

For each service:

1. Create the directory structure
2. Create the `go.mod` file
3. Create the `main.go` file
4. Install dependencies:

```bash
go mod download
```

### Building and Running Each Service

```bash
# In each service directory
go build -o main .
./main
```

### Environment Variables

Each service uses the following environment variables:

- `DB_HOST` - PostgreSQL host
- `DB_PORT` - PostgreSQL port
- `DB_USER` - PostgreSQL user
- `DB_PASSWORD` - PostgreSQL password
- `DB_NAME` - PostgreSQL database name
- `PORT` - Service port (default varies by service)

## Improvements for Production

This is a minimal implementation. For production, consider:

1. Adding authentication with JWT tokens
2. Adding input validation
3. Implementing proper error handling
4. Adding logging and monitoring
5. Adding unit and integration tests
6. Implementing API gateway
7. Adding rate limiting
8. Adding caching