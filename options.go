package opentelemetry

import (
	"go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/trace"
)

type config struct {
	spanExporters   []trace.SpanExporter
	metricExporters []metric.Exporter
	logExporters    []log.Exporter
}

type Option func(*config)

func WithLogExporters(exporters ...log.Exporter) Option {
	return func(c *config) {
		c.logExporters = append(c.logExporters, exporters...)
	}
}

func WithMetricExporters(exporters ...metric.Exporter) Option {
	return func(c *config) {
		c.metricExporters = append(c.metricExporters, exporters...)
	}
}

func WithSpanExporters(exporters ...trace.SpanExporter) Option {
	return func(c *config) {
		c.spanExporters = append(c.spanExporters, exporters...)
	}
}
