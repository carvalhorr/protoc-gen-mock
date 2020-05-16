package grpchandler

import (
	"context"
	"fmt"
	"github.com/carvalhorr/protoc-gen-mock/stub"
	log "github.com/sirupsen/logrus"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

// MockInterceptor intercepts the gRPC calls for the registered services return canned responses previously loaded through the REST API.
var MockHandler = func(ctx context.Context, stubsMatcher stub.StubsMatcher, fullMethod string, req interface{}, resp interface{}) (_ interface{}, err error) {
	paramsJson, err := getRequestInJSON(req)
	if err != nil {
		logError(fullMethod, paramsJson, err)
		return nil, err
	}
	s := stubsMatcher.Match(ctx, fullMethod, paramsJson)
	if s == nil {
		log.Infof("NO mock response found for %s --> %s", fullMethod, paramsJson)
		return nil, fmt.Errorf("no response found")
	}
	return stub.GetResponse(s, paramsJson, resp)
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
