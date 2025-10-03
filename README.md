# pgxorb

[![Go Reference](https://pkg.go.dev/badge/github.com/moeryomenko/pgxorb.svg)](https://pkg.go.dev/github.com/moeryomenko/pgxorb)
[![Go Report Card](https://goreportcard.com/badge/github.com/moeryomenko/pgxorb)](https://goreportcard.com/report/github.com/moeryomenko/pgxorb)

**PostGIS geometry type support for pgx v5 through orb**

A Go library that seamlessly integrates [PostGIS](https://postgis.net/) geometry types with [`github.com/jackc/pgx/v5`](https://pkg.go.dev/github.com/jackc/pgx/v5) using the [`github.com/paulmach/orb`](https://pkg.go.dev/github.com/paulmach/orb) geometry library. This package provides a codec for encoding and decoding PostGIS geometries directly to and from orb types.

---

## 📑 Table of Contents

- [Features](#-features)
- [Installation](#-installation)
- [Usage](#-usage)
  - [Single Connection](#single-connection)
  - [Connection Pool](#connection-pool)
  - [Example Usage](#example-usage)
- [Technology Stack](#-technology-stack)
- [Project Structure](#-project-structure)
- [Development](#-development)
- [Testing](#-testing)
- [License](#-license)
- [Author](#-author)
- [Contributing](#-contributing)
- [FAQ](#-faq)

---

## ✨ Features

- **Full PostGIS geometry support** via orb types (Point, LineString, Polygon, MultiPoint, MultiLineString, MultiPolygon, GeometryCollection)
- **Binary and text format support** for optimal performance
- **EWKB encoding/decoding** with SRID support
- **Type-safe codec** implementation using pgx v5's type system
- **Zero-copy decoding** where possible
- **Easy registration** with both single connections and connection pools
- **Comprehensive test coverage** with testcontainers for integration testing

---

## 📦 Installation

```bash
go get github.com/moeryomenko/pgxorb
```

**Requirements:**
- Go 1.24.4 or higher
- PostgreSQL with PostGIS extension
- [github.com/jackc/pgx/v5](https://github.com/jackc/pgx) v5.7.5+
- [github.com/paulmach/orb](https://github.com/paulmach/orb) v0.11.1+

---

## 🚀 Usage

### Single Connection

```go
package main

import (
    "context"
    "log"
    "os"

    "github.com/jackc/pgx/v5"
    "github.com/moeryomenko/pgxorb"
    "github.com/paulmach/orb"
)

func main() {
    ctx := context.Background()

    connectionStr := os.Getenv("DATABASE_URL")
    conn, err := pgx.Connect(ctx, connectionStr)
    if err != nil {
        log.Fatal(err)
    }
    defer conn.Close(ctx)

    // Register pgxorb codec for PostGIS geometry types
    if err := pgxorb.Register(ctx, conn); err != nil {
        log.Fatal(err)
    }

    // Now you can use orb types directly
    point := orb.Point{1.5, 2.5}
    var result orb.Point

    err = conn.QueryRow(ctx, "SELECT $1::geometry", point).Scan(&result)
    if err != nil {
        log.Fatal(err)
    }

    log.Printf("Point: %v\n", result)
}
```

### Connection Pool

```go
package main

import (
    "context"
    "log"
    "os"

    "github.com/jackc/pgx/v5"
    "github.com/jackc/pgx/v5/pgxpool"
    "github.com/moeryomenko/pgxorb"
)

func main() {
    ctx := context.Background()

    connectionStr := os.Getenv("DATABASE_URL")
    config, err := pgxpool.ParseConfig(connectionStr)
    if err != nil {
        log.Fatal(err)
    }

    // Register codec on each new connection
    config.AfterConnect = func(ctx context.Context, conn *pgx.Conn) error {
        return pgxorb.Register(ctx, conn)
    }

    pool, err := pgxpool.NewWithConfig(ctx, config)
    if err != nil {
        log.Fatal(err)
    }
    defer pool.Close()

    // Use the pool normally
    // ...
}
```

### Example Usage

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/jackc/pgx/v5"
    "github.com/moeryomenko/pgxorb"
    "github.com/paulmach/orb"
)

type Location struct {
    ID       int
    Name     string
    Position orb.Point
}

func main() {
    ctx := context.Background()
    conn, err := pgx.Connect(ctx, "postgres://user:pass@localhost/dbname")
    if err != nil {
        log.Fatal(err)
    }
    defer conn.Close(ctx)

    if err := pgxorb.Register(ctx, conn); err != nil {
        log.Fatal(err)
    }

    // Create table with geometry column
    _, err = conn.Exec(ctx, `
        CREATE TABLE IF NOT EXISTS locations (
            id SERIAL PRIMARY KEY,
            name TEXT,
            position GEOMETRY(POINT, 4326)
        )
    `)
    if err != nil {
        log.Fatal(err)
    }

    // Insert location with orb.Point
    location := Location{
        Name:     "Tokyo Tower",
        Position: orb.Point{139.7454, 35.6586},
    }

    err = conn.QueryRow(ctx, `
        INSERT INTO locations (name, position)
        VALUES ($1, ST_SetSRID($2::geometry, 4326))
        RETURNING id
    `, location.Name, location.Position).Scan(&location.ID)
    if err != nil {
        log.Fatal(err)
    }

    // Query location
    var result Location
    err = conn.QueryRow(ctx, `
        SELECT id, name, position
        FROM locations
        WHERE id = $1
    `, location.ID).Scan(&result.ID, &result.Name, &result.Position)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Location: %s at %v\n", result.Name, result.Position)
}
```

---

## 🛠 Technology Stack

**Core Dependencies:**

- **Language:** Go 1.24.4
- **Database Driver:** [pgx v5](https://github.com/jackc/pgx) v5.7.5 - High-performance PostgreSQL driver
- **Geometry Library:** [orb](https://github.com/paulmach/orb) v0.11.1 - 2D geometry types and utilities
- **PostGIS:** PostGIS extension for spatial operations

**Development & Testing:**

- **Testing:** [testcontainers-go](https://github.com/testcontainers/testcontainers-go) v0.37.0 - Integration testing with Docker containers
- **Test Comparison:** [go-cmp](https://github.com/google/go-cmp) v0.7.0 - Deep equality testing
- **Linting:** golangci-lint (configured via `.golangci.yml`)

**Build Tools:**

- `Makefile` for common development tasks
- `go.work` for multi-module workspace support

---

## 📁 Project Structure

```
pgxorb/
├── geom.go              # Core geometry codec implementation (EWKB encoding/decoding)
├── geom_test.go         # Comprehensive integration tests with PostGIS
├── pgxorb.go            # Public API (Register function)
├── go.mod               # Go module dependencies
├── go.sum               # Dependency checksums
├── go.work              # Go workspace configuration
├── go.work.sum          # Workspace checksums
├── Makefile             # Development commands (test, lint, cover, mod)
├── .golangci.yml        # Linter configuration
├── .gitignore           # Git ignore rules
├── LICENSE-MIT          # MIT License
├── LICENSE-APACHE       # Apache 2.0 License
├── README.md            # This file
└── tools/               # Development tools directory
    ├── go.mod
    └── go.sum
```

**Key Files:**

- **`geom.go`** - Implements the pgx codec interface for PostGIS geometry types, handling both binary (EWKB) and text format encoding/decoding
- **`pgxorb.go`** - Main entry point providing the `Register()` function to register the codec with a pgx connection
- **`geom_test.go`** - Integration tests using testcontainers to spin up a PostGIS database for testing

---

## 🔧 Development

The project uses a Makefile for common development tasks:

```bash
# Run tests with coverage
make test

# Run tests with race detector
RACE_DETECTOR=1 make test

# View coverage in browser
make cover

# Run linter
make lint

# Tidy dependencies
make mod

# Show all available commands
make help
```

---

## 🧪 Testing

Tests use [testcontainers-go](https://github.com/testcontainers/testcontainers-go) to automatically spin up a PostGIS-enabled PostgreSQL container:

```bash
# Run all tests
go test ./...

# Run with race detector
go test -race ./...

# Run with coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

The test suite covers:
- Null value handling
- Binary and text format encoding/decoding
- All orb geometry types (Point, LineString, Polygon, etc.)
- SRID support via PostGIS functions

---

## 📄 License

This project is dual-licensed under:

- [MIT License](./LICENSE-MIT)
- [Apache License 2.0](./LICENSE-APACHE)

You may choose either license for your use of this software.

---

## 👤 Author

**Maxim Eryomenko**

- GitHub: [@moeryomenko](https://github.com/moeryomenko)
- Email: maxim_eryomenko@rambler.ru

---

## 🤝 Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

**Guidelines:**

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Run tests and linter (`make test lint`)
4. Commit your changes (`git commit -m 'Add some amazing feature'`)
5. Push to the branch (`git push origin feature/amazing-feature`)
6. Open a Pull Request

---

## ❓ FAQ

### Why use pgxorb instead of other PostGIS libraries?

pgxorb is specifically designed for pgx v5 (the recommended PostgreSQL driver for Go) and uses the popular orb geometry library, providing a modern, type-safe, and performant solution with zero external C dependencies.

### What geometry types are supported?

All PostGIS geometry types through orb: `Point`, `LineString`, `Polygon`, `MultiPoint`, `MultiLineString`, `MultiPolygon`, and `GeometryCollection`.

### Does it support SRID (Spatial Reference System Identifier)?

Yes, SRID is preserved during encoding/decoding through EWKB format. Use PostGIS functions like `ST_SetSRID()` to set SRID values.

### Is it production-ready?

Yes, the library includes comprehensive test coverage with integration tests against a real PostGIS database. It follows pgx v5's type system best practices.

### Can I use this with pgx v4 or earlier?

No, pgxorb is designed specifically for pgx v5's new type system. For pgx v4, consider other libraries like `github.com/cridenour/go-postgis`.

---

<div align="center">

**If you find Squad useful, please consider giving it a ⭐️!**

Made with ❤️ by [Maxim Eryomenko](https://github.com/moeryomenko)

</div>
