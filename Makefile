# run/api: Run the prism pos api
.PHONY: run/api
run/api:
	@go run ./cmd/api

.PHONY: test/api
test/api:
	@go test -v ./cmd/api

.PHONY: test/cover
test/cover:
	@echo 'Creating Coverage File...'
	@go test -coverprofile=coverage.out ./...
	@mkdir -p coverage
	@echo 'Creating Coverage HTML File...'
	@go tool cover -html=coverage.out -o=coverage/cover.html
	@echo 'Serving Coverage HTML File...'
	@python3 -m http.server 8080 -d coverage

.PHONY: build/api
build/api:
	@go build -ldflags='-s' -o=./bin/api ./cmd/api
