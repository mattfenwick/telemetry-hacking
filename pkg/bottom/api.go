package bottom

import (
	"context"
	"fmt"
	"github.com/mattfenwick/telemetry-hacking/pkg/utils"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/trace"
	"io/ioutil"
	"net/http"

	"github.com/sirupsen/logrus"
)

type Function struct {
	Name string
	Args []int
}

type FunctionResult struct {
	Value int
}

type Responder interface {
	RunFunction(ctx context.Context, function *Function) (*FunctionResult, int, error)

	NotFound(w http.ResponseWriter, r *http.Request)
	Error(w http.ResponseWriter, r *http.Request, err error, statusCode int)
}

func SetupHTTPServer(responder Responder) {
	handleJob := func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		span := trace.SpanFromContext(ctx)
		span.AddEvent("handling-function")

		logrus.Infof("handling function request")
		switch r.Method {
		case "POST":
			body, err := ioutil.ReadAll(r.Body)
			if err != nil {
				logrus.Errorf("unable to read body for RunFunction POST: %s", err.Error())
				responder.Error(w, r, err, 400)
				return
			}
			var f Function
			err = utils.ParseJson(&f, body)
			if err != nil {
				logrus.Errorf("unable to ummarshal JSON for RunFunction POST: %s", err.Error())
				responder.Error(w, r, err, 400)
				return
			}
			jobStatus, code, err := responder.RunFunction(ctx, &f)
			header := w.Header()
			header.Set(http.CanonicalHeaderKey("content-type"), "application/json")
			if code < 200 || code > 299 {
				http.Error(w, err.Error(), code)
			} else {
				fmt.Fprint(w, utils.DumpJSON(jobStatus))
			}
		default:
			responder.NotFound(w, r)
		}
	}
	http.Handle("/function", otelhttp.NewHandler(http.HandlerFunc(handleJob), "function"))
}
