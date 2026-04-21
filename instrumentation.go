package opentelemetry

import (
	"context"
	"errors"
	"log/slog"
	"strings"

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
	slogHandler := otelslog.NewHandler(name, otelslog.WithLoggerProvider(lp))
	slog.SetDefault(slog.New(slogHandler))

	return &OpenTelemetry{
		TracerProvider: tp,
		MeterProvider:  mp,
		LoggerProvider: lp,
	}, nil
}

func (o *OpenTelemetry) Shutdown(ctx context.Context) error {
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
