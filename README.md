# pgxorb [![Go Reference](https://pkg.go.dev/badge/github.com/moeryomenko/pgxorb.svg)](https://pkg.go.dev/github.com/moeryomenko/pgxorb)

Package pgxorb provides [PostGIS](https://postgis.net/) and [`github.com/jackc/pgx/v5`](https://pkg.go.dev/github.com/jackc/pgx/v5) via [`github.com/paulmach/orb.`](https://pkg.go.dev/github.com/paulmach/orb.).

## Usage

### Single connection

```go
import (
    // ...

	"github.com/jackc/pgx/v5"
	"github.com/moeryomenko/pgxorb"
	"github.com/paulmach/orb"
)

// ...

    connectionStr := os.Getenv("DATABASE_URL")
    conn, err := pgx.Connect(context.Background(), connectionStr)
    if err != nil {
        return err
    }
    if err := pgxorb.Register(ctx, conn); err != nil {
        return err
    }
```

### Connection pool

```go
import (
    // ...

    "github.com/jackc/pgx/v5/pgxpool"
)

// ...

    config, err := pgxpool.ParseConfig(connectionStr)
    if err != nil {
        return err
    }
    config.AfterConnect = func(ctx context.Context, conn *pgx.Conn) error {
        if err := pgxorb.Register(ctx, conn); err != nil {
            return err
        }
        return nil
    }

    pool, err := pgxpool.NewWithConfig(context.Background(), config)
    if err != nil {
        return err
    }
```

## License

pgxorb is primarily distributed under the terms of both the MIT license and the Apache License (Version 2.0).

See [LICENSE-APACHE](LICENSE-APACHE) and/or [LICENSE-MIT](LICENSE-MIT) for details.
