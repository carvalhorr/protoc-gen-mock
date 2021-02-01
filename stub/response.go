package stub

import (
	"fmt"
	"github.com/golang/protobuf/jsonpb"
	githubproto "github.com/golang/protobuf/proto"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	protojson22 "google.golang.org/protobuf/encoding/protojson"
	proto22 "google.golang.org/protobuf/proto"
	protoreflect22 "google.golang.org/protobuf/reflect/protoreflect"
	"strings"
)

var errorEngine CustomErrorEngine

func SetErrorEngine(engine CustomErrorEngine) {
	errorEngine = engine
}

func GetResponse(stub *Stub, requestJson string, resp interface{}) (interface{}, error) {
	if stub == nil {
		return nil, nil
	}
	if stub.Response.Type == "error" {
		return createErrorResponse(errorEngine, stub.Response.Error)
	}
	resp, transformErr := jsonToResponse(stub.Response.Content.String(), resp)
	if transformErr != nil {
		log.WithFields(log.Fields{"Error": transformErr.Error()}).
			Errorf("Error handling request %s --> %s", stub.FullMethod, requestJson)

		return nil, fmt.Errorf("could not unmarshal response")
	}
	log.WithFields(log.Fields{"response": resp}).
		Infof("Found MOCK response for %s --> %s", stub.FullMethod, requestJson)
	return resp, nil
}

func createErrorResponse(errorEngine CustomErrorEngine, stubError *ErrorResponse) (interface{}, error) {
	st := status.New(codes.Code(stubError.Code), stubError.Message)
	if stubError.Details != nil {
		log.Debugf("Creating instance of base error from spec /%s/%s", stubError.Details.Spec.Import, stubError.Details.Spec.Type)
		baseErrorType, err := errorEngine.GetNewInstance(stubError.Details.Spec)
		if err != nil {
			log.Errorf("Expansion of error response failed: %s", err.Error())
			return nil, status.New(codes.Internal, "Expansion of error response failed").Err()
		}
		detailsMessages := make([]githubproto.Message, 0)
		for _, errDetailValue := range stubError.Details.Values {
			errorType := baseErrorType
			if errDetailValue.SpecOverride != nil && errDetailValue.SpecOverride.Import != "" {
				log.Debugf("Creating instance of error from spec /%s/%s", errDetailValue.SpecOverride.Import, errDetailValue.SpecOverride.Type)
				errorType, err = errorEngine.GetNewInstance(errDetailValue.SpecOverride)
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

func jsonToResponse(jsonString string, returnTypeInstance interface{}) (interface{}, error) {
	var err error
	if isCompatibleWithProtobug22(returnTypeInstance) {
		protoMessage := returnTypeInstance.(proto22.Message)
		err = protojson22.Unmarshal([]byte(jsonString), protoMessage)
	} else {
		protoMessage := returnTypeInstance.(githubproto.Message)
		err = jsonpb.Unmarshal(strings.NewReader(jsonString), protoMessage)
	}
	if err != nil {
		return nil, err
	}
	return returnTypeInstance, nil
}

func isCompatibleWithProtobug22(instance interface{}) bool {
	switch instance.(type) {
	case protoreflect22.ProtoMessage:
		return true
	}
	return false
}
