package bottom

import (
	"context"
	bottomProto "github.com/mattfenwick/telemetry-hacking/pkg/bottom/protobuf"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"net/http"
	"sync"
	"time"
)

type Bottom struct {
	ThreadCount int
	Actions     chan func()
	Tracer      trace.Tracer

	bottomProto.UnimplementedBottomServer
}

func NewBottom(threadCount int, stop <-chan struct{}) *Bottom {
	actions := make(chan func())
	for i := 0; i < threadCount; i++ {
		go func() {
			for {
				select {
				case <-stop:
					return
				case f := <-actions:
					f()
				}
			}
		}()
	}
	return &Bottom{
		ThreadCount: threadCount,
		Actions:     actions,
		Tracer:      otel.Tracer("worker"),
	}
}

func reduce[A, B any](xs []A, f func(a A, b B) B, b B) B {
	out := b
	for _, x := range xs {
		out = f(x, out)
	}
	return out
}

type Number interface {
	int64 | float64 | int
}

func sum[A Number](nums []A) A {
	return reduce(nums, func(a A, b A) A { return a + b }, 0)
}

func product[A Number](nums []A) A {
	return reduce(nums, func(a A, b A) A { return a * b }, 1)
}

func runJob(name string, args []int) (int, error) {
	switch name {
	case "+", "add":
		return sum(args), nil
	case "*", "multiply":
		return product(args), nil
	case "sleep":
		duration := sum(args)
		time.Sleep(time.Duration(duration) * time.Millisecond)
		return duration, nil
	default:
		return 0, errors.Errorf("invalid operation: %s (args: %+v)", name, args)
	}
}

func myMap[A, B any](f func(a A) B, xs []A) []B {
	var out []B
	for _, x := range xs {
		out = append(out, f(x))
	}
	return out
}

func (w *Bottom) RunFunction(ctx context.Context, f *bottomProto.Function) (*bottomProto.FunctionResult, error) {
	result, err := w.RunFunctionHttp(ctx, &Function{
		Name: f.Name,
		Args: myMap(func(x int32) int { return int(x) }, f.Args),
	})
	if err != nil {
		return nil, err
	}
	return &bottomProto.FunctionResult{Value: int32(result.Value)}, nil
}

func (w *Bottom) RunFunctionHttp(ctx context.Context, f *Function) (*FunctionResult, error) {
	wg := sync.WaitGroup{}
	var result int
	var err error
	wg.Add(1)
	action := func() {
		result, err = runJob(f.Name, f.Args)
		wg.Done()
	}

	_, span := w.Tracer.Start(ctx, "run function")
	defer span.End()

	select {
	case w.Actions <- action:
		wg.Wait()
		span.AddEvent("finished function run")
		if err == nil {
			return &FunctionResult{Value: result}, nil
		} else {
			return nil, err
		}
	default:
		logrus.Warnf("worker service unavailable")
		span.SetStatus(codes.Error, "worker service unavailable")
		return nil, errors.Errorf("worker service unavailable")
	}
}

func (w *Bottom) NotFound(writer http.ResponseWriter, r *http.Request) {
	logrus.Errorf("HTTPResponder not found from request %+v", r)
	//recordHTTPNotFound(r) // TODO metrics
	http.NotFound(writer, r)
}

func (w *Bottom) Error(writer http.ResponseWriter, r *http.Request, err error, statusCode int) {
	logrus.Errorf("HTTPResponder error %s with code %d from request %+v", err.Error(), statusCode, r)
	//recordHTTPError(r, err, statusCode) // TODO metrics
	http.Error(writer, err.Error(), statusCode)
}
