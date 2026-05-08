# Docker & Docker Compose Analysis

## Multi-Stage Build — Stage Breakdown

```text
Stage 1 — builder  (golang:1.26-alpine)
  ├── COPY go.mod + go.sum → RUN go mod download   ← cached dependency layer
  ├── COPY . .
  └── RUN CGO_ENABLED=0 GOOS=linux go build -o /api-server ./cmd/api

Stage 2 — runtime  (alpine:3.19)
  ├── RUN apk add ca-certificates
  ├── COPY --from=builder /api-server .             ← only the binary crosses the boundary
  └── ENTRYPOINT ["./api-server"]
```

**Stage 1 — Builder:** Compiles the application using the full Go toolchain. Dependencies are downloaded in a separate layer before the source copy, so source-only changes reuse the cached module layer.

**Stage 2 — Runtime:** Starts from a minimal Alpine base (~7 MB). Only the compiled binary is copied over — the Go toolchain, module cache, and source tree are discarded entirely. `ca-certificates` is added for TLS support.

---

## `CGO_ENABLED=0`

Disables Go's C interoperability layer, producing a **statically linked binary** with zero runtime C library dependencies. This is required so the binary runs inside the minimal Alpine runtime image without needing `glibc` or `musl` symbols from the OS.

---

## Image Size Comparison

| Approach | Size |
| :--- | :--- |
| Single-stage (`golang:1.26-alpine`) | ~260 MB |
| Multi-stage final image | **28.4 MB** |

The multi-stage build achieves a **~90 % size reduction** by shipping only the compiled binary and CA certificates — no compiler, no source code, no module cache.

---

## Docker Compose — Service Architecture

Two services are defined:

* **`db` (`postgres:16-alpine`):** Runs with a `pg_isready` healthcheck. Data is persisted on the named volume `pgdata` and survives `docker compose down`.
* **`api` (built from `Dockerfile`):** Waits for `db` to be healthy before starting (`depends_on: condition: service_healthy`). Reaches the database via Docker's internal DNS hostname `db`.

```text
$ docker compose up --build

 Container ci-cd-mcm-golic-db-1   Started → Healthy
 Container ci-cd-mcm-golic-api-1  Started
 Final image: ci-cd-mcm-golic-api:latest  (28.4 MB)
```

---

## CRUD Test — Commands & Results

> **Note on port 8081:** Port 8080 was already occupied by another local service during testing. To avoid modifying `docker-compose.yml`, the `api` container was started manually with `-p 8081:8080` while the `db` service was brought up via Compose as usual. All commands below use port `8081` accordingly.

**Setup (run once before the CRUD commands):**

```bash
docker compose up -d db

docker run -d --name api-test \
  --network ci-cd-mcm-golic_default \
  -p 8081:8080 \
  -e PORT=8080 \
  -e DB_HOST=db \
  -e DB_PORT=5432 \
  -e DB_USER=catalog \
  -e DB_PASSWORD=catalog123 \
  -e DB_NAME=productcatalog \
  ci-cd-mcm-golic-api
```

### Create (POST)

```bash
curl -s -X POST http://localhost:8081/products \
  -H "Content-Type: application/json" \
  -d '{"name":"Nocco Ramonade","price":1.79}'
# {"id":1,"name":"Nocco Ramonade","price":1.79}

curl -s -X POST http://localhost:8081/products \
  -H "Content-Type: application/json" \
  -d '{"name":"Nocco Grand Sour","price":1.89}'
# {"id":2,"name":"Nocco Grand Sour","price":1.89}

curl -s -X POST http://localhost:8081/products \
  -H "Content-Type: application/json" \
  -d '{"name":"Nocco Limon del Sol","price":1.99}'
# {"id":3,"name":"Nocco Limon del Sol","price":1.99}
```

### Read (GET)

```bash
curl -s http://localhost:8081/products
# [{"id":1,"name":"Nocco Ramonade","price":1.79},{"id":2,"name":"Nocco Grand Sour","price":1.89},{"id":3,"name":"Nocco Limon del Sol","price":1.99}]

curl -s http://localhost:8081/products/2
# {"id":2,"name":"Nocco Grand Sour","price":1.89}
```

### Update (PUT)

```bash
curl -s -X PUT http://localhost:8081/products/1 \
  -H "Content-Type: application/json" \
  -d '{"name":"Nocco Ramonade","price":1.89}'
# {"id":1,"name":"Nocco Ramonade","price":1.89}
```

### Delete (DELETE)

```bash
curl -s -X DELETE http://localhost:8081/products/3
# {"result":"success"}

curl -s http://localhost:8081/products
# [{"id":1,"name":"Nocco Ramonade","price":1.89},{"id":2,"name":"Nocco Grand Sour","price":1.89}]
```

---

## Persistence Test

```bash
docker stop api-test && docker rm api-test
docker compose down

docker compose up -d db
docker run -d --name api-test \
  --network ci-cd-mcm-golic_default \
  -p 8081:8080 \
  -e PORT=8080 -e DB_HOST=db -e DB_PORT=5432 \
  -e DB_USER=catalog -e DB_PASSWORD=catalog123 -e DB_NAME=productcatalog \
  ci-cd-mcm-golic-api

curl -s http://localhost:8081/products
# [{"id":1,"name":"Nocco Ramonade","price":1.89},{"id":2,"name":"Nocco Grand Sour","price":1.89}]
```

Both products survived the full restart. The named volume `pgdata` retains the Postgres data directory across `docker compose down`, so no rows are lost.
