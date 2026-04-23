# opentelemetry-golang
Boilerplate for OpenTelemetry instrumentation.
----

## Getting started:

The first step is to instrument your application code with [OpenTelemetry](https://opentelemetry.io) and send its metrics, traces, and logs to the OpenTelemetry [collector](https://opentelemetry.io/docs/collector).

For this example, we will use the OTLP HTTP exporters:

```bash
go get go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp
go get go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp
go get go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp
```

Then set up instrumentation in your application:

```go
package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"time"

	opentelemetry "github.com/leandroschabarum/opentelemetry-golang"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	traceExporter, _ := otlptracehttp.New(ctx,
		otlptracehttp.WithEndpointURL("http://localhost:4318/v1/traces"),
	)
	metricExporter, _ := otlpmetrichttp.New(ctx,
		otlpmetrichttp.WithEndpointURL("http://localhost:4318/v1/metrics"),
	)
	logExporter, _ := otlploghttp.New(ctx,
		otlploghttp.WithEndpointURL("http://localhost:4318/v1/logs"),
	)

	o, _ := opentelemetry.New(ctx, "my-app:1.0.0",
		opentelemetry.WithSpanExporters(traceExporter),
		opentelemetry.WithMetricExporters(metricExporter),
		opentelemetry.WithLogExporters(logExporter),
	)
	defer o.Shutdown(ctx)

	// Use the global tracer and slog as usual
	tracer := otel.Tracer("my-app")
	ctx, span := tracer.Start(ctx, "example-operation")
	defer span.End()

	slog.InfoContext(ctx, "application started")
}
```

You can use other available [exporters](https://opentelemetry.io/docs/languages/go/exporters), just remember to install them in your project first.

## Infrastructure setup:

After your code is instrumented, the next step is to spin up the OpenTelemetry [collector](https://opentelemetry.io/docs/collector) and the LGTM stack ([Loki](https://grafana.com/oss/loki), [Grafana](https://grafana.com/oss/grafana), [Tempo](https://grafana.com/oss/tempo) and [Mimir](https://grafana.com/oss/mimir)).

If you need instructions for configuring your deployment, refer to the examples in Grafana’s [intro-to-mltp](https://github.com/grafana/intro-to-mltp) repository or the official [Grafana Labs](https://grafana.com) documentation.

----

## Notice

This project, **opentelemetry-golang**, makes use of the [OpenTelemetry](https://opentelemetry.io) packages published under the [Apache License 2.0](https://www.apache.org/licenses/LICENSE-2.0).

This package itself is licensed under the [MIT license](./LICENSE).

**Disclaimer:** This project is provided “as is”, without warranty of any kind. The author assumes no responsibility for how this package is used.
