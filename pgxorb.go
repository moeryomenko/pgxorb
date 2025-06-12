package pgxorb

import (
	"context"

	"github.com/jackc/pgx/v5"
)

func Register(ctx context.Context, conn *pgx.Conn) error {
	return registerGeom(ctx, conn)
}
