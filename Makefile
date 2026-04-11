include .envrc

# ==================================================================================== #
# HELPERS 
# ==================================================================================== #

## help: print this help message
.PHONY: help
help:
	@echo 'Usage:'
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' | sed -e 's/^/ /'

.PHONY: confirm
confirm:
	@echo -n "Are you sure? [y/N] " && read ans && [$${ans:-N} = y]

# ==================================================================================== #
# DEVELOPMENT 
# ==================================================================================== #

## run/api: run the cmd/api application
.PHONY: run/api
run/api:
	@go run ./cmd/api -db-dsn ${PRISM_DB_DSN}

## db/psql: connect to the database using psql
.PHONY: db/psql
db/psql:
	psql ${PRISM_DB_DSN}

## db/migrations/new name=$1: create a new database migration
.PHONY: db/migrations/new
db/migrations/new:
	@echo 'Creating migration file for ${name}...'
	@migrate create -seq -ext=.sql -dir=./migrations/ ${name}

## db/migrations/up: apply all up database migrations
.PHONY: db/migrations/up
db/migrations/up:
	@echo 'Running up migrations...'
	@migrate -path ./migrations/ -database ${PRISM_DB_DSN} up

## db/migrations/goto version=$1: roll back to specific migration version
.PHONY: db/migrations/goto
db/migrations/goto:
	@echo 'Rolling back to migration version ${version}...'
	@migrate -path ./migrations/ -database ${PRISM_DB_DSN} goto ${version}

## db/migrations/version: showing current migration version
.PHONY: db/migrations/version
db/migrations/version:
	@migrate -path ./migrations/ -database ${PRISM_DB_DSN} version

## db/migrations/down: apply all down database migrations
.PHONY: db/migrations/down
db/migrations/down:
	@echo 'Running down migrations...'
	@migrate -path ./migrations/ -database ${PRISM_DB_DSN} down

# ==================================================================================== #
# TESTING 
# ==================================================================================== #

## test/api: test the application
.PHONY: test/api
test/api:
	@go test -v ./cmd/api

## test/cover: generate test coverage report (html) and serves it over http
.PHONY: test/cover
test/cover:
	@echo 'Creating Coverage File...'
	@go test -coverprofile=coverage.out ./...
	@mkdir -p coverage
	@echo 'Creating Coverage HTML File...'
	@go tool cover -html=coverage.out -o=coverage/cover.html
	@echo 'Serving Coverage HTML File...'
	@python3 -m http.server 8080 -d coverage


# ==================================================================================== #
# BUILD 
# ==================================================================================== #

## build/api: build the cmd/api application
.PHONY: build/api
build/api:
	@echo 'Building cmd/api...'
	@go build -ldflags='-s' -o=./bin/api ./cmd/api
