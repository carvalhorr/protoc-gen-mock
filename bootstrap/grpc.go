package bootstrap

import (
	"fmt"
	"github.com/carvalhorr/protoc-gen-mock/grpchandler"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
	"net"
	"os"
	"os/signal"
	"syscall"
)

// var server grpc_server.GrpcServer
var server *grpc.Server
var listener net.Listener

// Start the server for the previously registered services
func StarGRPCServer(port uint, services []grpchandler.MockService) {

	server = grpc.NewServer()
	grpc_health_v1.RegisterHealthServer(server, health.NewServer())
	reflection.Register(server)

	for _, service := range services {
		service.Register(server)
	}

	var err error
	addr := fmt.Sprintf("0.0.0.0:%d", port)
	listener, err = net.Listen("tcp", addr)

	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}
	log.Infof("gRPC Server listening on port: %d", port)
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
