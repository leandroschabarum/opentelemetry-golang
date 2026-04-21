package opentelemetry

import (
	"context"
	"log/slog"
)

type transports struct {
	stdout slog.Handler
	otel   slog.Handler
}

func createLogger(stdout, otel slog.Handler) *transports {
	return &transports{stdout: stdout, otel: otel}
}

func (t *transports) Enabled(ctx context.Context, level slog.Level) bool {
	return t.stdout.Enabled(ctx, level) || t.otel.Enabled(ctx, level)
}

func (t *transports) Handle(ctx context.Context, record slog.Record) error {
	if t.stdout.Enabled(ctx, record.Level) {
		if err := t.stdout.Handle(ctx, record.Clone()); err != nil {
			return err
		}
	}

	if t.otel.Enabled(ctx, record.Level) {
		if err := t.otel.Handle(ctx, record); err != nil {
			return err
		}
	}

	return nil
}

func (t *transports) WithAttrs(attrs []slog.Attr) slog.Handler {
	return createLogger(t.stdout.WithAttrs(attrs), t.otel.WithAttrs(attrs))
}

func (t *transports) WithGroup(name string) slog.Handler {
	return createLogger(t.stdout.WithGroup(name), t.otel.WithGroup(name))
}
