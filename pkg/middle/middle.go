package middle

import (
	"context"
	"github.com/mattfenwick/telemetry-hacking/pkg/bottom"
	"github.com/mattfenwick/telemetry-hacking/pkg/utils"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
	"net/http"
)

type Middle struct {
	//Actions      chan func()
	//Jobs         map[JobState][]*JobStatus
	//Tracer       trace.Tracer
	BottomClient     *bottom.Client
	BottomGRPCClient *bottom.GRPCClient
}

func NewMiddle(stop <-chan struct{}, workerHost string, workerPort int, grpcPort int) *Middle {
	grpcClient, err := bottom.NewGRPCClient(otel.Tracer("bottom/grpc-client"), workerHost, grpcPort)
	utils.DoOrDie(err)
	return &Middle{
		//Actions: actionsChannel,
		//Jobs: map[JobState][]*JobStatus{
		//	JobStateError:      nil,
		//	JobStateInProgress: nil,
		//	JobStateSuccess:    nil,
		//	JobStateTodo:       nil,
		//},
		//Tracer:       otel.Tracer("queue"),
		BottomClient:     bottom.NewClient(otel.Tracer("bottom/http-client"), workerHost, workerPort),
		BottomGRPCClient: grpcClient,
	}
}

//func (q *Middle) State(ctx context.Context) (*State, error) {
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

func (q *Middle) SubmitJobGRPC(ctx context.Context, job *JobRequest) (*JobResult, error) {
	span := trace.SpanFromContext(ctx)
	span.AddEvent("starting submitjob-grpc action")

	result, err := q.BottomGRPCClient.RunFunction(ctx, &bottom.Function{
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

func (q *Middle) SubmitJobHttp(ctx context.Context, job *JobRequest) (*JobResult, error) {
	span := trace.SpanFromContext(ctx)
	span.AddEvent("starting submitjob-http action")

	result, err := q.BottomClient.RunFunction(ctx, &bottom.Function{
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

func (q *Middle) SubmitJob(ctx context.Context, job *JobRequest) (*JobResult, error) {
	span := trace.SpanFromContext(ctx)
	span.AddEvent("starting submitjob action")

	grpcResult, err := q.SubmitJobGRPC(ctx, job)
	if err != nil {
		return nil, err
	}
	httpResult, err := q.SubmitJobHttp(ctx, job)
	if err != nil {
		return nil, err
	}

	if grpcResult.Answer != httpResult.Answer {
		return nil, errors.Errorf("got different answers from grpc and http: %d vs %d", grpcResult.Answer, httpResult.Answer)
	}

	return grpcResult, nil
}

func (q *Middle) NotFound(w http.ResponseWriter, r *http.Request) {
	logrus.Errorf("HTTPResponder not found from request %+v", r)
	//recordHTTPNotFound(r) // TODO metrics
	http.NotFound(w, r)
}

func (q *Middle) Error(w http.ResponseWriter, r *http.Request, err error, statusCode int) {
	logrus.Errorf("HTTPResponder error %s with code %d from request %+v", err.Error(), statusCode, r)
	//recordHTTPError(r, err, statusCode) // TODO metrics
	http.Error(w, err.Error(), statusCode)
}
