.PHONY: test test-race bench vet fmt lint build clean

test:
	go test ./...

test-race:
	go test -race ./...

bench:
	go test -bench=. -benchmem ./...

vet:
	go vet ./...

fmt:
	gofmt -l .

fmt-fix:
	gofmt -w .

lint: vet fmt

build:
	go build ./...

clean:
	go clean -testcache

check: build test-race vet fmt
