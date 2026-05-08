# Architecture Overview

## Project Structure

We follow a clean, idiomatic Go project layout:

```text
CI-CD-MCM-Golic/
├── cmd/api/
│   └── main.go                  # Application entry point (wires up store, handlers, and routing)
├── internal/
│   ├── handler/
│   │   ├── handler.go           # HTTP handlers dedicated to the MemoryStore
│   │   ├── postgres_handler.go  # HTTP handlers dedicated to the PostgresStore
│   │   └── handler_test.go
│   ├── model/
│   │   ├── product.go           # Product domain model & Validate() logic
│   │   └── product_test.go
│   └── store/
│       ├── memory.go            # In-memory storage implementation
│       ├── memory_test.go
│       └── postgres.go          # PostgreSQL storage implementation
├── Dockerfile
├── docker-compose.yml
└── .github/workflows/ci.yml
```

---

## Request Lifecycle

1. **Routing:** Incoming HTTP requests are intercepted by the **gorilla/mux router**, which matches the method and path to the appropriate handler. Unmatched routes automatically fall back to `404 Not Found` or `405 Method Not Allowed`.
2. **Handling & Validation:** The handler decodes the incoming JSON payload (for `POST`/`PUT` requests) and triggers `model.Validate()` to ensure data integrity. 
3. **Data Operation:** Once validated, the request is passed down to the store. 
4. **Response Formatting:** * On success, `respondJSON` formats the payload and writes the appropriate HTTP status code.
   * On failure, `respondError` catches the issue and wraps the error message in a standardized JSON format (`{"error":"..."}`). If the store returns a `store.ErrNotFound`, it maps to a `404`. Any unhandled store errors default to a `500 Internal Server Error`.

**High-Level Flow:**
`HTTP Client → Router → Handler → Store → Database (if using Postgres)`

---

## Store Implementations

### MemoryStore
Designed for local development and fast testing. It stores products in a `map[int]Product` and ensures thread safety using a `sync.RWMutex`. 
* **Pros:** Blazing fast, zero external dependencies.
* **Cons:** Highly volatile (data is wiped upon application restart) and `GetAll` results are returned in a non-deterministic order due to Go's randomized map iteration.

### PostgresStore
Built for production. It wraps a standard `*sql.DB` connection pool. 
* **Initialization:** On boot, it calls `EnsureTable` to automatically run migrations and create the `products` table if it's missing. 
* **Operations:** `GetAll` enforces a deterministic order (`ORDER BY id`). `Create` leverages `INSERT … RETURNING id` to safely capture the primary key assigned by the database. 

---

## Storage Backend Comparison

| Dimension | MemoryStore | PostgresStore |
| :--- | :--- | :--- |
| **Persistence** | Volatile (lost on restart) | Persistent (survives restarts) |
| **Scalability** | Single-node only | Horizontally scalable across replicas |
| **Setup Overhead** | None | Requires an active PostgreSQL instance |
| **Performance** | Sub-microsecond (in-process) | ~1–10 ms (network + disk I/O) |
| **Result Ordering** | Non-deterministic | Deterministic (`ORDER BY id`) |
| **ACID Compliance** | None | Full PostgreSQL transactional guarantees |
| **Primary Use Case**| Unit testing, CI/CD, local dev | Production, multi-instance environments |
