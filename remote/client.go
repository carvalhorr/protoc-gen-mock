package remote

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	httputils "github.com/carvalhorr/goutils/http"
	log "github.com/sirupsen/logrus"
	"github.com/thebaasco/protoc-gen-mock/stub"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"net/http"
)

type MockServerClient interface {
	AddStub(
		fullMethod string,
		ctx context.Context,
		req proto.Message,
		resp proto.Message,
		error *status.Status,
	) error

	DeleteAllStubs() error
}

func New(
	host string,
	port int,
) MockServerClient {
	return &client{
		HttpClient: &http.Client{},
		host:       host,
		port:       port,
	}
}

type client struct {
	// TODO: Should HttpClient be exported?
	// The only need would be in case we want allow mocking it outside of this package.
	// We are sorted for unit testing since we the tests are in the same package (white-box testing)
	// Is there a need to mock http.Client for integration tests?
	HttpClient httputils.Client
	host       string
	port       int
}

func (c *client) AddStub(
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
	if err != nil {
		return err
	}
	writer := bytes.NewBuffer(b)
	r, e := c.HttpClient.Post(fmt.Sprintf("http://%s:%d/stubs", c.host, c.port), "application/json", writer)
	if e != nil {
		return e
	}
	defer r.Body.Close()
	return nil
}

func (c *client) DeleteAllStubs() error {
	deleteRequest, err := http.NewRequest(http.MethodDelete, fmt.Sprintf("http://%s:%d/stubs", c.host, c.port), bytes.NewReader([]byte{}))
	if err != nil {
		return err
	}
	resp, err := c.HttpClient.Do(deleteRequest)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode == 200 {
		return nil
	}
	return fmt.Errorf("error: status %s", resp.Status)
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
