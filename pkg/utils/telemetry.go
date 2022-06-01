package utils

import (
	"context"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/exporters/jaeger"
	stdout "go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.10.0"
	"go.opentelemetry.io/otel/trace"
)

func initTracer() (*tracesdk.TracerProvider, error) {
	exporter, err := stdout.New(stdout.WithPrettyPrint())
	if err != nil {
		return nil, err
	}

	tp := tracesdk.NewTracerProvider(
		tracesdk.WithSampler(tracesdk.AlwaysSample()),
		tracesdk.WithBatcher(exporter),
	)
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))
	return tp, nil
}

func SetUpTracerProvider(aggregatorURL string, service string) (*tracesdk.TracerProvider, error) {
	var tracerProvider *tracesdk.TracerProvider
	if aggregatorURL == "" {
		logrus.Infof("aggregator url is empty, so instantiating stdout tracer provider")

		exporter, err := stdout.New(stdout.WithPrettyPrint())
		if err != nil {
			return nil, errors.Wrapf(err, "unable to instantiate stdout tracer provider")
		}

		tracerProvider = tracesdk.NewTracerProvider(
			tracesdk.WithSampler(tracesdk.AlwaysSample()),
			tracesdk.WithBatcher(exporter),
		)
	} else {
		logrus.Infof("setting up jaeger at %s for service %s", aggregatorURL, service)

		exporter, err := jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(aggregatorURL)))
		if err != nil {
			return nil, errors.Wrapf(err, "unable to instantiate jaeger tracer provider")
		}
		tracerProvider = tracesdk.NewTracerProvider(
			tracesdk.WithSampler(tracesdk.AlwaysSample()), // TODO if prod, need to batch or something?
			tracesdk.WithBatcher(exporter),
			tracesdk.WithResource(resource.NewWithAttributes(
				semconv.SchemaURL,
				semconv.ServiceNameKey.String(service),
				//attribute.Int64("ID", id),
			)),
		)
	}

	otel.SetTracerProvider(tracerProvider)
	// TODO what does this do?
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))

	return tracerProvider, nil
}

func RunOperation(tracer trace.Tracer, ctx context.Context, spanName string, action func(span trace.Span) error, opts ...trace.SpanStartOption) {
	_, span := tracer.Start(ctx, spanName, opts...)
	defer span.End()

	err := action(span)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "TODO -- failed")
	}
}
