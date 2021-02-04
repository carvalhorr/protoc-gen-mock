package remote

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/carvalhorr/protoc-gen-mock/stub"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"io/ioutil"
	"net/http"
)

var httpClient = http.DefaultClient

func AddStub(
	host string,
	port int,
	fullMethod string,
	ctx context.Context,
	req proto.Message,
	resp proto.Message,
	error *status.Status,
) error {
	reqJson := toJsonString(req)
	respJson := toJsonString(resp)
	errResp := toErrorResponse(error)
	s := &stub.Stub{
		FullMethod: fullMethod,
		Type:       "mock",
		Request: &stub.StubRequest{
			Match:    "exact",
			Content:  reqJson,
			Metadata: getMetadata(ctx),
		},
		Response: &stub.StubResponse{
			Type:    getResponseType(resp, error),
			Content: respJson,
			Error:   errResp,
		},
		Forward: nil,
	}
	b, err := json.Marshal(s)
	writer := bytes.NewBuffer(b)
	r, e := httpClient.Post(fmt.Sprintf("http://%s:%d/stubs", host, port), "application/json", writer)
	if e != nil {
		return e
	}
	b, err = ioutil.ReadAll(r.Body)
	if err != nil {
		return err
	}
	return nil
}

func DeleteAllStubs(
	host string,
	port int,
) error {
	deleteRequest, err := http.NewRequest(http.MethodDelete, fmt.Sprintf("http://%s:%d/stubs", host, port), bytes.NewReader([]byte{}))
	if err != nil {
		return err
	}
	resp, err := httpClient.Do(deleteRequest)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode == 200 {
		return nil
	}
	body, err := ioutil.ReadAll(resp.Body)
	return fmt.Errorf("error: %s", body)
}

func getMetadata(ctx context.Context) map[string][]string {
	md, ok := metadata.FromOutgoingContext(ctx)
	if !ok {
		log.Errorf("Metadata error")
	}
	metas := make(map[string][]string)
	for key, value := range md {
		metas[key] = value
	}
	return metas
}

func getResponseType(resp proto.Message, error *status.Status) string {
	if error == nil {
		return "success"
	}
	return "error"
}

func toJsonString(msg proto.Message) stub.JsonString {
	if msg == nil {
		return stub.JsonString("")
	}
	b, err := protojson.Marshal(msg)
	if err != nil {
		log.Errorf("Failed to marshal to JSON.")
	}
	return stub.JsonString(b)
}

func toErrorResponse(st *status.Status) *stub.ErrorResponse {
	if st == nil {
		return nil
	}
	return &stub.ErrorResponse{
		Code:    uint32(st.Code()),
		Message: st.Message(),
		Details: nil,
	}
}
