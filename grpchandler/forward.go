package grpchandler

import (
	"context"
	"encoding/json"
	"github.com/carvalhorr/protoc-gen-mock/stub"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var supportedMockService MockService
var stubStore stub.StubsStore

func SetSupportedMockService(service MockService) {
	supportedMockService = service
}

func SetStubStore(store stub.StubsStore) {
	stubStore = store
}

func forwardAndRecord(s *stub.Stub, ctx context.Context, fullMethod string, req, resp interface{}) (_ interface{}, err error) {
	if s.Type != "forward" {
		return nil, status.Error(codes.Internal, "Attempt to cal forward for a stub that is not of type 'forward'")
	}
	log.Infof("Forwarding to %s (%s -> %s)", s.Forward.ServerAddress, fullMethod, s.Request.String())
	conn := createConnection(s.Forward)
	defer conn.Close()

	resp, err = supportedMockService.ForwardRequest(conn, ctx, fullMethod, req)
	log.Infof("Got forward response %s and error %s", toJson(resp), errToString(err))
	if s.Forward.Record {
		log.Infof("Recording is active for stub %s -> %s", fullMethod, toJson(s.Request))
		recordRequestAndResponse(fullMethod, req, resp, err)
	}
	return resp, err
}

func createConnection(forward *stub.StubForward) *grpc.ClientConn {

	options := make([]grpc.DialOption, 0)
	options = append(options, grpc.WithInsecure()) // TODO implement security later
	conn, err := grpc.Dial(forward.ServerAddress, options...)
	if err != nil {
		log.Errorf("Failed to create connection to %s", forward.ServerAddress)
	}
	return conn
}

func recordRequestAndResponse(fullMethod string, req, resp interface{}, err error) {
	s := &stub.Stub{
		FullMethod: fullMethod,
		Type:       "mock",
		Request: &stub.StubRequest{
			Match:    "exact",
			Content:  toJson(req),
			Metadata: nil, // TODO Record metadata
		},
		Response: &stub.StubResponse{
			Type:    getResponseType(resp, err),
			Content: toJson(resp),
			Error:   mapError(err),
		},
		Forward: nil,
	}
	addErr := stubStore.Add(s)
	if addErr != nil {
		log.Errorf("Failed to record forwarding result. Error: %s", addErr)
	}
}

func getResponseType(resp interface{}, err error) string {
	if err != nil {
		return "error"
	}
	return "success"
}

func toJson(instance interface{}) stub.JsonString {
	bytes, err := json.Marshal(instance)
	if err != nil {
		log.Errorf("Failed to marshal to JSON.")
	}
	return stub.JsonString(bytes)
}

func errToString(err error) string {
	if err == nil {
		return "<nil>"
	}
	return err.Error()
}

func mapError(err error) *stub.ErrorResponse {
	// TODO Map error response when forwarding for recording
	return nil
}
