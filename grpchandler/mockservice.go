package grpchandler

import (
	"github.com/carvalhorr/protoc-gen-mock/stub"
	"google.golang.org/grpc"
)

type MockService interface {
	Register(s *grpc.Server)
	GetSupportedMethods() []string
	GetPayloadExamples() []stub.Stub
	GetRequestInstance(methodName string) interface{}
	GetResponseInstance(methodName string) interface{}
	GetStubsValidator() stub.StubsValidator
}

func NewCompositeMockService(services []MockService) MockService {
	return compositeMockService{
		mockServices: services,
	}
}

type compositeMockService struct {
	mockServices []MockService
}

func (c compositeMockService) Register(s *grpc.Server) {
	for _, mockService := range c.mockServices {
		mockService.Register(s)
	}
}

func (c compositeMockService) GetSupportedMethods() []string {
	methods := make([]string, 0)
	for _, mockService := range c.mockServices {
		methods = append(methods, mockService.GetSupportedMethods()...)
	}
	return methods
}

func (c compositeMockService) GetPayloadExamples() []stub.Stub {
	examples := make([]stub.Stub, 0)
	for _, mockService := range c.mockServices {
		examples = append(examples, mockService.GetPayloadExamples()...)
	}
	return examples
}

func (c compositeMockService) GetRequestInstance(methodName string) interface{} {
	for _, mockService := range c.mockServices {
		instance := mockService.GetRequestInstance(methodName)
		if instance != nil {
			return instance
		}
	}
	return nil
}

func (c compositeMockService) GetResponseInstance(methodName string) interface{} {
	for _, mockService := range c.mockServices {
		instance := mockService.GetResponseInstance(methodName)
		if instance != nil {
			return instance
		}
	}
	return nil
}

func (c compositeMockService) GetStubsValidator() stub.StubsValidator {
	validators := make([]stub.StubsValidator, 0)
	for _, mockService := range c.mockServices {
		validators = append(validators, mockService.(stub.StubsValidator))
	}
	return stub.NewCompositeStubsValidator(validators)
}
