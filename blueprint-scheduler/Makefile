# Blueprint Scheduler Service Makefile

.PHONY: run build clean deps

# Run the scheduler service
run:
	go run cmd/scheduler/main.go

# Build the scheduler service
build:
	go build -o bin/scheduler cmd/scheduler/main.go

# Install dependencies
deps:
	go mod download
	go mod tidy

# Clean build artifacts
clean:
	rm -rf bin/

# Run with development environment
run-dev: deps
	go run cmd/scheduler/main.go

# Docker build
docker-build:
	docker build -t blueprint-scheduler .

# Docker run
docker-run:
	docker run --env-file .env blueprint-scheduler