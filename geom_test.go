package pgxorb_test

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxtest"
	"github.com/moeryomenko/pgxorb"
	"github.com/paulmach/orb"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

var defaultConnTestRunner pgxtest.ConnTestRunner

func TestMain(m *testing.M) {
	ctx := context.Background()

	container, err := postgres.Run(ctx, "postgis/postgis:15-3.3",
		postgres.WithDatabase("test-db"),
		postgres.WithUsername("postgres"),
		postgres.WithPassword("postgres"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).WithStartupTimeout(5*time.Second)),
	)
	if err != nil {
		log.Fatalf("failed to terminate pgContainer: %s", err)
	}

	defer func() {
		if err := container.Terminate(ctx); err != nil {
			log.Printf("failed to terminate pgContainer: %v", err)
		}
	}()

	p, err := container.MappedPort(ctx, "5432")
	if err != nil {
		log.Printf("failed to get container external port: %v", err)
		return
	}

	defaultConnTestRunner = pgxtest.DefaultConnTestRunner()
	defaultConnTestRunner.CreateConfig = func(ctx context.Context, t testing.TB) *pgx.ConnConfig {
		config, err := pgx.ParseConfig(
			fmt.Sprintf("postgres://postgres:postgres@localhost:%s/test-db?sslmode=disable", p.Port()),
		)
		if err != nil {
			t.Fatalf("ParseConfig failed: %v", err)
		}

		return config
	}
	defaultConnTestRunner.AfterConnect = func(ctx context.Context, tb testing.TB, conn *pgx.Conn) {
		tb.Helper()
		_, err := conn.Exec(ctx, "create extension if not exists postgis")
		if err != nil {
			tb.Fatalf("got unexpected error while creating postgis extension: %v", err)
		}
		err = pgxorb.Register(ctx, conn)
		if err != nil {
			tb.Fatalf("got unexpected error while registering pgxorb: %v", err)
		}
	}

	m.Run()
}

func TestGeometryCodecNull(t *testing.T) {
	defaultConnTestRunner.RunTest(context.Background(), t, func(ctx context.Context, tb testing.TB, conn *pgx.Conn) {
		tb.Helper()

		for _, format := range []int16{
			pgx.BinaryFormatCode,
			pgx.TextFormatCode,
		} {
			tb.(*testing.T).Run(strconv.Itoa(int(format)), func(t *testing.T) {
				var actual orb.Geometry
				err := conn.QueryRow(ctx, "select NULL::geometry", pgx.QueryResultFormats{format}).Scan(&actual)
				if err != nil {
					t.Fatal("got unexpected error", err)
				}

				if actual != nil {
					t.Fatalf("got unexpected value %v", actual)
				}
			})
		}
	})
}

func TestGeometryCodecPointer(t *testing.T) {
	defaultConnTestRunner.RunTest(context.Background(), t, func(ctx context.Context, tb testing.TB, conn *pgx.Conn) {
		tb.Helper()
		for _, format := range []int16{
			pgx.BinaryFormatCode,
			pgx.TextFormatCode,
		} {
			tb.(*testing.T).Run(strconv.Itoa(int(format)), func(t *testing.T) {
				want := orb.Point{1, 2}

				var got orb.Point

				err := conn.QueryRow(ctx, "select $1::geometry", pgx.QueryResultFormats{format}, want).Scan(&got)
				if err != nil {
					t.Fatal("got unexpected error", err)
				}

				if diff := cmp.Diff(want, got); diff != "" {
					t.Errorf("(-want +got):\\n%s", diff)
				}
			})
		}
	})
}

func TestGeometryCodecEncode(t *testing.T) {
	defaultConnTestRunner.RunTest(context.Background(), t, func(ctx context.Context, tb testing.TB, conn *pgx.Conn) {
		tb.Helper()
		for _, format := range []int16{
			pgx.BinaryFormatCode,
			pgx.TextFormatCode,
		} {
			tb.(*testing.T).Run(strconv.Itoa(int(format)), func(t *testing.T) {
				want := orb.Point{1, 2}

				var got orb.Point
				err := conn.QueryRow(ctx, "select $1::geometry", pgx.QueryResultFormats{format}, want).Scan(&got)
				if err != nil {
					t.Fatal("got unexpected error", err)
				}

				if diff := cmp.Diff(want, got); diff != "" {
					t.Errorf("(-want +got):\\n%s", diff)
				}
			})
		}
	})
}

func TestGeometryCodecEncodeNull(t *testing.T) {
	defaultConnTestRunner.RunTest(context.Background(), t, func(ctx context.Context, tb testing.TB, conn *pgx.Conn) {
		tb.Helper()
		for _, format := range []int16{
			pgx.BinaryFormatCode,
			pgx.TextFormatCode,
		} {
			tb.(*testing.T).Run(strconv.Itoa(int(format)), func(t *testing.T) {
				var got, want orb.Geometry
				err := conn.QueryRow(ctx, "select $1::geometry", pgx.QueryResultFormats{format}, got).Scan(&want)
				if err != nil {
					t.Fatal("got unexpected error", err)
				}

				if diff := cmp.Diff(got, want); diff != "" {
					t.Errorf("(-want +got):\\n%s", diff)
				}
			})
		}
	})
}

func TestGeometryCodecScan(t *testing.T) {
	defaultConnTestRunner.RunTest(context.Background(), t, func(ctx context.Context, tb testing.TB, conn *pgx.Conn) {
		tb.Helper()
		for _, format := range []int16{
			pgx.BinaryFormatCode,
			pgx.TextFormatCode,
		} {
			tb.(*testing.T).Run(strconv.Itoa(int(format)), func(t *testing.T) {
				original := orb.Point{1, 2}
				rows, err := conn.Query(ctx, "select $1::geometry", pgx.QueryResultFormats{format}, original)
				if err != nil {
					t.Fatalf("got unexpected error: %v", err)
				}
				if !rows.Next() {
					t.Fatal("expected row")
				}

				values, err := rows.Values()
				if err != nil {
					t.Fatalf("got unexpected error for rows.Values(): %v", err)
				}

				if len(values) != 1 {
					t.Fatalf("expected 1 value, got %d", len(values))
				}

				if rows.Next() {
					t.Fatalf("unexpected extra row: %v", rows.RawValues())
				}

				rows.Close()
			})
		}
	})
}

func TestGeometryCodecValue(t *testing.T) {
	defaultConnTestRunner.RunTest(context.Background(), t, func(ctx context.Context, tb testing.TB, conn *pgx.Conn) {
		tb.Helper()
		for _, format := range []int16{
			pgx.BinaryFormatCode,
			pgx.TextFormatCode,
		} {
			tb.(*testing.T).Run(strconv.Itoa(int(format)), func(t *testing.T) {
				var got orb.Point
				err := conn.
					QueryRow(ctx, "select ST_SetSRID('POINT(3 4)'::geometry, 4326)", pgx.QueryResultFormats{format}).
					Scan(&got)
				if err != nil {
					t.Fatal("got unexpected error", err)
				}

				want := orb.Point{3, 4}
				if diff := cmp.Diff(want, got); diff != "" {
					t.Errorf("(-want +got):\\n%s", diff)
				}
			})
		}
	})
}
