package main

import (
	"context"
	"log/slog"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.37.0"
)

// initTracerProvider initializes the OpenTelemetry TracerProvider.
// This is boilerplate code that sets up the OTel SDK to export traces.
// In a real STACKIT integration, the OTLP_ENDPOINT would be set to the
// STACKIT Observability OTLP ingest-URL.
func initTracerProvider(logger *slog.Logger) (*sdktrace.TracerProvider, error) {
	ctx := context.Background()

	// Configure the OTLP/HTTP exporter
	// We use insecure here for local dev; in prod, this would use TLS.
	// The endpoint is usually set via ENV var OTEL_EXPORTER_OTLP_ENDPOINT
	exporter, err := otlptracehttp.New(ctx, otlptracehttp.WithInsecure())
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
