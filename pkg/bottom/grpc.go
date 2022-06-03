package bottom

import (
	"context"
	"fmt"
	bottomProto "github.com/mattfenwick/telemetry-hacking/pkg/bottom/protobuf"
	"github.com/mattfenwick/telemetry-hacking/pkg/utils"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
	"net"
	"time"
)

type GRPCServer struct {
	Port   int
	Server *grpc.Server
}

func NewGRPCServer(port int, bottom *Bottom) (*GRPCServer, error) {
	g := &GRPCServer{Port: port}

	interceptor := grpc.UnaryInterceptor(func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		logrus.Infof("received %s request: ", info.FullMethod)
		start := time.Now()

		// awkward "composition" of unary interceptors: see https://github.com/grpc/grpc-go/issues/935
		resp, err = otelgrpc.UnaryServerInterceptor()(ctx, req, info, handler)

		// parse grpc status code
		code := codes.OK
		if err != nil {
			st, ok := status.FromError(err)
			if ok {
				code = st.Code()
			} else {
				code = codes.Unknown
			}
		}

		utils.RecordEventDuration(info.FullMethod, int(code), start)
		logrus.Infof("finished handling %s request: (%+v, %+v)", info.FullMethod, start, code)

		return resp, err
	})

	g.Server = grpc.NewServer(
		interceptor,
		grpc.StreamInterceptor(otelgrpc.StreamServerInterceptor()))

	bottomProto.RegisterBottomServer(g.Server, bottom)
	reflection.Register(g.Server)

	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", g.Port))
	if err != nil {
		return nil, errors.Wrapf(err, "unable to listen on port %d", g.Port)
	}

	go func() {
		utils.DoOrDie(g.Server.Serve(listener))
	}()

	return g, nil
}

type GRPCClient struct {
	Connection *grpc.ClientConn
	Client     bottomProto.BottomClient
}

func NewGRPCClient(tracer trace.Tracer, serverHost string, serverPort int) (*GRPCClient, error) {
	address := fmt.Sprintf("%s:%d", serverHost, serverPort)
	connection, err := grpc.Dial(address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithUnaryInterceptor(otelgrpc.UnaryClientInterceptor()),
		grpc.WithStreamInterceptor(otelgrpc.StreamClientInterceptor()))
	if err != nil {
		return nil, errors.Wrapf(err, "unable to connect to grpc server")
	}
	return &GRPCClient{Connection: connection, Client: bottomProto.NewBottomClient(connection)}, nil
}

func (g *GRPCClient) RunFunction(methodContext context.Context, f *Function) (*FunctionResult, error) {
	result, err := g.Client.RunFunction(methodContext, &bottomProto.Function{
		Name: f.Name,
		Args: myMap(func(x int) int32 { return int32(x) }, f.Args),
	})
	if err != nil {
		return nil, errors.Wrapf(err, "unable to run grpc method")
	}
	return &FunctionResult{Value: int(result.Value)}, nil
}
