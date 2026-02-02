package aspecto

import (
	"context"
	"log"
	"strings"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	"go.opentelemetry.io/otel/trace"
)

// Initialize initializes OpenTelemetry with Aspecto exporter
func AspectoTraceProvider(serviceName, token string) (*sdktrace.TracerProvider, error) {
	ctx := context.Background()
	exp, err := otlptracehttp.New(
		ctx,
		otlptracehttp.WithEndpoint("collector.aspecto.io"),
		otlptracehttp.WithHeaders(map[string]string{"Authorization": token}),
	)

	if err != nil {
		return nil, err
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(&CustomSampler{}),
		sdktrace.WithBatcher(exp),
		sdktrace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String(serviceName),
		)),
	)

	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))
	log.Println("Aspecto Tracer initialized")
	return tp, nil
}

// CustomSampler implements the sdktrace.Sampler interface.
type CustomSampler struct{}

func (cs *CustomSampler) ShouldSample(p sdktrace.SamplingParameters) sdktrace.SamplingResult {
	var path string
	for _, kv := range p.Attributes {
		if kv.Key == semconv.HTTPRouteKey {
			path = kv.Value.AsString()
			break
		}
	}

	if strings.Contains(path, "health") || path == "" {
		return sdktrace.SamplingResult{
			Decision:   sdktrace.Drop,
			Tracestate: trace.SpanContextFromContext(p.ParentContext).TraceState(),
		}
	}
	return sdktrace.SamplingResult{
		Decision:   sdktrace.RecordAndSample,
		Tracestate: trace.SpanContextFromContext(p.ParentContext).TraceState(),
	}
}

func (as *CustomSampler) Description() string {
	return "CustomSampler"
}
