# Production-Ready Todo API 📝

A robust, RESTful task management API built in Go. This project serves as a blueprint for production-grade backend architecture, featuring a strict separation of concerns, comprehensive middleware, automated CI/CD, and containerized deployment.

## 🏗 Architecture & Features

This API was designed moving beyond basic CRUD operations to incorporate real-world enterprise patterns:

* **Standard Go Layout:** Strict separation between the application entrypoint (`/cmd`) and private business logic (`/internal`), ensuring clean module boundaries.
* **Robust Middleware Chain:** Custom interceptors for JWT Authentication, API Rate Limiting, Request Logging, and Panic Recovery to ensure high availability and secure access.
* **Automated Documentation:** Fully integrated Swagger (OpenAPI) specification generated directly from code annotations.
* **CI/CD Integration:** Automated testing and build pipelines configured via GitHub Actions.
* **Containerized Infrastructure:** Seamless local deployment using Docker and Docker Compose with automated database schema initialization.

## 🛠 Tech Stack

* **Language:** Go (Golang)
* **Datastore:** Relational Database (SQL)
* **Infrastructure:** Docker, Docker Compose
* **API Documentation:** Swagger / swaggo
* **CI/CD:** GitHub Actions

## 📂 Repository Structure

```text
todo-api/
├── .github/workflows/
│   └── ci.yaml               # Automated CI/CD pipeline
├── cmd/api/
│   └── main.go               # Application entrypoint
├── docs/                     # Swagger/OpenAPI generated files
│   ├── docs.go
│   ├── swagger.json
│   └── swagger.yaml
├── init-db/
│   └── schema.sql            # Automated DB initialization script
├── internal/                 # Private application logic
│   ├── database/
│   │   └── db.go             # Database connection pooling
│   ├── handlers/
│   │   ├── auth.go           # Authentication routing
│   │   ├── todo.go           # Core domain logic
│   │   └── *_test.go         # Unit test suites
│   ├── middleware/
│   │   ├── auth.go           # JWT verification
│   │   ├── logging.go        # Request tracing
│   │   ├── rate_limit.go     # DDoS protection
│   │   └── recovery.go       # Panic handling
│   └── models/
│       ├── todo.go           # Data structures
│       └── user.go
├── docker-compose.yaml       # Local infrastructure orchestration
├── Dockerfile                # Multi-stage Go build
├── go.mod
└── go.sum
```

## 🚀 Getting Started

### Prerequisites
* Docker and Docker Compose installed on your local machine.

### Quickstart

1. **Clone and Configure:**
   ```bash
   git clone <your-repository-url>
   cd todo-api
   cp .env.example .env
   ```
   *(Update the `.env` file with your specific database credentials and JWT secrets if necessary).*

2. **Boot the Infrastructure:**
   This command will spin up the database, execute the `schema.sql` initialization, and build the Go API container.
   ```bash
   docker compose up --build -d
   ```

3. **Verify Health:**
   Check the logs to ensure the server started successfully and connected to the database.
   ```bash
   docker logs <your-api-container-name>
   ```

## 📖 API Documentation (Swagger)

Once the server is running, you can interact with the API via the automated Swagger UI.
* **Endpoint:** `http://localhost:<YOUR_PORT>/swagger/index.html`

### Core Endpoints

| Method | Endpoint | Description | Auth Required |
| :--- | :--- | :--- | :--- |
| `POST` | `/signup` | Register a new user | No |
| `POST` | `/login` | Authenticate and receive JWT | No |
| `GET` | `/todos` | Retrieve all user tasks | Yes |
| `POST` | `/todos` | Create a new task | Yes |
| `GET` | `/todos/{id}` | Get a todo | Yes |
| `PUT` | `/todos/{id}` | Update an existing task | Yes |
| `DELETE`| `/todos/{id}` | Remove a task | Yes |

## 🧪 Testing

The internal business logic is covered by unit tests. To run the test suite locally:

```bash
go test ./internal/... -v
```

## 👤 Author

**Faisal Amir Dipto** Engineered to demonstrate best practices in Go backend development, system architecture, and API security.
