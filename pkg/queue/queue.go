package queue

import (
	"context"
	"github.com/mattfenwick/telemetry-hacking/pkg/worker"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
	"net/http"
)

type Queue struct {
	//Actions      chan func()
	//Jobs         map[JobState][]*JobStatus
	Tracer       trace.Tracer
	WorkerClient *worker.Client
}

func NewQueue(stop <-chan struct{}, workerHost string, workerPort int) *Queue {
	return &Queue{
		//Actions: actionsChannel,
		//Jobs: map[JobState][]*JobStatus{
		//	JobStateError:      nil,
		//	JobStateInProgress: nil,
		//	JobStateSuccess:    nil,
		//	JobStateTodo:       nil,
		//},
		Tracer:       otel.Tracer("queue"),
		WorkerClient: worker.NewClient(workerHost, workerPort),
	}
}

//func (q *Queue) State(ctx context.Context) (*State, error) {
//	state := &State{Jobs: map[string][]*JobStatus{
//		JobStateError.String():      nil,
//		JobStateSuccess.String():    nil,
//	}}
//	wg := sync.WaitGroup{}
//	wg.Add(1)
//
//	span := trace.SpanFromContext(ctx)
//	span.AddEvent("enqueueing state action")
//
//	//utils.RunOperation(q.tracer, ctx, )
//	_, newSpan := q.Tracer.Start(ctx, "TODO get state")
//	defer newSpan.End()
//
//	q.Actions <- func() {
//		for status, jobs := range q.Jobs {
//			for _, job := range jobs {
//				state.Jobs[status.String()] = append(state.Jobs[status.String()], job)
//			}
//		}
//		wg.Done()
//	}
//	wg.Wait()
//	span.AddEvent("finished state action")
//	return state, nil
//}

// TODO start, finish, fail job

func (q *Queue) SubmitJob(ctx context.Context, job *JobRequest) (*JobResult, error) {
	span := trace.SpanFromContext(ctx)
	span.AddEvent("starting submitjob action")

	result, err := q.WorkerClient.RunFunction(&worker.Function{
		Name: job.Function,
		Args: job.Args,
	})
	if err != nil {
		return nil, err
	}
	return &JobResult{
		Request: job,
		Answer:  result.Value,
	}, nil
}

// NotFound logs the http client not found error
func (q *Queue) NotFound(w http.ResponseWriter, r *http.Request) {
	logrus.Errorf("HTTPResponder not found from request %+v", r)
	//recordHTTPNotFound(r) // TODO metrics
	http.NotFound(w, r)
}

// Error logs the http client errors
func (q *Queue) Error(w http.ResponseWriter, r *http.Request, err error, statusCode int) {
	logrus.Errorf("HTTPResponder error %s with code %d from request %+v", err.Error(), statusCode, r)
	//recordHTTPError(r, err, statusCode) // TODO metrics
	http.Error(w, err.Error(), statusCode)
}
