package queue

import (
	"bytes"
	"context"
	"fmt"
	"github.com/mattfenwick/telemetry-hacking/pkg/utils"
	"net/http"
	"strings"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

type Client struct {
	HttpClient http.Client
	ServerURL  string
	Tracer     trace.Tracer
}

func NewClient(serverHost string, serverPort int) *Client {
	return &Client{
		HttpClient: http.Client{Transport: otelhttp.NewTransport(transport())},
		ServerURL:  fmt.Sprintf("http://%s:%d", serverHost, serverPort),
		Tracer:     otel.Tracer("queue/client"),
	}
}

func transport() *http.Transport {
	return &http.Transport{
		Proxy:               http.ProxyFromEnvironment,
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 100,
		IdleConnTimeout:     90 * time.Second,
	}
}

func (c *Client) SubmitJob(methodContext context.Context, job *JobRequest) (*JobResult, error) {
	makeRequest := func(ctx context.Context) *http.Request {
		request, err := http.NewRequestWithContext(
			ctx,
			"POST",
			strings.Join([]string{c.ServerURL, "job"}, "/"),
			bytes.NewBuffer([]byte(utils.DumpJSON(job))))
		utils.DoOrDie(err)
		return request
	}
	text, err := utils.IssueRequest(c.HttpClient, makeRequest, methodContext, c.Tracer)
	if err != nil {
		return nil, err
	}
	status := JobResult{}
	err = utils.ParseJson(&status, []byte(text))
	return &status, err
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
