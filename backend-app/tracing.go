package main

import (
	"context"
	"log/slog"
	"os"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.37.0"
)

// initTracerProvider initializes the OpenTelemetry TracerProvider.
// This is boilerplate code that sets up the OTel SDK to export traces.
// It checks for the OTEL_EXPORTER_OTLP_ENDPOINT env var.
// If set (in SKE), it sends to that STACKIT endpoint.
// If NOT set (local), it defaults to the local Jaeger (http://localhost:4318).
func initTracerProvider(logger *slog.Logger) (*sdktrace.TracerProvider, error) {
	ctx := context.Background()

	// Check if custom endpoint is set
	stackitEndpoint := os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT")

	var exporter sdktrace.SpanExporter
	var err error

	if stackitEndpoint != "" {
		// For Observability Tempo
		logger.Info("OTel Exporter initializing for STACKIT Observability...", "endpoint", stackitEndpoint)
		exporter, err = otlptracehttp.New(ctx,
			otlptracehttp.WithEndpoint(stackitEndpoint),
			otlptracehttp.WithInsecure(),
		)
	} else {
		// For Local Jaeger
		logger.Info("OTel Local (Jaeger) Exporter initializing...", "endpoint", "http://localhost:4318")
		exporter, err = otlptracehttp.New(ctx,
			otlptracehttp.WithInsecure(),
			otlptracehttp.WithEndpoint("localhost:4318"),
		)
	}

	if err != nil {
		return nil, err
	}

	// Define the "resource" (what this service is)
	res, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName("backend-app"),
			semconv.ServiceVersion("v1.0.0"),
		),
	)
	if err != nil {
		return nil, err
	}

	// Create the TracerProvider with the exporter and resource
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
	)

	// Set the global TracerProvider
	otel.SetTracerProvider(tp)

	// Set the global Propagator (important for context propagation)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))

	logger.Info("OTel TracerProvider initialized.")
	return tp, nil
}
