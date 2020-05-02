package stub

import (
	"context"
	"fmt"
	"github.com/gogo/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"sort"
	"strings"
)

// Search and match stubs in the StubsStore
type StubsMatcher interface {
	GetResponse(ctx context.Context, fullMethod, requestJson string, resp interface{}) (interface{}, error)
}

// Creates new stubs matcher
func NewStubsMatcher(store StubsStore, errorEngine CustomErrorEngine) StubsMatcher {
	return &stubsMatcher{
		StubsStore:        store,
		CustomErrorEngine: errorEngine,
	}
}

type stubsMatcher struct {
	StubsStore        StubsStore
	CustomErrorEngine CustomErrorEngine
}

// Returns the Stub in the StubsStore that matches the method and requestJSON provided OR nil if no stub is found
func (m *stubsMatcher) match(ctx context.Context, fullMethod, requestJson string) *Stub {
	stubsForMethod := m.StubsStore.GetStubsMapForMethod(fullMethod)
	if stubsForMethod == nil {
		return nil
	}
	for _, stub := range stubsForMethod {
		switch stub.Request.Match {
		case "exact":
			if string(stub.Request.Content) == requestJson && matchMetadata(ctx, stub) {
				return stub
			}
		case "partial":
			// not implemented
		}
	}
	return nil
}

func (m *stubsMatcher) GetResponse(ctx context.Context, fullMethod, requestJson string, resp interface{}) (interface{}, error) {
	stub := m.match(ctx, fullMethod, requestJson)
	if stub != nil {
		if stub.Response.Type == "error" {
			return m.createErrorResponse(stub.Response.Error)
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

func (m *stubsMatcher) createErrorResponse(stubError *ErrorResponse) (interface{}, error) {
	st := status.New(codes.Code(uint32(stubError.Code)), stubError.Message)
	if stubError.Details != nil {
		log.Debugf("Creating instance of base error error from spec /%s/%s", stubError.Details.Spec.Import, stubError.Details.Spec.Type)
		baseErrorType, err := m.CustomErrorEngine.GetNewInstance(stubError.Details.Spec)
		if err != nil {
			log.Errorf("Expansion of error response failed: %s", err.Error())
			return nil, status.New(codes.Internal, "Expansion of error response failed").Err()
		}
		detailsMessages := make([]proto.Message, 0)
		for _, errDetailValue := range stubError.Details.Values {
			log.Debug(errDetailValue.Value)
			errorType := baseErrorType
			if errDetailValue.SpecOverride != nil && errDetailValue.SpecOverride.Import != "" {
				log.Debugf("Creating instance of error from spec /%s/%s", errDetailValue.SpecOverride.Import, errDetailValue.SpecOverride.Type)
				errorType, err = m.CustomErrorEngine.GetNewInstance(errDetailValue.SpecOverride)
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
			detailsMessages = append(detailsMessages, detailMessage.(proto.Message))
		}
		st, err = st.WithDetails(detailsMessages...)
		if err != nil {
			log.Errorf("Error creating error details: %s", err.Error())
			return nil, fmt.Errorf("")
		}
	}
	return nil, st.Err()
}

func matchMetadata(ctx context.Context, stub *Stub) bool {
	if len(stub.Request.Metadata) == 0 {
		return true
	}
	// read metadata from context
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return false
	}
	stubMetadata := getStubMetadata(stub)
	// compare
	for key, values := range stubMetadata {
		contextMetadata := md.Get(key)
		sort.Strings(contextMetadata)
		sort.Strings(values)
		if strings.Join(values, ",") != strings.Join(contextMetadata, ",") {
			return false
		}
	}
	return true
}

func getStubMetadata(stub *Stub) (stubMetadata map[string][]string) {
	stubMetadata = make(map[string][]string, 0)
	for key, value := range stub.Request.Metadata {
		parts := strings.Split(value, ",")
		for _, part := range parts {
			stubMetadata[key] = append(stubMetadata[key], strings.TrimSpace(part))
		}
	}
	return
}

func jsonToResponse(jsonString string, returnTypeInstance interface{}) (interface{}, error) {
	protoMessage := returnTypeInstance.(proto.Message)
	err := jsonpb.Unmarshal(strings.NewReader(jsonString), protoMessage)

	if err != nil {
		return nil, err
	}
	return returnTypeInstance, nil
}
