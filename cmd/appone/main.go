package main

import (
	"context"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.7.0"
	trace2 "go.opentelemetry.io/otel/trace"
	"log"
	"math/rand"
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
			semconv.ServiceNameKey.String("app-one"),
		)),
	)

	otel.SetTracerProvider(tp)
	tracer = tp.Tracer("app-one")

	http.Handle("/", otelhttp.NewHandler(http.HandlerFunc(SimpleHandler), "Hello"))
	http.Handle("/complex", otelhttp.NewHandler(http.HandlerFunc(ComplexHandler), "Complex"))
	log.Fatal(http.ListenAndServe(":8081", nil))
}

func SimpleHandler(w http.ResponseWriter, r *http.Request) {
	_, _ = w.Write([]byte("Hello, World!"))
}

func ComplexHandler(w http.ResponseWriter, r *http.Request) {
	time.Sleep(time.Second * 3)
	// Start a new span
	ctx, span := tracer.Start(r.Context(), "complexHandler")
	defer span.End()

	// Call app-two
	if err := callAppTwo(ctx); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	time.Sleep(time.Second * 1)

	_, _ = w.Write([]byte("Request processed successfully"))
}

func callAppTwo(ctx context.Context) error {
	// Create a span for this function
	ctx, span := tracer.Start(ctx, "callAppTwo")
	defer span.End()

	// Preparing the request to app-two
	req, err := http.NewRequestWithContext(ctx, "GET", "http://localhost:8082/", nil)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return err
	}

	// Adding event to span to indicate call to app-two
	span.AddEvent("Calling app-two")

	// Make the call to app-two
	client := http.Client{Transport: otelhttp.NewTransport(http.DefaultTransport)}
	resp, err := client.Do(req)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return err
	}

	// Record response status in the span
	span.SetAttributes(attribute.String("app-two.response.status", resp.Status))
	defer resp.Body.Close()

	return nil
}

// processRequest simulates an operation like a database call or external API request
func processRequest(ctx context.Context) error {
	// Start a child span
	_, span := tracer.Start(ctx, "db_lookup")
	defer span.End()

	// Simulate some processing time
	delay := time.Duration(rand.Intn(500)) * time.Millisecond
	time.Sleep(delay)

	// Randomly return an error to demonstrate error handling
	if rand.Float32() < 0.5 {
		return traceError("an error occurred in processRequest")
	}

	// Add some attributes to the span
	span.SetAttributes(attribute.String("process.status", "success"),
		attribute.Int("process.delay_ms", int(delay.Milliseconds())))

	return nil
}

func traceError(msg string) error {
	return &customError{msg}
}

type customError struct {
	msg string
}

func (e *customError) Error() string {
	return e.msg
}
