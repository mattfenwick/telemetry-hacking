package utils

import (
	"context"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/sdk/resource"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.10.0"
	"go.opentelemetry.io/otel/trace"
)

func SetUpJaegerTracerProvider(aggregatorURL string, service string) (*tracesdk.TracerProvider, error) {
	logrus.Infof("setting up jaeger at %s for service %s", aggregatorURL, service)

	exporter, err := jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(aggregatorURL)))
	if err != nil {
		return nil, errors.Wrapf(err, "unable to instantiate jaeger")
	}
	tracerProvider := tracesdk.NewTracerProvider(
		// Always be sure to batch in production.
		tracesdk.WithBatcher(exporter),
		tracesdk.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String(service),
			//attribute.Int64("ID", id),
		)),
	)

	otel.SetTracerProvider(tracerProvider)

	return tracerProvider, nil
}

func RunOperation(ctx context.Context, spanName string, action func(span trace.Span) error) {
	tr := otel.Tracer("TODO -- what should this be?")
	// TODO -- add span config
	_, span := tr.Start(ctx, spanName)
	//span.SetAttributes(attribute.Key("TODO").String("TODO"))
	defer span.End()

	err := action(span)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "TODO -- failed")
	}
}
