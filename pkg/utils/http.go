package utils

import (
	"context"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"net/http/httptrace"

	"go.opentelemetry.io/contrib/instrumentation/net/http/httptrace/otelhttptrace"
	semconv "go.opentelemetry.io/otel/semconv/v1.10.0"
	"go.opentelemetry.io/otel/trace"
)

func IssueRequest(client http.Client, makeRequest func(ctx context.Context) *http.Request, rootContext context.Context, tracer trace.Tracer) (string, error) {
	spanContext, span := tracer.Start(rootContext, "queue/request", trace.WithAttributes(semconv.PeerServiceKey.String("ExampleService")))
	defer span.End()

	httpContext := httptrace.WithClientTrace(spanContext, otelhttptrace.NewClientTrace(spanContext))
	request := makeRequest(httpContext)

	logrus.Infof("issuing %s request to %s", request.Method, request.URL)

	response, err := client.Do(request)
	if err != nil {
		return "", errors.Wrapf(err, "unable to issue http request: %+v", err)
	}

	logrus.Debugf("response code: %d", response.StatusCode)

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return "", errors.Errorf("unable to ReadAll from response body")
	}

	logrus.Debugf("received response: %s", string(body))

	if response.StatusCode < 200 || response.StatusCode > 299 {
		return string(body), errors.Errorf("request failed with %d", response.StatusCode)
	}

	return string(body), errors.Wrapf(response.Body.Close(), "unable to close body")
}
