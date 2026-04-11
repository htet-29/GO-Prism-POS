# Prism POS

Prism POS is a Backend API Service Provider built with GO

## Application Structure

### Main Application Logic

- main application -> cmd/api/main.go
- http server -> cmd/api/server.go
- routing logic -> cmd/api/routes.go
- middlewares -> cmd/api/middlewares.go
- error responses -> cmd/api/errors.go
- helper functions -> cmd/api/helpers.go

### Internal Packages

- test assertion -> internal/assert
- version tagging -> internal/vcs

## How to start a API Server

### Command Line Flags

| Flags   | Usage                                           | Default       |
| ------- | ----------------------------------------------- | ------------- |
| port    | `'Server Port for API to listen to '`           | 4000          |
| env     | `"Environment (development,staging,production"` | "development" |
| version | "Display version and exit" (binary only)        | false         |

```bash
go run ./cmd/api -port=8080 -env=staging
```

## Makefile Usage

- make help -> print usage of each makefile command
- make run/api -> Run API server in development mode

### DATABASE Related

- make db/psql -> Open DATABASE in psql
- make db/migrations/new name=${name} -> create migration files with provided name
- make db/migrations/up -> apply up migrations
- make db/migrations/goto version=${version} -> rolling back to migration version
- make db/migrations/version -> show current migration version
- make db/migrations/down -> apply all down migrations

### Test Related

- make test/api -> Test API server
- make test/cover -> Serve coverage report over http server
- make build/api -> build API application to /bin folder

## Creating Database

```bash
psql
```

### Creating DATABASE and EXTENSION

```sql
CREATE DATABASE ${DB_NAME};
CREATE EXTENSION IF NOT EXISTS citext;
CREATE EXTENSION IF NOT EXISTS pg_trgm;
```

### Creating ROLE and ALTER ROLE and DATABASE

```sql
CREATE ROLE ${DB_USER} WITH LOGIN PASSWORD '${DB_USER_PASSWORD}';
GRANT ALL ON SCHEMA public to ${DB_USER};
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO ${DB_URSER};
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO ${DB_URSER};
ALTER DATABASE ${DB_NAME} OWNER TO ${DB_USER};
```
