package queue

import (
	"context"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel/trace"
	"net/http"
	"sync"
	"time"
)

type Queue struct {
	Actions chan func()
	Jobs    map[JobState][]*JobStatus
}

func NewQueue(stop <-chan struct{}) *Queue {
	actionsChannel := make(chan func(), 20)
	go func() {
		for {
			select {
			case <-stop:
				return
			case f := <-actionsChannel:
				logrus.Debugf("handling queue action")
				f()
				time.Sleep(2 * time.Second)
			}
		}
	}()
	return &Queue{Actions: actionsChannel,
		Jobs: map[JobState][]*JobStatus{
			JobStateError:      nil,
			JobStateInProgress: nil,
			JobStateSuccess:    nil,
			JobStateTodo:       nil,
		}}
}

func (q *Queue) State(ctx context.Context) (*State, error) {
	state := &State{Jobs: map[string][]*JobStatus{
		JobStateError.String():      nil,
		JobStateInProgress.String(): nil,
		JobStateSuccess.String():    nil,
		JobStateTodo.String():       nil,
	}}
	wg := sync.WaitGroup{}
	wg.Add(1)

	span := trace.SpanFromContext(ctx)
	span.AddEvent("enqueueing state action")

	q.Actions <- func() {
		for status, jobs := range q.Jobs {
			for _, job := range jobs {
				state.Jobs[status.String()] = append(state.Jobs[status.String()], job)
			}
		}
		wg.Done()
	}
	wg.Wait()
	span.AddEvent("finished state action")
	return state, nil
}

// TODO start, finish, fail job

func (q *Queue) SubmitJob(ctx context.Context, job *JobRequest) (*JobStatus, error) {
	status := &JobStatus{
		Request: job,
		State:   JobStateTodo,
		Answer:  nil,
		Error:   "",
	}

	span := trace.SpanFromContext(ctx)
	span.AddEvent("enqueueing submitjob action")

	q.Actions <- func() {
		q.Jobs[JobStateTodo] = append(q.Jobs[JobStateTodo], status)
	}
	span.AddEvent("finished submitjob action")
	return status, nil
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
