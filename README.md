# System Design

A code-first system design repository. Each subdirectory is a standalone implementation of a classic system, built to be runnable, testable, and scalable.

## Purpose

- **Learn by building**: Implement real systems (URL shortener, chat, feed, etc.) with proper storage, caching, and APIs.
- **Reference implementations**: Use these as templates or references for interviews and production-style designs.
- **Consistent structure**: Each system includes architecture notes, API docs, deployment (e.g. Docker), and capacity/load metrics where applicable.

## Systems

| System | Description | Stack |
|--------|-------------|--------|
| [url-shortener](./url-shortener) | Shorten URLs, redirect by short code, cache for read-heavy traffic | Go, PostgreSQL, Redis, Docker |

*(More systems will be added over time.)*

## Repo structure

```
system_design/
├── README.md                 # This file
├── url-shortener/            # URL shortener service
│   ├── cmd/server/           # Application entrypoint
│   ├── internal/             # Handlers, services, repository, DB, migrations
│   ├── pkg/                  # Shared packages (e.g. base62 encoding)
│   ├── docker-compose.yml
│   └── README.md             # Service-specific docs and metrics
└── ...
```

## Running a system

Each system has its own README with:

- Prerequisites (Go, Docker, etc.)
- Environment variables
- Run instructions (`go run`, `docker-compose up`)
- API examples and capacity/metrics

Start with the system’s directory and follow its README.

## Contributing

Add new systems as top-level directories with a clear structure, Docker support when possible, and a README that covers design, API, and load/capacity notes.
