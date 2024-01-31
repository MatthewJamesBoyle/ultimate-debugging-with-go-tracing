package main

import (
	"fmt"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.7.0"
	trace2 "go.opentelemetry.io/otel/trace"
	"log"
	"net/http"
	"time"
)

var tracer trace2.Tracer

func main() {
	// Initialize Jaeger Exporter
	exporter, err := jaeger.New(jaeger.WithCollectorEndpoint())
	if err != nil {
		log.Fatal(err)
	}

	// Create Trace Provider
	tp := trace.NewTracerProvider(
		trace.WithBatcher(exporter),
		trace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String("app-two"),
		)),
	)

	otel.SetTracerProvider(tp)
	tracer = tp.Tracer("app-two")

	// Define HTTP server and routes
	http.Handle("/", otelhttp.NewHandler(http.HandlerFunc(SimpleHandler), "Index"))
	log.Fatal(http.ListenAndServe(":8082", nil))
}

func SimpleHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	_, span := tracer.Start(ctx, "SimpleHandler")
	defer span.End()

	time.Sleep(time.Millisecond * 500)
	// Perform some operations here
	fmt.Fprintf(w, "Hello from app-two!")
}
