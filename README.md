![alt tag](https://upload.wikimedia.org/wikipedia/commons/2/23/Golang.png)

[![License](https://img.shields.io/github/license/Massad/gin-boilerplate)](https://github.com/Massad/gin-boilerplate/blob/master/LICENSE) [![GitHub release (latest by date)](https://img.shields.io/github/v/release/Massad/gin-boilerplate)](https://github.com/Massad/gin-boilerplate/releases) [![Go Version](https://img.shields.io/github/go-mod/go-version/Massad/gin-boilerplate)](https://github.com/Massad/gin-boilerplate/blob/master/go.mod) [![DB Version](https://img.shields.io/badge/DB-PostgreSQL--latest-blue)](https://github.com/Massad/gin-boilerplate/blob/master/go.mod) [![DB Version](https://img.shields.io/badge/DB-Redis--latest-blue)](https://github.com/Massad/gin-boilerplate/blob/master/go.mod)

# Golang Gin Boilerplate v3

The fastest way to deploy a RESTful API with [Gin Framework](https://github.com/gin-gonic/gin/) — structured with **PostgreSQL**, **JWT** authentication stored in **Redis**, and ready to build on.

## What's Included

- [sqlx](https://github.com/jmoiron/sqlx): Lightweight SQL extensions for Go
- [jwt-go](https://github.com/golang-jwt/jwt): JSON Web Tokens (JWT) middleware
- [go-redis](https://github.com/go-redis/redis): Redis client
- [maroto](https://github.com/johnfercher/maroto): Pure Go PDF generation
- Built-in **CORS**, **RequestID**, and **Auth** middleware
- Built-in **custom form validators** with reusable error translation
- **Invoice example** — HTML preview and PDF download
- **Swagger API documentation**
- PostgreSQL with JSON/JSONB queries and trigger functions
- SSL support
- Go Modules

## Getting Started

### Prerequisites

- Go 1.24+
- PostgreSQL
- Redis

### Setup

Clone the repository:

```bash
git clone https://github.com/Massad/gin-boilerplate.git
cd gin-boilerplate
```

Install dependencies:

```bash
go mod download
```

Set up your environment:

```bash
cp .env_rename_me .env
# Edit .env with your database credentials
```

Import the database schema:

```bash
psql -U postgres -h localhost < ./db/database.sql
```

The database includes trigger functions (`created_at_column()` and `update_at_column()`) that automatically manage `created_at` and `updated_at` timestamps on the **user** and **article** tables.

### Running

```bash
make run
```

Or directly:

```bash
go run *.go
```

### Building

```bash
go build -v
./gin-boilerplate
```

### Testing

Tests are integration tests that require running PostgreSQL and Redis:

```bash
go test -v -tags=all ./tests/*
```

### SSL (Optional)

To enable SSL, set `SSL=TRUE` in `.env` and generate certificates:

```bash
mkdir cert/
sh generate-certificate.sh
```

To disable SSL, set `SSL=FALSE` in `.env`.

## API Endpoints

### User
| Method | Path | Auth | Description |
|--------|------|------|-------------|
| POST | `/v1/user/register` | No | Register a new user |
| POST | `/v1/user/login` | No | Login and receive JWT tokens |
| GET | `/v1/user/logout` | Bearer | Logout (invalidates token) |

### Article
| Method | Path | Auth | Description |
|--------|------|------|-------------|
| POST | `/v1/article` | Bearer | Create an article |
| GET | `/v1/articles` | Bearer | Get all user articles |
| GET | `/v1/article/:id` | Bearer | Get one article |
| PUT | `/v1/article/:id` | Bearer | Update an article |
| DELETE | `/v1/article/:id` | Bearer | Delete an article |

### Auth
| Method | Path | Auth | Description |
|--------|------|------|-------------|
| POST | `/v1/token/refresh` | No | Refresh access and refresh tokens |

### Invoice (Example)
| Method | Path | Auth | Description |
|--------|------|------|-------------|
| GET | `/v1/invoice` | No | HTML preview of a sample invoice with download button |
| GET | `/v1/invoice/download` | No | Download sample invoice as PDF |

## Swagger Docs

Generate and view API documentation:

```bash
make generate_docs
make run
```

Then open: `https://localhost:9000/swagger/index.html`

## Invoice Demo

Once the server is running, open the invoice preview in your browser:

```
http://localhost:9000/v1/invoice
```

The page shows an HTML invoice with a **Download PDF** button that generates a PDF using [maroto](https://github.com/johnfercher/maroto) (pure Go, no external dependencies).

## Trusted Proxies & CORS

By default, `SetTrustedProxies(nil)` is configured — Gin will not trust any proxy headers. If you deploy behind a reverse proxy (nginx, CloudFlare, etc.), set your trusted proxy IPs:

```go
r.SetTrustedProxies([]string{"192.168.1.0/24", "10.0.0.0/8"})
```

CORS is configured in `middleware/cors.go`. Update the `Access-Control-Allow-Origin` header for your domain in production.

## Authentication

This boilerplate uses **Bearer Token** authentication:

1. **Login** returns an `access_token` (15 min) and `refresh_token` (7 days)
2. Include the access token in requests: `Authorization: Bearer <access_token>`
3. When the access token expires, use `/v1/token/refresh` with the refresh token to get new tokens
4. Both tokens are stored in Redis and invalidated on logout

## Project Structure

```
controllers/    HTTP handlers
models/         Database structs and queries (sqlx + raw SQL)
forms/          Request validation structs and reusable error translation
middleware/     CORS, RequestID, and JWT auth middleware
invoice/        Invoice HTML/PDF generation example
db/             PostgreSQL and Redis initialization
docs/           Swagger documentation (auto-generated)
public/         Static files and HTML templates
```

## Adding a New Resource

1. Create a model in `models/` with your SQL queries
2. Create form structs in `forms/` with binding tags and a `ValidationMessages` map
3. Create a controller in `controllers/` — use `forms.Translate(err, messages)` for validation errors
4. Register routes in `main.go`
5. Run `make generate_docs` to update Swagger

## Previous Versions

- **v2.0** — JWT authentication with Redis (current architecture)
- **v1.x** — Session/cookie authentication ([v1 branch](https://github.com/Massad/gin-boilerplate/tree/v1-session-cookies-auth))

## License

MIT
