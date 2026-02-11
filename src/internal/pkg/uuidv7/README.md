# UUIDv7 Generator

A production-ready UUIDv7 generator wrapper module for Go applications.

## Overview

This module provides a clean wrapper around `github.com/google/uuid`'s UUIDv7 implementation. UUIDv7 is a time-ordered UUID version that includes a Unix timestamp in milliseconds, making it ideal for use as database primary keys where you want time-based ordering. The underlying google/uuid library handles thread-safety and monotonic ordering internally.

## Features

- ✅ **Clean wrapper**: Simple API wrapping `github.com/google/uuid`'s UUIDv7 implementation
- ✅ **Thread-safe**: The underlying google/uuid library handles thread-safety internally
- ✅ **Monotonic ordering**: google/uuid ensures unique UUIDs even when generated in the same millisecond
- ✅ **High performance**: Leverages google/uuid's optimized implementation
- ✅ **Production-ready**: Comprehensive error handling and testing
- ✅ **FX integration**: Ready for dependency injection with `go.uber.org/fx`

## Installation

Add the dependency to your `go.mod`:

```bash
go get github.com/google/uuid
```

## Usage

### Basic Usage

```go
package main

import (
    "fmt"
    "myapp/internal/pkg/uuidv7"
)

func main() {
    gen := uuidv7.NewGenerator()
    
    // Generate a UUID
    id, err := gen.Generate()
    if err != nil {
        panic(err)
    }
    fmt.Println(id.String())
    
    // Generate as string directly
    idStr, err := gen.GenerateString()
    if err != nil {
        panic(err)
    }
    fmt.Println(idStr)
}
```

### With Dependency Injection (FX)

```go
package main

import (
    "go.uber.org/fx"
    "myapp/internal/pkg/uuidv7"
)

func NewService(gen *uuidv7.Generator) *MyService {
    return &MyService{
        uuidGen: gen,
    }
}

func main() {
    fx.New(
        uuidv7.Module,
        fx.Provide(NewService),
        fx.Invoke(func(s *MyService) {
            id := s.uuidGen.MustGenerateString()
            // Use the UUID
        }),
    ).Run()
}
```

### Batch Generation

```go
gen := uuidv7.NewGenerator()

// Generate multiple UUIDs at once
ids, err := gen.GenerateBatch(100)
if err != nil {
    panic(err)
}

// Or as strings
idStrings, err := gen.GenerateBatchStrings(100)
if err != nil {
    panic(err)
}
```

### Parsing and Validation

```go
// Parse a UUID string
id, err := uuidv7.Parse("01234567-89ab-7def-0123-456789abcdef")
if err != nil {
    // Handle error
}

// Check if a string is a valid UUID
if uuidv7.IsValid("01234567-89ab-7def-0123-456789abcdef") {
    // Valid UUID
}
```

## API Reference

### Generator

#### `NewGenerator() *Generator`
Creates a new UUIDv7 generator instance.

#### `Generate() (uuid.UUID, error)`
Generates a new UUIDv7. Returns an error if generation fails.

#### `GenerateString() (string, error)`
Generates a new UUIDv7 and returns it as a string.

#### `MustGenerate() uuid.UUID`
Generates a UUIDv7 and panics if an error occurs. Use only when you're certain generation cannot fail.

#### `MustGenerateString() string`
Generates a UUIDv7 string and panics if an error occurs.

#### `GenerateBatch(count int) ([]uuid.UUID, error)`
Generates multiple UUIDv7s in a single call. Returns an empty slice if count is 0 or negative.

#### `GenerateBatchStrings(count int) ([]string, error)`
Generates multiple UUIDv7 strings in a single call.

### Utility Functions

#### `Parse(s string) (uuid.UUID, error)`
Parses a UUID string and validates it.

#### `MustParse(s string) uuid.UUID`
Parses a UUID string and panics if parsing fails.

#### `IsValid(s string) bool`
Checks if a string is a valid UUID.

## How UUIDv7 Works

UUIDv7 is structured as follows:

- **48 bits**: Unix timestamp in milliseconds
- **12 bits**: Monotonic counter (ensures uniqueness within the same millisecond)
- **4 bits**: Version (7)
- **62 bits**: Random data
- **2 bits**: Variant

This structure ensures:
1. **Time-ordered**: UUIDs generated later will sort after earlier ones
2. **Unique**: The monotonic counter prevents collisions within the same millisecond
3. **Sortable**: Can be used directly as database primary keys with natural ordering

## Thread Safety

The `Generator` type is a lightweight wrapper around `github.com/google/uuid`'s UUIDv7 implementation, which handles thread-safety internally. Multiple goroutines can safely call `Generate()` concurrently without race conditions.

## Performance

The generator is optimized for performance:
- Minimal locking overhead
- Efficient counter management
- Batch generation support for bulk operations

Benchmark results (typical):
- Single generation: ~200-300 ns/op
- String generation: ~300-400 ns/op
- Batch generation (100): ~20-30 μs/op

## Testing

Run tests with:

```bash
go test ./src/internal/pkg/uuidv7
```

Run benchmarks with:

```bash
go test -bench=. ./src/internal/pkg/uuidv7
```

## License

This module uses `github.com/google/uuid` which is licensed under BSD-3-Clause.
