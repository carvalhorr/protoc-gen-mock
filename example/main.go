package main

import (
	"github.com/carvalhorr/protoc-gen-mock/bootstrap"
	"github.com/carvalhorr/protoc-gen-mock/grpchandler"
	"github.com/carvalhorr/protoc-gen-mock/stub"
	testservice "github.com/carvalhorr/protoc-gen-mock/test-service"
)

func main() {
	bootstrap.BootstrapServers(1068, 10010, MockServicesRegistersCallback)
}

var MockServicesRegistersCallback = func(stubsMatcher stub.StubsMatcher) []grpchandler.MockService {
	return []grpchandler.MockService{
		testservice.NewTestProtobufMockService(stubsMatcher),
	}
}
