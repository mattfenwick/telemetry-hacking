package worker

import (
	"bytes"
	"context"
	"fmt"
	"github.com/mattfenwick/telemetry-hacking/pkg/utils"
	"net/http"
	"strings"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"

	"go.opentelemetry.io/otel/trace"
)

type Client struct {
	HttpClient http.Client
	ServerURL  string
	Tracer     trace.Tracer
}

func NewClient(tracer trace.Tracer, serverHost string, serverPort int) *Client {
	return NewClientFromURL(tracer, fmt.Sprintf("http://%s:%d", serverHost, serverPort))
}

func NewClientFromURL(tracer trace.Tracer, serverURL string) *Client {
	return &Client{
		HttpClient: http.Client{Transport: otelhttp.NewTransport(transport())},
		ServerURL:  serverURL,
		Tracer:     tracer,
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

func (c *Client) RunFunction(methodContext context.Context, f *Function) (*FunctionResult, error) {
	makeRequest := func(ctx context.Context) *http.Request {
		request, err := http.NewRequestWithContext(
			ctx,
			"POST",
			strings.Join([]string{c.ServerURL, "function"}, "/"),
			bytes.NewBuffer([]byte(utils.DumpJSON(f))))
		utils.DoOrDie(err)
		return request
	}
	text, err := utils.IssueRequest(c.HttpClient, makeRequest, methodContext, c.Tracer)
	if err != nil {
		return nil, err
	}
	result := FunctionResult{}
	err = utils.ParseJson(&result, []byte(text))
	return &result, err
}
