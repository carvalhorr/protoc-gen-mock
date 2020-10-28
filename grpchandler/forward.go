package grpchandler

import (
	"context"
	"github.com/carvalhorr/protoc-gen-mock/stub"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

var supportedMockService MockService
var recordingsStore stub.RecordingsStore

func SetSupportedMockService(service MockService) {
	supportedMockService = service
}

func SetRecordingsStore(store stub.RecordingsStore) {
	recordingsStore = store
}

func forwardAndRecord(s *stub.Stub, ctx context.Context, fullMethod string, req, resp interface{}) (_ interface{}, err error) {
	if s.Type != "forward" {
		return nil, status.Error(codes.Internal, "Attempt to cal forward for a stub that is not of type 'forward'")
	}
	log.Infof("Forwarding to %s (%s -> %s)", s.Forward.ServerAddress, fullMethod, s.Request.String())
	conn := createConnection(s.Forward)
	defer conn.Close()

	resp, err = supportedMockService.ForwardRequest(conn, ctx, fullMethod, req)
	log.Infof("Got forward response %s and error %s", toProtoJson(resp), errToString(err))
	if s.Forward.Record {
		log.Infof("Recording is active for stub %s -> %s", fullMethod, s.Request.String())
		recordRequestAndResponse(ctx, fullMethod, req, resp, err)
	}
	return resp, err
}

func createConnection(forward *stub.StubForward) *grpc.ClientConn {

	options := make([]grpc.DialOption, 0)
	options = append(options, grpc.WithInsecure()) // TODO deal with security
	conn, err := grpc.Dial(forward.ServerAddress, options...)
	if err != nil {
		log.Errorf("Failed to create connection to %s", forward.ServerAddress)
	}
	return conn
}

func recordRequestAndResponse(ctx context.Context, fullMethod string, req, resp interface{}, err error) {
	s := &stub.Stub{
		FullMethod: fullMethod,
		Type:       "mock",
		Request: &stub.StubRequest{
			Match:    "exact",
			Content:  toProtoJson(req),
			Metadata: getMetadata(ctx),
		},
		Response: &stub.StubResponse{
			Type:    getResponseType(resp, err),
			Content: toProtoJson(resp),
			Error:   mapError(err),
		},
		Forward: nil,
	}
	addErr := recordingsStore.Add(s)
	if addErr != nil {
		log.Errorf("Failed to record forwarding result. Error: %s", addErr)
	}
}

func getMetadata(ctx context.Context) map[string][]string {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil
	}
	return md
}

func getResponseType(resp interface{}, err error) string {
	if err != nil {
		return "error"
	}
	return "success"
}

func toProtoJson(instance interface{}) stub.JsonString {
	bytes, err := protojson.Marshal(instance.(proto.Message))
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
