package shared

import (
	"context"

	"booker/pkg/logger"

	"github.com/jackc/pgx/v5"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

var pgxTracer = otel.Tracer("pgx")

// PgxTracer implements pgx.QueryTracer for OTel span creation + logging.
type PgxTracer struct {
	log logger.Logger
}

func NewPgxTracer(log logger.Logger) *PgxTracer {
	return &PgxTracer{log: log}
}

func (t *PgxTracer) TraceQueryStart(ctx context.Context, _ *pgx.Conn, data pgx.TraceQueryStartData) context.Context {
	var span trace.Span
	ctx, span = pgxTracer.Start(ctx, "pgx.Query",
		trace.WithAttributes(
			attribute.String("db.statement", data.SQL),
		),
	)
	if span == nil {
		t.log.With("error", "span creation returned nil").Warn("failed to start pgx trace span")
	}
	return ctx
}

func (t *PgxTracer) TraceQueryEnd(ctx context.Context, _ *pgx.Conn, data pgx.TraceQueryEndData) {
	span := trace.SpanFromContext(ctx)
	defer span.End()

	if data.Err != nil {
		span.SetStatus(codes.Error, data.Err.Error())
		span.RecordError(data.Err)
		t.log.With("error", data.Err.Error()).Warn("pgx query failed")
	} else {
		span.SetAttributes(attribute.Int64("db.rows_affected", data.CommandTag.RowsAffected()))
	}
}
