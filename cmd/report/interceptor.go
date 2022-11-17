package main

import (
	"gitlab.ozon.dev/dev.gulkoalexey/gulko-alexey/internal/logger"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

func SimpleUnaryInterceptor(ctx context.Context, method string, req interface{},
	reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
	logger.Info("I Did Not Hit Her. I Did Not.")
	return invoker(ctx, method, req, reply, cc, opts...)
}
