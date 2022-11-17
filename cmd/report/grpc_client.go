package main

import (
	"context"
	"log"
	"strconv"
	"time"

	api "gitlab.ozon.dev/dev.gulkoalexey/gulko-alexey/gen/proto/go"
	"gitlab.ozon.dev/dev.gulkoalexey/gulko-alexey/internal/logger"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Connection struct {
	Conn *grpc.ClientConn
	cfg  Config
}

func NewClient(cfg Config) *Connection {
	return &Connection{Conn: nil, cfg: cfg}
}

func (c *Connection) GetConnection() *grpc.ClientConn {
	return c.Conn
}

func (c *Connection) startGrpcClient(ctx context.Context) {
	conn, err := grpc.Dial(
		c.cfg.BotHost()+":"+strconv.Itoa(int(c.cfg.BotGrpcPort())),
		grpc.WithUnaryInterceptor(SimpleUnaryInterceptor),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)

	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}

	c.Conn = conn

	go func() {
		<-ctx.Done()
		err = conn.Close()
		if err != nil {
			logger.Error("close grpc client", zap.Error(err))
		}
	}()
}

type ReporterClient struct {
	api.ReporterClient
}

func NewReporterClient(conn *Connection) *ReporterClient {
	return &ReporterClient{
		ReporterClient: api.NewReporterClient(conn.GetConnection()),
	}
}

func (c ReporterClient) SendReport(ctx context.Context, request *api.ReportRequest) *api.ReportResponse {
	ctx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()

	r, err := c.ReporterClient.SendReport(ctx, request)

	if err != nil {
		logger.Error("could not send", zap.Error(err))
	}

	logger.Info("response", zap.String("k", r.String()))
	return r
}
