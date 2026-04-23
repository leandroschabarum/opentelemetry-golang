package opentelemetry

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"os"
	"strings"
	"time"

	"go.opentelemetry.io/contrib/bridges/otelslog"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
)

type OpenTelemetry struct {
	LoggerProvider *log.LoggerProvider
	MeterProvider  *metric.MeterProvider
	TracerProvider *trace.TracerProvider
}

func New(ctx context.Context, service string, opts ...Option) (*OpenTelemetry, error) {
	var cfg config

	for _, opt := range opts {
		opt(&cfg)
	}

	name, version, _ := strings.Cut(service, ":")

	resources, err := resource.Merge(
		resource.Default(),
		resource.NewSchemaless(
			attribute.String("service.name", name),
			attribute.String("service.version", version),
		),
	)

	if err != nil {
		return nil, err
	}

	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	otel.SetErrorHandler(otel.ErrorHandlerFunc(func(err error) {
		var netErr *net.OpError
		if errors.As(err, &netErr) {
			return
		}

		slog.Error(err.Error())
	}))

	// Tracer provider
	tpOpts := []trace.TracerProviderOption{trace.WithResource(resources)}

	for _, exp := range cfg.spanExporters {
		tpOpts = append(tpOpts, trace.WithBatcher(exp))
	}

	tp := trace.NewTracerProvider(tpOpts...)
	otel.SetTracerProvider(tp)

	// Meter provider
	mpOpts := []metric.Option{metric.WithResource(resources)}

	for _, exp := range cfg.metricExporters {
		mpOpts = append(mpOpts, metric.WithReader(metric.NewPeriodicReader(exp)))
	}

	mp := metric.NewMeterProvider(mpOpts...)
	otel.SetMeterProvider(mp)

	// Logger provider
	lpOpts := []log.LoggerProviderOption{log.WithResource(resources)}

	for _, exp := range cfg.logExporters {
		lpOpts = append(lpOpts, log.WithProcessor(log.NewBatchProcessor(exp)))
	}

	lp := log.NewLoggerProvider(lpOpts...)
	otelHandler := otelslog.NewHandler(name, otelslog.WithLoggerProvider(lp))
	stdoutHandler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug})
	slog.SetDefault(slog.New(createLogger(stdoutHandler, otelHandler)))

	return &OpenTelemetry{
		TracerProvider: tp,
		MeterProvider:  mp,
		LoggerProvider: lp,
	}, nil
}

func (o *OpenTelemetry) Shutdown(ctx context.Context) error {
	if ctx.Err() != nil {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
	}

	var errs []error

	if o.TracerProvider != nil {
		if err := o.TracerProvider.Shutdown(ctx); err != nil {
			errs = append(errs, err)
		}
	}

	if o.MeterProvider != nil {
		if err := o.MeterProvider.Shutdown(ctx); err != nil {
			errs = append(errs, err)
		}
	}

	if o.LoggerProvider != nil {
		if err := o.LoggerProvider.Shutdown(ctx); err != nil {
			errs = append(errs, err)
		}
	}

	return errors.Join(errs...)
}
