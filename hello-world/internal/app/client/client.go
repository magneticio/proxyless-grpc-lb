package client

import (
	"os"
	"os/signal"
	"strings"
	"syscall"

	"time"

	"go.uber.org/zap"

	"golang.org/x/net/context"
	"google.golang.org/grpc"

	logger "github.com/asishrs/proxyless-grpc-lb/common/pkg/logger"
	hello "github.com/asishrs/proxyless-grpc-lb/hello-world/internal/app/http/rpc"
	helper "github.com/asishrs/proxyless-grpc-lb/hello-world/internal/pkg"
	_ "google.golang.org/grpc/xds"
)

const ()

var (
	conn *grpc.ClientConn
)

func StartClient(server string) {
	stop := make(chan os.Signal)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)

	logger.Logger.Info("Starting gRPC Client", zap.String("host", server))

	var opts []grpc.DialOption
	opts = append(opts, grpc.WithInsecure())

	conn, err := grpc.Dial(server, opts...)

	if err != nil {
		logger.Logger.Fatal("Unable to start", zap.Error(err))
	}

	go helper.ShutdownClient(stop, conn)

	c := hello.NewHelloClient(conn)
	ctx := context.Background()

	count := uint64(0)
	service1 := uint64(0)
	service2 := uint64(0)

	for {
		msg, err := c.SayHello(ctx, &hello.HelloRequest{Name: "gRPC Proxyless LB"})
		if err != nil {
			logger.Logger.Error("Unable to send Hello message", zap.Any("Response", msg), zap.Error(err))
		} else {
			logger.Logger.Info("Hello Response", zap.Any("Response", msg))

			count++

			if strings.Contains(msg.Message, "hello-server-1") {
				service1++
			} else {
				service2++
			}

			if count%10 == 0 {

				percentage1 := (service1 * 100) / count
				percentage2 := (service2 * 100) / count

				logger.Logger.Error("Current percentages", zap.Any("hello-server-1", percentage1), zap.Any("hello-server-2", percentage2))

			}
		}
		time.Sleep(3 * time.Second)
	}
}
