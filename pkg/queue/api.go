package queue

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

type JobState string

const (
	JobStateTodo       JobState = "JobStateTodo"
	JobStateInProgress JobState = "JobStateInProgress"
	JobStateError      JobState = "JobStateError"
	JobStateSuccess    JobState = "JobStateSuccess"
)

func (s JobState) String() string {
	return string(s)
}

type State struct {
	Jobs map[string][]*JobStatus
}

type JobStatus struct {
	Request *JobRequest
	State   JobState
	Answer  any
	Error   string
}

type JobRequest struct {
	JobId    string
	Function string
	Args     []string
}

type Responder interface {
	State(ctx context.Context) (*State, error)
	SubmitJob(ctx context.Context, job *JobRequest) (*JobStatus, error)

	NotFound(w http.ResponseWriter, r *http.Request)
	Error(w http.ResponseWriter, r *http.Request, err error, statusCode int)
}

func SetupHTTPServer(responder Responder) {

	handleState := func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		span := trace.SpanFromContext(ctx)
		span.AddEvent("handling-state")

		logrus.Infof("handling state request")
		if r.Method == "GET" {
			state, err := responder.State(ctx)
			if err != nil {
				responder.Error(w, r, err, 500)
				return
			}
			header := w.Header()
			header.Set(http.CanonicalHeaderKey("content-type"), "application/json")
			fmt.Fprint(w, utils.DumpJSON(state))
		} else {
			responder.NotFound(w, r)
		}
	}
	http.Handle("/state", otelhttp.NewHandler(http.HandlerFunc(handleState), "wrapper-state"))

	handleJob := func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		span := trace.SpanFromContext(ctx)
		span.AddEvent("handling-job")

		logrus.Infof("handling job request")
		switch r.Method {
		case "POST":
			body, err := ioutil.ReadAll(r.Body)
			if err != nil {
				logrus.Errorf("unable to read body for AddJob POST: %s", err.Error())
				responder.Error(w, r, err, 400)
				return
			}
			var job JobRequest
			err = utils.ParseJson(&job, body)
			if err != nil {
				logrus.Errorf("unable to ummarshal JSON for AddJob POST: %s", err.Error())
				responder.Error(w, r, err, 400)
				return
			}
			jobStatus, err := responder.SubmitJob(ctx, &job)
			if err != nil {
				responder.Error(w, r, err, 500)
				return
			}
			header := w.Header()
			header.Set(http.CanonicalHeaderKey("content-type"), "application/json")
			fmt.Fprint(w, utils.DumpJSON(jobStatus))
		default:
			responder.NotFound(w, r)
		}
	}
	http.Handle("/job", otelhttp.NewHandler(http.HandlerFunc(handleJob), "wrapper-job"))
}
