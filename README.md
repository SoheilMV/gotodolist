# Go Todo List

A complete RESTful Todo List API built with Go, Gin framework, and MongoDB. This project is a Go implementation of the [jstodolist](https://github.com/SoheilMV/jstodolist) project.

## 📋 Features

- **Authentication & Authorization**
  - JWT-based authentication with refresh tokens
  - User registration and login
  - Secure logout mechanism
  - Token refresh for long-term sessions
  - Protected routes

- **Task Management**
  - CRUD operations for tasks (Create, Read, Update, Delete)
  - Filtering tasks by status, priority
  - Sorting by various fields
  - Pagination support

- **Database**
  - MongoDB integration with official Go driver
  - Proper data validation
  - Effective error handling

- **API Documentation**
  - Swagger UI at `/api-docs`
  - Complete API specifications

- **Logging System**
  - Custom file and console logging
  - Request logging with method, status code, latency, IP
  - Different log levels (DEBUG, INFO, WARNING, ERROR, SUCCESS)
  - Environment-aware logging (debug/release mode)

- **Development**
  - Environment configuration (.env)
  - Development and Production modes
  - CORS support
  - Proper error handling

## 📦 Tech Stack

- [Go](https://golang.org/) - Programming language
- [Gin](https://github.com/gin-gonic/gin) - Web framework
- [MongoDB](https://www.mongodb.com/) - Database
- [JWT](https://github.com/golang-jwt/jwt) - Authentication
- [Swagger](https://swagger.io/) - API documentation

## 🚀 Installation

### Prerequisites

- Go 1.21 or later
- MongoDB (local or remote)

### Setup

1. **Clone the repository**
   ```bash
   git clone https://github.com/yourusername/gotodolist.git
   cd gotodolist
   ```

2. **Install dependencies**
   ```bash
   go mod tidy
   ```

3. **Configure environment**
   
   Create a `.env` file in the root directory with the following variables:
   ```
   MONGO_URI=mongodb://localhost:27017
   DB_NAME=todolist
   PORT=8080
   GIN_MODE=debug # or 'release' for production
   CORS_ORIGIN=*
   JWT_SECRET=your-secret-key
   JWT_EXPIRE=24h
   LOG_FILE=logs/app.log
   ```

## 🏃‍♂️ Running the Application

### Development Mode
```bash
go run main.go
```

### Production Mode
1. Set `GIN_MODE=release` in your `.env` file or environment
2. Build and run:
   ```bash
   go build
   ./gotodolist
   ```

The API will be available at `http://localhost:8080` (or the PORT you specified).

## 📝 API Documentation

API documentation is available via Swagger UI at `/api-docs` when the application is running.

## 📊 Logging System

The application includes a comprehensive logging system that works differently based on the current environment:

### Log Levels

- **DEBUG**: Detailed information, typically of interest only when diagnosing problems
- **INFO**: Confirmation that things are working as expected
- **WARNING**: Indication that something unexpected happened, but the application is still working
- **ERROR**: Due to a more serious problem, the application couldn't perform some function
- **SUCCESS**: Successful completion of an important operation

### Log Format

Logs are formatted as:
```
[TIMESTAMP] [LEVEL] [FILE:LINE] MESSAGE
```

Example:
```
[2025-01-15 14:30:45] [INFO] [main.go:42] Server running on port 8080
```

### Request Logging

All HTTP requests are automatically logged with:
- HTTP method
- Status code
- Response time (latency)
- Client IP
- Request path
- User agent

### Configuration

In your `.env` file, set the path for log files:
```
LOG_FILE=logs/app.log
```

### Development vs. Production

- In development mode (`GIN_MODE=debug`), logs are written to both console and file
- In production mode (`GIN_MODE=release`), logs are written only to file to optimize performance

## 📌 API Endpoints

### Authentication

| Method | Endpoint         | Description                            | Authentication |
|--------|------------------|----------------------------------------|---------------|
| POST   | /auth/register   | Register a new user                    | No            |
| POST   | /auth/login      | User login                             | No            |
| POST   | /auth/refresh-token | Refresh access token                | No            |
| POST   | /auth/logout     | Logout and invalidate refresh token    | Yes           |
| GET    | /auth/me         | Get user info                          | Yes           |

### Tasks

| Method | Endpoint    | Description                | Authentication |
|--------|-------------|----------------------------|---------------|
| GET    | /tasks      | Get all tasks with filters | Yes           |
| GET    | /tasks/:id  | Get a specific task        | Yes           |
| POST   | /tasks      | Create a new task          | Yes           |
| PUT    | /tasks/:id  | Update a task              | Yes           |
| DELETE | /tasks/:id  | Delete a task              | Yes           |

### System

| Method | Endpoint    | Description       | Authentication |
|--------|-------------|-------------------|---------------|
| GET    | /health     | API health check  | No            |
| GET    | /api-docs   | API documentation | No            |

## 📄 Task Model

```go
type Task struct {
    ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
    Title       string             `bson:"title" json:"title" binding:"required"`
    Description string             `bson:"description,omitempty" json:"description"`
    Completed   bool               `bson:"completed" json:"completed"`
    DueDate     *time.Time         `bson:"dueDate,omitempty" json:"dueDate"`
    Priority    string             `bson:"priority" json:"priority"`
    User        primitive.ObjectID `bson:"user" json:"user"`
    CreatedAt   time.Time          `bson:"createdAt" json:"createdAt"`
    UpdatedAt   time.Time          `bson:"updatedAt" json:"updatedAt"`
}
```

## 📊 Query Parameters (GET /tasks)

| Parameter | Type    | Description                             | Example                   |
|-----------|---------|-----------------------------------------|---------------------------|
| completed | boolean | Filter by completion status             | ?completed=true           |
| priority  | string  | Filter by priority (low/medium/high)    | ?priority=high            |
| sort      | string  | Field to sort by                        | ?sort=createdAt           |
| sortDir   | string  | Sort direction (asc/desc)               | ?sortDir=desc             |
| page      | integer | Page number for pagination              | ?page=2                   |
| limit     | integer | Number of items per page                | ?limit=20                 |

Combined example: `/tasks?completed=false&priority=high&sort=dueDate&sortDir=asc&page=1&limit=10`

## 🔐 Authentication

This API uses JWT (JSON Web Tokens) for authentication with a refresh token system for improved security.

### Access Tokens
- Short-lived tokens (24h by default, configurable via JWT_EXPIRE)
- Used for authenticating API requests
- Must be included in the Authorization header: `Authorization: Bearer <your_token>`

### Refresh Tokens
- Long-lived tokens (7 days)
- Used to obtain new access tokens when they expire
- Securely stored in the database (hashed, not in raw form)
- Can be invalidated by user logout

### Token Flow
1. **Login/Register**: User receives both access and refresh tokens
2. **API Requests**: Access token is used for authentication
3. **Token Expiry**: When access token expires, use refresh token to get a new pair
4. **Logout**: Invalidates the refresh token, requiring a new login

## ✅ Example Usage

### Register a User

```bash
curl -X POST http://localhost:8080/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "johndoe",
    "email": "john@example.com",
    "password": "password123"
  }'
```

### Login

```bash
curl -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "john@example.com",
    "password": "password123"
  }'
```

### Refresh Token

```bash
curl -X POST http://localhost:8080/auth/refresh-token \
  -H "Content-Type: application/json" \
  -d '{
    "refreshToken": "your-refresh-token-here"
  }'
```

### Logout

```bash
curl -X POST http://localhost:8080/auth/logout \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

### Create a Task

```bash
curl -X POST http://localhost:8080/tasks \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -d '{
    "title": "Learn Go",
    "description": "Study Go programming basics",
    "priority": "high",
    "dueDate": "2023-12-31T23:59:59Z"
  }'
```

### Get All Tasks

```bash
curl http://localhost:8080/tasks \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

### Update a Task

```bash
curl -X PUT http://localhost:8080/tasks/task_id_here \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -d '{
    "completed": true
  }'
```

### Delete a Task

```bash
curl -X DELETE http://localhost:8080/tasks/task_id_here \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

## 🌟 Project Structure

```
gotodolist/
├── main.go              # Application entry point
├── go.mod               # Go module definition
├── go.sum               # Go module checksums
├── .env                 # Environment variables
├── .env.example         # Example environment variables
├── swagger.yaml         # API documentation
├── controllers/         # Request handlers
│   ├── auth_controller.go
│   └── task_controller.go
├── models/              # Data models
│   ├── task.go
│   └── user.go
├── routes/              # API routes
│   ├── auth_routes.go
│   └── task_routes.go
├── middleware/          # Middleware components
│   ├── auth.go
│   ├── logger.go        # Logging middleware
│   └── swagger.go
├── configs/             # Configuration code
│   └── db.go
├── utils/               # Utility functions
│   ├── env.go
│   ├── http.go
│   ├── token.go         # Token management utilities
│   └── logger.go        # Logging utilities
└── logs/                # Log files directory
    └── app.log          # Application logs
```

## 📜 License

This project is licensed under the MIT License - see the LICENSE file for details. 