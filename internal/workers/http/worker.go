package http

import (
	"context"
	"log"
	"net"
	"net/http"
	"strconv"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	api "gitlab.ozon.dev/dev.gulkoalexey/gulko-alexey/gen/proto/go"
	"gitlab.ozon.dev/dev.gulkoalexey/gulko-alexey/internal/logger"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type Worker struct {
	cfg         Config
	UserStorage UserStorage
	tg          MessageSender
}

type Config interface {
	BotGrpcPort() int64
	BotHttpPort() int64
}

func New(cfg Config, storage UserStorage, tg MessageSender) *Worker {
	return &Worker{cfg: cfg, UserStorage: storage, tg: tg}
}

func (w *Worker) Run(ctx context.Context) {
	grpcListener, err := net.Listen("tcp", ":"+strconv.Itoa(int(w.cfg.BotGrpcPort())))
	if err != nil {
		log.Fatalf("failed to listen grpc: %v", err)
	}

	s := grpc.NewServer(grpc.UnaryInterceptor(HiMarkInterceptor))
	api.RegisterReporterServer(s, NewServer(w.tg, w.UserStorage))

	rmux := runtime.NewServeMux()
	mux := http.NewServeMux()
	mux.Handle("/", rmux)
	mux.Handle("/metrics", promhttp.Handler())

	err = api.RegisterReporterHandlerServer(ctx, rmux, Server{})
	if err != nil {
		log.Fatal(err)
	}

	httpListener, err := net.Listen("tcp", ":"+strconv.Itoa(int(w.cfg.BotHttpPort())))
	if err != nil {
		log.Fatalf("failed to listen http: %v", err)
	}

	go func() {
		logger.Info("start serve http")
		for {
			select {
			case <-ctx.Done():
				logger.Info("listen http stopped")
				return
			default:
				err = http.Serve(httpListener, mux)
				if err != nil {
					logger.Fatal("error starting http server", zap.Error(err))
					return
				}
			}
		}
	}()

	go func() {
		logger.Info("start serve grpc")
		for {
			select {
			case <-ctx.Done():
				logger.Info("listen grpc stopped")
				return
			default:
				reflection.Register(s)
				if err := s.Serve(grpcListener); err != nil {
					logger.Fatal("error starting grpc server", zap.Error(err))
				}
			}
		}
	}()
}

func HiMarkInterceptor(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (resp interface{}, err error) {
	logger.Info("Oh Hi Mark!")
	m, err := handler(ctx, req)
	return m, err
}
