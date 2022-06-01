package queue

import (
	"bytes"
	"context"
	"fmt"
	"github.com/mattfenwick/telemetry-hacking/pkg/utils"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"net/http/httptrace"
	"strings"

	"go.opentelemetry.io/contrib/instrumentation/net/http/httptrace/otelhttptrace"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"

	"go.opentelemetry.io/otel"
	semconv "go.opentelemetry.io/otel/semconv/v1.10.0"
	"go.opentelemetry.io/otel/trace"
)

type Client struct {
	HttpClient http.Client
	ServerURL  string
	Tracer     trace.Tracer
}

func NewClient(serverHost string, serverPort int) *Client {
	return &Client{
		HttpClient: http.Client{Transport: otelhttp.NewTransport(http.DefaultTransport)},
		ServerURL:  fmt.Sprintf("http://%s:%d", serverHost, serverPort),
		Tracer:     otel.Tracer("queue/client"),
	}
}

func (c *Client) GetState() (*State, error) {
	// TODO add baggage?
	//bag, _ := baggage.Parse("username=donuts")
	//ctx := baggage.ContextWithBaggage(context.Background(), bag)

	makeRequest := func(ctx context.Context) *http.Request {
		request, err := http.NewRequestWithContext(ctx, "GET", strings.Join([]string{c.ServerURL, "state"}, "/"), nil)
		utils.DoOrDie(err)
		return request
	}
	text, err := IssueRequest(c.HttpClient, makeRequest, context.Background(), c.Tracer)
	if err != nil {
		return nil, err
	}
	state := State{}
	err = utils.ParseJson(&state, []byte(text))
	return &state, err
}

func (c *Client) SubmitJob(job *JobRequest) (*JobStatus, error) {
	makeRequest := func(ctx context.Context) *http.Request {
		request, err := http.NewRequestWithContext(
			ctx,
			"POST",
			strings.Join([]string{c.ServerURL, "job"}, "/"),
			bytes.NewBuffer([]byte(utils.DumpJSON(job))))
		utils.DoOrDie(err)
		return request
	}
	text, err := IssueRequest(c.HttpClient, makeRequest, context.Background(), c.Tracer)
	if err != nil {
		return nil, err
	}
	status := JobStatus{}
	err = utils.ParseJson(&status, []byte(text))
	return &status, err
}

func IssueRequest(client http.Client, makeRequest func(ctx context.Context) *http.Request, rootContext context.Context, tracer trace.Tracer) (string, error) {
	spanContext, span := tracer.Start(rootContext, "say hello", trace.WithAttributes(semconv.PeerServiceKey.String("ExampleService")))
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

	return string(body), errors.Wrapf(response.Body.Close(), "unable to close body")
}

/*
func main() {
	tr := otel.Tracer("example/client")
	err = func(rootContext context.Context) error {
		spanContext, span := tr.Start(rootContext, "say hello", trace.WithAttributes(semconv.PeerServiceKey.String("ExampleService")))
		defer span.End()

		httpContext = httptrace.WithClientTrace(spanContext, otelhttptrace.NewClientTrace(spanContext))
		req, _ := http.NewRequestWithContext(httpContext, "GET", *url, nil)

		fmt.Printf("Sending request...\n")
		res, err := client.Do(req)
		if err != nil {
			panic(err)
		}
		body, err = ioutil.ReadAll(res.Body)
		_ = res.Body.Close()

		return err
	}(ctx)

	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Response Received: %s\n\n\n", body)
	fmt.Printf("Waiting for few seconds to export spans ...\n\n")
	time.Sleep(10 * time.Second)
	fmt.Printf("Inspect traces on stdout\n")
}
*/
