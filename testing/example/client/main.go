package main

import (
	"github.com/thebaasco/protoc-gen-mock/bootstrap"
	greetermock "github.com/thebaasco/protoc-gen-mock/greeter-service"
	"github.com/thebaasco/protoc-gen-mock/grpchandler"
	"github.com/thebaasco/protoc-gen-mock/stub"
)

func main() {
	bootstrap.BootstrapServers("./tmp/", 1068, 10010, MockServicesRegistersCallback)
}

var MockServicesRegistersCallback = func(stubsMatcher stub.StubsMatcher) grpchandler.MockService {
	return grpchandler.NewCompositeMockService([]grpchandler.MockService{ // Use CompositeMockService to add multiple mock services, otherwise you can return a single mock service.
		greetermock.NewGreeterMockService(stubsMatcher),
	})
}
