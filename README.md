# semanticcache
A Go library for caching semantic data using the LRU cache eviction policy.

## Project Overview
The semanticcache package is designed to provide a caching mechanism for semantic data. It leverages the LRU (Least Recently Used) cache eviction policy to optimize performance and scalability. The package is written in Go and utilizes the `github.com/hashicorp/golang-lru/v2` library for cache implementation.

## Key Features
* **Semantic Caching**: The package provides a caching mechanism for semantic data, allowing for efficient storage and retrieval of data.
* **LRU Cache Eviction**: The package uses the LRU cache eviction policy to ensure that the most recently used data is retained in the cache.
* **Customizable Capacity**: The package allows users to specify the capacity of the cache, enabling them to balance memory usage and performance.
* **Comparator Interface**: The package provides a comparator interface that allows users to define custom comparison logic for cache entries.

## Getting Started
To get started with the semanticcache package, follow these steps:

### Prerequisites
* **Go**: The package is written in Go and requires a Go environment to build and run.
* **github.com/hashicorp/golang-lru/v2**: The package depends on the `github.com/hashicorp/golang-lru/v2` library for cache implementation.

### Installation
To install the semanticcache package, run the following command:
```go
go get github.com/botirk38/semanticcache
```

## Usage
The semanticcache package provides a simple API for caching semantic data. Here's an example of how to use the package:
```go
package main

import (
	"github.com/botirk38/semanticcache"
)

func main() {
	// Create a new semantic cache with a capacity of 100
	cache, err := semanticcache.New(100, nil)
	if err != nil {
		// Handle error
	}

	// Add an entry to the cache
	cache.Add("key", "value")

	// Retrieve an entry from the cache
	value, ok := cache.Get("key")
	if ok {
		// Handle retrieved value
	}
}
```

## Architecture Overview
The semanticcache package consists of the following components:

* **SemanticCache**: The main cache struct that provides methods for adding, retrieving, and removing cache entries.
* **Entry**: A struct that represents a cache entry, containing the key, value, and other metadata.
* **Comparator**: An interface that allows users to define custom comparison logic for cache entries.

## Configuration
The semanticcache package provides several configuration options, including:

* **Capacity**: The maximum number of entries that the cache can hold.
* **Comparator**: A custom comparison function that can be used to compare cache entries.

## Contributing Guidelines
To contribute to the semanticcache package, follow these steps:

1. Fork the repository.
2. Create a new branch for your changes.
3. Commit your changes with a descriptive commit message.
4. Open a pull request against the main branch.

## License
The semanticcache package is licensed under the MIT License. See the LICENSE file for more information.

## Dependencies
The semanticcache package depends on the following libraries:

* **github.com/hashicorp/golang-lru/v2**: The LRU cache implementation.

## Compatibility
The semanticcache package is compatible with Go version 1.17 and later. It is recommended to use the latest version of Go for optimal performance and security.

## Testing
The semanticcache package includes a test suite that covers the main functionality of the package. To run the tests, use the following command:
```bash
go test -v
```
Note: The test suite is still under development and may not cover all edge cases. Contributions to the test suite are welcome.
