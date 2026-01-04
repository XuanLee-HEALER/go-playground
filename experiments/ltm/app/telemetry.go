package main

import (
	"context"
	"errors"
	"os"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/log"
	"go.opentelemetry.io/otel/log/global"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/propagation"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	sdkresource "go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
	"go.opentelemetry.io/otel/trace"
)

type Telemetry struct {
	Tracer  trace.Tracer
	Meter   metric.Meter
	Logger  log.Logger
	Metrics Metrics
}

type Metrics struct {
	Requests  metric.Int64Counter
	LatencyMs metric.Float64Histogram
	Inflight  metric.Int64UpDownCounter
}

func setupTelemetry(ctx context.Context) (*Telemetry, func(context.Context) error, error) {
	endpoint := getenv("OTEL_EXPORTER_OTLP_ENDPOINT", "localhost:4317")

	resource, err := sdkresource.New(
		ctx,
		sdkresource.WithAttributes(
			semconv.ServiceNameKey.String("ltm-lab"),
			semconv.ServiceVersionKey.String("0.1.0"),
		),
	)
	if err != nil {
		return nil, nil, err
	}

	traceExporter, err := otlptracegrpc.New(ctx,
		otlptracegrpc.WithEndpoint(endpoint),
		otlptracegrpc.WithInsecure(),
	)
	if err != nil {
		return nil, nil, err
	}

	traceProvider := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(traceExporter),
		sdktrace.WithResource(resource),
	)
	otel.SetTracerProvider(traceProvider)
	otel.SetTextMapPropagator(propagation.TraceContext{})

	metricExporter, err := otlpmetricgrpc.New(ctx,
		otlpmetricgrpc.WithEndpoint(endpoint),
		otlpmetricgrpc.WithInsecure(),
	)
	if err != nil {
		return nil, nil, err
	}

	metricReader := sdkmetric.NewPeriodicReader(metricExporter, sdkmetric.WithInterval(5*time.Second))
	metricProvider := sdkmetric.NewMeterProvider(
		sdkmetric.WithReader(metricReader),
		sdkmetric.WithResource(resource),
	)
	otel.SetMeterProvider(metricProvider)

	logExporter, err := otlploggrpc.New(ctx,
		otlploggrpc.WithEndpoint(endpoint),
		otlploggrpc.WithInsecure(),
	)
	if err != nil {
		return nil, nil, err
	}

	logProcessor := sdklog.NewBatchProcessor(logExporter)
	logProvider := sdklog.NewLoggerProvider(
		sdklog.WithProcessor(logProcessor),
		sdklog.WithResource(resource),
	)
	global.SetLoggerProvider(logProvider)

	meter := metricProvider.Meter("ltm-lab")
	requests, err := meter.Int64Counter(
		"requests_total",
		metric.WithDescription("Total number of HTTP requests."),
	)
	if err != nil {
		return nil, nil, err
	}

	latency, err := meter.Float64Histogram(
		"request_latency_ms",
		metric.WithUnit("ms"),
		metric.WithDescription("Request latency in milliseconds."),
	)
	if err != nil {
		return nil, nil, err
	}

	inflight, err := meter.Int64UpDownCounter(
		"inflight_requests",
		metric.WithDescription("Number of in-flight HTTP requests."),
	)
	if err != nil {
		return nil, nil, err
	}

	telemetry := &Telemetry{
		Tracer: otel.Tracer("ltm-lab"),
		Meter:  meter,
		Logger: global.Logger("ltm-lab"),
		Metrics: Metrics{
			Requests:  requests,
			LatencyMs: latency,
			Inflight:  inflight,
		},
	}

	shutdown := func(ctx context.Context) error {
		var shutdownErr error
		if err := traceProvider.Shutdown(ctx); err != nil {
			shutdownErr = errors.Join(shutdownErr, err)
		}
		if err := metricProvider.Shutdown(ctx); err != nil {
			shutdownErr = errors.Join(shutdownErr, err)
		}
		if err := logProvider.Shutdown(ctx); err != nil {
			shutdownErr = errors.Join(shutdownErr, err)
		}
		return shutdownErr
	}

	return telemetry, shutdown, nil
}

func emitLog(ctx context.Context, logger log.Logger, message string, attrs ...log.KeyValue) {
	var record log.Record
	record.SetTimestamp(time.Now())
	record.SetSeverity(log.SeverityInfo)
	record.SetBody(log.StringValue(message))
	record.AddAttributes(attrs...)

	logger.Emit(ctx, record)
}

func getenv(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}
