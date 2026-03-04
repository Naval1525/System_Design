# URL Shortener

A production-style URL shortener: shorten long URLs to short codes, redirect by code, with Redis caching for read-heavy traffic.

## Architecture

```
                    ┌─────────────┐
                    │   Client    │
                    └──────┬──────┘
                           │
              POST /shorten │  GET /{code}
                           ▼
                    ┌─────────────┐
                    │  Go Server  │  (Gorilla Mux, port 8080)
                    │  Handlers   │
                    └──────┬──────┘
                           │
              ┌────────────┼────────────┐
              ▼            ▼            ▼
       ┌──────────┐  ┌──────────┐  ┌──────────┐
       │  Redis   │  │ Postgres │  │  Base62  │
       │  Cache   │  │  (urls)  │  │ Encoder  │
       └──────────┘  └──────────┘  └──────────┘
```

- **Shorten**: `POST /shorten` → insert URL in Postgres (auto-increment `id`) → encode `id` to Base62 short code → update row → cache code → original in Redis (24h TTL).
- **Redirect**: `GET /{code}` → lookup Redis → on miss, query Postgres by `short_code` → cache result → HTTP 302 redirect to original URL.

## Tech stack

| Layer | Technology |
|-------|------------|
| Language | Go 1.24+ |
| HTTP | Gorilla Mux |
| Database | PostgreSQL 15 (urls table, index on `short_code`) |
| Cache | Redis 7 (short code → original URL, 24h TTL) |
| Migrations | golang-migrate (embedded SQL) |
| Deployment | Docker + docker-compose |

## Project structure

```
url-shortener/
├── cmd/server/main.go          # Entrypoint: DB, Redis, migrations, router
├── internal/
│   ├── handler/                # HTTP: ShortenURL, RedirectURL, router
│   ├── service/                # ShortenURL, GetOriginalURL (cache + DB)
│   ├── repository/             # CreateURL, UpdateShortCode, GetByShortCode
│   ├── model/                  # URL struct
│   └── database/               # Postgres connection, Redis client, migrations
├── pkg/hash/                   # Base62 encode (id → short code)
├── docker-compose.yml
├── Dockerfile
└── README.md
```

## API

### Shorten URL

**Request**

```http
POST /shorten
Content-Type: application/json

{"url": "https://example.com/very/long/url"}
```

**Response** (200)

```json
{"short_url": "http://localhost:8080/1"}
```

Short code is Base62-encoded from the database `id` (e.g. `1` → `1`, `62` → `10`).

### Redirect

**Request**

```http
GET /{code}
```

**Response**

- **302 Found**: `Location: <original_url>`
- **404 Not Found**: Short code does not exist

## Configuration

Environment variables (with defaults where applicable):

| Variable | Description | Default |
|----------|-------------|---------|
| `POSTGRES_HOST` | PostgreSQL host | (required) |
| `POSTGRES_PORT` | PostgreSQL port | (required) |
| `POSTGRES_USER` | PostgreSQL user | (required) |
| `POSTGRES_PASSWORD` | PostgreSQL password | (required) |
| `POSTGRES_DB` | Database name | (required) |
| `REDIS_HOST` | Redis host | `localhost` |
| `REDIS_PORT` | Redis port | `6379` |

Optional: create a `.env` file in the project root; the app uses `godotenv` to load it. If Redis is unavailable, the app still runs and skips caching (reads go to Postgres only).

## Running locally

### With Docker (recommended)

```bash
docker-compose up --build
```

- App: http://localhost:8080  
- Postgres: localhost:5432 (user `admin`, password `admin`, db `urlshortener`)  
- Redis: localhost:6379  

### Without Docker

1. Run Postgres and Redis locally (or point env vars to existing instances).
2. Create database `urlshortener` (or your `POSTGRES_DB`).
3. From repo root:

```bash
go run ./cmd/server
```

Migrations run on startup. Server listens on `:8080`.

### Example usage

```bash
# Shorten
curl -X POST http://localhost:8080/shorten -H "Content-Type: application/json" -d '{"url":"https://google.com"}'
# {"short_url":"http://localhost:8080/1"}

# Redirect (follow in browser or with -L)
curl -I http://localhost:8080/1
# HTTP/1.1 302 Found
# Location: https://google.com
```

---

## Capacity & metrics

Estimates below assume a single instance of the app, one Postgres and one Redis node, and typical cloud VM (e.g. 4 vCPU, 8 GB RAM). Actual numbers depend on hardware, network, and traffic mix.

### Throughput

| Operation | Estimated throughput | Notes |
|-----------|----------------------|--------|
| **Redirect (GET /{code})** | **~15,000–40,000 req/s** | With high Redis hit ratio (>90%), most requests are in-memory; Go + Redis can handle this range on a single node. |
| **Redirect (cache miss)** | **~2,000–5,000 req/s** | Limited by Postgres single-node read capacity and connection handling. |
| **Shorten (POST /shorten)** | **~1,000–3,000 req/s** | Two writes per request (INSERT + UPDATE) plus one Redis SET; DB is the bottleneck. |

### Concurrent users

- **Read-heavy (e.g. 100:1 redirect vs shorten)**  
  - If 90%+ redirects are cache hits: **~10,000–30,000 concurrent users** (assuming ~1–2 req/s per user).
  - If redirects often miss cache: **~2,000–5,000 concurrent users** before Postgres becomes the limit.

- **Write-heavy (e.g. 10:1 redirect vs shorten)**  
  - **~1,000–2,000 concurrent users** (shorten throughput dominates).

### Storage (order of magnitude)

| Component | Assumption | Rough capacity |
|-----------|------------|----------------|
| **Postgres** | ~200 bytes per row (id, short_code, original_url, created_at) | ~5M URLs per 1 GB (table + index). 10 GB → tens of millions of URLs. |
| **Redis** | ~100–500 bytes per key (code + URL + TTL overhead), 24h TTL | 1 GB → hundreds of thousands to low millions of cached entries; eviction by TTL. |

### Latency (p95, typical)

| Operation | Target | Notes |
|-----------|--------|--------|
| Redirect (cache hit) | &lt; 2 ms | In-process + Redis round-trip. |
| Redirect (cache miss) | &lt; 10–20 ms | + Postgres query by indexed `short_code`. |
| Shorten | &lt; 20–50 ms | Two DB writes + one Redis SET. |

### Scaling beyond one node

- **Horizontal scaling**: Run multiple app instances behind a load balancer; Redis and Postgres are shared. Connection pooling and Redis connection limits should be tuned per instance.
- **Redis**: Add replicas for read scaling; use Redis Cluster for very high redirect QPS and memory.
- **Postgres**: Add read replicas and route redirect lookups (by `short_code`) to replicas; keep shorten traffic on primary. For huge scale, consider sharding by `id` or short code range.
- **Short code space**: Base62 on a 64-bit `id` gives billions of unique codes; no practical limit for most use cases.

### Summary table

| Metric | Value (single node, read-heavy) |
|--------|----------------------------------|
| Redirect (cached) | ~15k–40k req/s |
| Redirect (DB fallback) | ~2k–5k req/s |
| Shorten | ~1k–3k req/s |
| Concurrent users (read-heavy) | ~10k–30k (with high cache hit) |
| Concurrent users (write-heavy) | ~1k–2k |
| Redirect p95 latency | &lt; 2 ms (cached), &lt; 20 ms (DB) |
| Shorten p95 latency | &lt; 50 ms |
| Storage (URLs) | Millions to tens of millions per 10 GB DB |

These are engineering estimates; run load tests (e.g. `wrk`, `k6`, or `ab`) against your deployment to validate for your environment and traffic pattern.
