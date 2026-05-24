.PHONY: build test test-race lint fmt vet bench cover tidy clean

# Build all packages
build:
	go build ./...

# Run all tests
test:
	go test ./...

# Run tests with race detection
test-race:
	go test -race -count=1 ./...

# Run linter (requires golangci-lint v2)
lint:
	golangci-lint run

# Format code
fmt:
	gofmt -w .

# Run go vet
vet:
	go vet ./...

# Run benchmarks
bench:
	go test -bench=. -benchmem ./...

# Generate coverage report
cover:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

# Tidy dependencies
tidy:
	go mod tidy

# Clean generated files
clean:
	rm -f coverage.out coverage.html

# Run all checks (CI equivalent)
ci: fmt vet lint test-race
