package grpchandler

import (
	"context"
	"fmt"
	"github.com/carvalhorr/protoc-gen-mock/stub"
	githubproto "github.com/golang/protobuf/proto"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

type MockService interface {
	Register(s *grpc.Server)
	GetSupportedMethods() []string
	GetPayloadExamples() []stub.Stub
	GetRequestInstance(methodName string) interface{}
	GetResponseInstance(methodName string) interface{}
	GetStubsValidator() stub.StubsValidator
}

// MockInterceptor intercepts the gRPC calls for the registered services return canned responses previously loaded through the REST API.
var MockHandler = func(ctx context.Context, stubsMatcher stub.StubsMatcher, fullMethod string, req interface{}, resp interface{}) (_ interface{}, err error) {
	paramsJson, err := getRequestInJSON(req)
	if err != nil {
		logError(fullMethod, paramsJson, err)
		return nil, err
	}

	return getResponse(stubsMatcher, ctx, fullMethod, paramsJson, resp)
}

func getResponse(stubMatcher stub.StubsMatcher, ctx context.Context, fullMethod, requestJson string, resp interface{}) (interface{}, error) {
	stub := stubMatcher.Match(ctx, fullMethod, requestJson)
	if stub != nil {
		if stub.Response.Type == "error" {
			return createErrorResponse(stubMatcher, stub.Response.Error)
		}
		resp, transformErr := jsonToResponse(stub.Response.Content.String(), resp)
		if transformErr != nil {
			log.WithFields(log.Fields{"Error": transformErr.Error()}).
				Errorf("Error handling request %s --> %s", fullMethod, requestJson)

			return nil, fmt.Errorf("could not unmarshal response")
		}
		log.WithFields(log.Fields{"response": resp}).
			Infof("Found MOCK response for %s --> %s", fullMethod, requestJson)
		return resp, nil
	}

	log.Infof("NO mock response found for %s --> %s", fullMethod, requestJson)
	return nil, fmt.Errorf("no response found")
}

func createErrorResponse(stubMatcher stub.StubsMatcher, stubError *stub.ErrorResponse) (interface{}, error) {
	st := status.New(codes.Code(uint32(stubError.Code)), stubError.Message)
	if stubError.Details != nil {
		log.Debugf("Creating instance of base error from spec /%s/%s", stubError.Details.Spec.Import, stubError.Details.Spec.Type)
		baseErrorType, err := stubMatcher.GetErrorEngine().GetNewInstance(stubError.Details.Spec)
		if err != nil {
			log.Errorf("Expansion of error response failed: %s", err.Error())
			return nil, status.New(codes.Internal, "Expansion of error response failed").Err()
		}
		detailsMessages := make([]githubproto.Message, 0)
		for _, errDetailValue := range stubError.Details.Values {
			errorType := baseErrorType
			if errDetailValue.SpecOverride != nil && errDetailValue.SpecOverride.Import != "" {
				log.Debugf("Creating instance of error from spec /%s/%s", errDetailValue.SpecOverride.Import, errDetailValue.SpecOverride.Type)
				errorType, err = stubMatcher.GetErrorEngine().GetNewInstance(errDetailValue.SpecOverride)
				if err != nil {
					log.Errorf("Expansion of error response failed: %s", err.Error())
					return nil, status.New(codes.Internal, "Expansion of error response failed").Err()
				}
			}
			log.Debugf("Loading JSON into error: %s", errDetailValue.Value.String())
			detailMessage, err := jsonToResponse(errDetailValue.Value.String(), errorType)
			if err != nil {
				log.Errorf("Expansion of error response failed: %s", err.Error())
				return nil, status.New(codes.Internal, "Expansion of error response failed").Err()
			}
			detailsMessages = append(detailsMessages, detailMessage.(githubproto.Message))
		}
		st, err = st.WithDetails(detailsMessages...)
		if err != nil {
			log.Errorf("Error creating error details: %s", err.Error())
			return nil, fmt.Errorf("")
		}
	}
	return nil, st.Err()
}

func logError(fullMethod, paramsJSON string, err error) {
	log.WithFields(log.Fields{"Error": err.Error()}).
		Errorf("Error handling request %s --> %s", fullMethod, paramsJSON)
}

func getRequestInJSON(req interface{}) (requestJSON string, err error) {
	message := req.(proto.Message)
	bytes, err := protojson.Marshal(message)
	if err != nil {
		return "", fmt.Errorf("could not marshal the request to JSON: %w", err)
	}
	requestJSON = string(bytes)
	return
}

func jsonToResponse(jsonString string, returnTypeInstance interface{}) (interface{}, error) {
	protoMessage := returnTypeInstance.(proto.Message)
	err := protojson.Unmarshal([]byte(jsonString), protoMessage)
	if err != nil {
		return nil, err
	}
	return returnTypeInstance, nil
}
