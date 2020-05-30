package main

import (
	"context"
	"fmt"
	greetermock "github.com/carvalhorr/protoc-gen-mock/greeter-service"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
	"net"
	"os"
	"os/signal"
	"syscall"
)

var server *grpc.Server
var listener net.Listener

func main() {
	server = grpc.NewServer()
	reflection.Register(server)
	greetermock.RegisterGreeterServer(server, &greeterService{})
	var err error
	addr := fmt.Sprintf("0.0.0.0:%d", 50010)
	listener, err = net.Listen("tcp", addr)

	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}
	go serv(listener)

	if err != nil {
		log.Fatalf("GRPC server failed to start %+v", err)
	}
	AwaitTermination(func() {
		log.Warn("Shutting down the server")
	})
}

func AwaitTermination(shutdownHook func()) {
	interruptSignal := make(chan os.Signal, 1)
	signal.Notify(interruptSignal, syscall.SIGINT, syscall.SIGTERM)
	<-interruptSignal
	cleanup()
	if shutdownHook != nil {
		shutdownHook()
	}
}

func cleanup() {
	log.Info("Stopping the server")
	server.GracefulStop()
	log.Info("Closing the listener")
	listener.Close()
	log.Info("End of Program")
}

func serv(listener net.Listener) {
	if err := server.Serve(listener); err != nil {
		log.Errorf("failed to serve: %v", err)
	}
}

type greeterService struct{}

func (s greeterService) Hello(ctx context.Context, in *greetermock.Request) (*greetermock.Response, error) {
	//return &greetermock.Response{Greeting: fmt.Sprintf("Hello, %s", in.Name)}, nil
	return nil, status.Error(codes.FailedPrecondition, "an internal error blah blah")
}
