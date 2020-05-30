package grpchandler

import (
	"context"
	"fmt"
	"github.com/carvalhorr/protoc-gen-mock/stub"
	"github.com/stretchr/stew/slice"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type MockService interface {
	Register(s *grpc.Server)
	GetSupportedMethods() []string
	GetPayloadExamples() []stub.Stub
	GetRequestInstance(methodName string) interface{}
	GetResponseInstance(methodName string) interface{}
	ForwardRequest(conn grpc.ClientConnInterface, ctx context.Context, methodName string, req interface{}) (interface{}, error)
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

func (c compositeMockService) ForwardRequest(conn grpc.ClientConnInterface, ctx context.Context, methodName string, req interface{}) (interface{}, error) {
	for _, service := range c.mockServices {
		if slice.Contains(service.GetSupportedMethods(), methodName) {
			return service.ForwardRequest(conn, ctx, methodName, req)
		}
	}
	return nil, status.Error(codes.NotFound, fmt.Sprintf("Method %s is not supported.", methodName))
}
