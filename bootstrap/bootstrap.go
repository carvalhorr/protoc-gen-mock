package bootstrap

import (
	"github.com/carvalhorr/protoc-gen-mock/grpchandler"
	"github.com/carvalhorr/protoc-gen-mock/stub"
	log "github.com/sirupsen/logrus"
	"strings"
)

func BootstrapServers(tmpPath string, restPort uint, grpcPort uint, servicesRegistersCallback func(stubsStore stub.StubsMatcher) []grpchandler.MockService) {
	setupLogrus()
	stubsStore := stub.NewInMemoryStubsStore()
	errorsEngine, err := stub.NewCustomErrorEngine(tmpPath)
	if err != nil {
		panic(err)
	}
	stubsMatcher := stub.NewStubsMatcher(stubsStore, errorsEngine)
	services := servicesRegistersCallback(stubsMatcher)
	var supportedFullMethodNames = make([]string, 0)
	for _, service := range services {
		supportedFullMethodNames = append(supportedFullMethodNames, service.GetSupportedMethods()...)
	}
	log.Info("Supported methods: ", strings.Join(supportedFullMethodNames, "  |  "))
	var stubsExamples = make([]stub.Stub, 0)
	for _, service := range services {
		stubsExamples = append(stubsExamples, service.GetPayloadExamples()...)
	}
	validators := make([]stub.StubsValidator, 0)
	for _, service := range services {
		validators = append(validators, service.GetStubsValidator())
	}
	go StartRESTServer(restPort, CreateRESTControllers(stubsExamples, stubsStore, supportedFullMethodNames, validators))
	StarGRPCServer(grpcPort, services)
}

func setupLogrus() {
	log.SetFormatter(&log.TextFormatter{
		FullTimestamp: true,
	})
	log.SetLevel(log.DebugLevel)
}
