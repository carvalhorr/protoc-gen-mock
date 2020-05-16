package bootstrap

import (
	"github.com/carvalhorr/protoc-gen-mock/grpchandler"
	"github.com/carvalhorr/protoc-gen-mock/stub"
	log "github.com/sirupsen/logrus"
	"strings"
)

// BootstrapServers starts the gRPC server with the mock services added by serviceregistersCallback.
// The REST server for the stub API management is also started.
// Parameters:
// - tmpPath : temporary path to store temporary files
// - restPort : the port where the REST server will be started
// - grpcPort : the port where the gRPC server will be started
// - servicesRegistrationCallback : a function called when the grpc server is ready so that the mock services can be registered
func BootstrapServers(tmpPath string, restPort uint, grpcPort uint, serviceRegisterCallback func(stubsStore stub.StubsMatcher) grpchandler.MockService) {
	setupLogrus()

	errorsEngine, err := stub.NewCustomErrorEngine(tmpPath)
	if err != nil {
		panic(err)
	}
	stub.SetErrorEngine(errorsEngine)

	stubsStore := stub.NewInMemoryStubsStore()
	stubsMatcher := stub.NewStubsMatcher(stubsStore)

	service := serviceRegisterCallback(stubsMatcher)
	log.Info("Supported methods: ", strings.Join(service.GetSupportedMethods(), "  |  "))
	stubsExamples := service.GetPayloadExamples()
	go StartRESTServer(restPort, CreateRESTControllers(stubsExamples, stubsStore, service))
	StarGRPCServer(grpcPort, service)
}

func setupLogrus() {
	log.SetFormatter(&log.TextFormatter{
		FullTimestamp: true,
	})
	log.SetLevel(log.DebugLevel)
}
