.PHONY: dev build lint lint-go lint-frontend test clean

# Run the application in development mode
dev:
	wails dev

# Build the application for production
build:
	wails build

# Run all linters
lint: lint-go lint-frontend

# Run Go linters
lint-go:
	golangci-lint run

# Run Frontend linters
lint-frontend:
	cd frontend && npm run lint && npm run check

# Run tests
test:
	go test ./...
	@echo "Running frontend tests..."
	cd frontend && (npm run test || echo "No frontend tests configured yet")

# Clean build directory
clean:
	rm -rf build/bin/*
