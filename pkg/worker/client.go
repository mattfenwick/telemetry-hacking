package worker

import (
	"bytes"
	"context"
	"fmt"
	"github.com/mattfenwick/telemetry-hacking/pkg/utils"
	"net/http"
	"strings"

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
		HttpClient: http.Client{Transport: otelhttp.NewTransport(http.DefaultTransport)},
		ServerURL:  fmt.Sprintf("http://%s:%d", serverHost, serverPort),
		Tracer:     otel.Tracer("queue/client"),
	}
}

func (c *Client) RunFunction(f *Function) (*FunctionResult, error) {
	makeRequest := func(ctx context.Context) *http.Request {
		request, err := http.NewRequestWithContext(
			ctx,
			"POST",
			strings.Join([]string{c.ServerURL, "function"}, "/"),
			bytes.NewBuffer([]byte(utils.DumpJSON(f))))
		utils.DoOrDie(err)
		return request
	}
	text, err := utils.IssueRequest(c.HttpClient, makeRequest, context.Background(), c.Tracer)
	if err != nil {
		return nil, err
	}
	result := FunctionResult{}
	err = utils.ParseJson(&result, []byte(text))
	return &result, err
}
