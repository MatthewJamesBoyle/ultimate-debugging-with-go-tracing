package main

import (
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.7.0"
	trace2 "go.opentelemetry.io/otel/trace"
	"log"
	"net/http"
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
			semconv.ServiceNameKey.String("exercise"),
		)),
	)

	otel.SetTracerProvider(tp)
	tracer = tp.Tracer("exercise")

	// Define HTTP server and routes
	http.Handle("/", otelhttp.NewHandler(http.HandlerFunc(CalcHandler), "calc"))
	log.Fatal(http.ListenAndServe(":8082", nil))
}

func CalcHandler(w http.ResponseWriter, r *http.Request) {
	_ = calcOne(100000)

	_ = calcTwo(100000)
}

func calcOne(n int) int {
	return n * (n + 1) / 2
}

func calcTwo(n int) int {
	sum := 0
	for i := 1; i <= n; i++ {
		for j := 1; j <= i; j++ {
			sum += 1
		}
	}
	return sum
}
