package grpchandler

import (
	"github.com/thebaasco/protoc-gen-mock/stub"
	"github.com/stretchr/testify/mock"
)

type MockStubsMatcher struct {
	mock.Mock
}

func (m MockStubsMatcher) Match(method string, reqJSON string) *stub.Stub {
	args := m.Called(method, reqJSON)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*stub.Stub)
}

type Response struct {
	Name string `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
}

func (*Response) ProtoMessage() {}

func (x *Response) Reset() {}

func (*Response) String() string { return "" }

type Request struct {
	Name string `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
}

/*
func TestMockHandler_Success_FoundResponse(t *testing.T) {
	method := "grpc_method_1"

	// Setup mock dependencies
	mockStubsMatcher := new(MockStubsMatcher)
	mockStubsMatcher.On("Match", mock.Anything, mock.Anything).
		Return(&stub.Stub{
			FullMethod: method,
			Response: &stub.StubResponse{
				Content: "{\"name\":\"Rodrigo de Carvalho\"}",
			},
		})

	foundStub, _ := MockHandler(context.Background(), mockStubsMatcher, method, new(Request), new(Response))
	assert.Equal(t, "Rodrigo de Carvalho", foundStub.(*Response).Name)
}
*/

/*
func TestMockHandler_Success_FoundError(t *testing.T) {
	method := "grpc_method_1"

	// Setup mock dependencies
	mockStubsMatcher := new(MockStubsMatcher)
	mockStubsMatcher.On("Match", mock.Anything, mock.Anything).
		Return(&stub.Stub{
			FullMethod: method,
			Response: &stub.StubResponse{
				Type:    "error",
				Content: "",
				Error:   "return an error",
			},
		})

	_, err := MockHandler(context.Background(), mockStubsMatcher, method, new(Request), new(Response))
	assert.EqualError(t, err, "return an error")
}
*/

/*
func TestMockHandler_Success_NoStubFound(t *testing.T) {
	method := "grpc_method_1"

	// Setup mock dependencies
	mockStubsMatcher := new(MockStubsMatcher)
	mockStubsMatcher.On("Match", mock.Anything, mock.Anything).
		Return(nil)

	_, err := MockHandler(context.Background(), mockStubsMatcher, method, new(Request), new(Response))
	assert.EqualError(t, err, "no response found")
}
*/

/*
func TestMockHandler_ResponseJsonWrongFormat(t *testing.T) {
	method := "grpc_method_1"

	// Setup mock dependencies
	mockStubsMatcher := new(MockStubsMatcher)
	mockStubsMatcher.On("Match", mock.Anything, mock.Anything).
		Return(&stub.Stub{
			FullMethod: method,
			Response: stub.StubResponse{
				Content: "wrong_json",
			},
		})

	_, err := MockHandler(context.Background(), mockStubsMatcher, method, new(Request), new(Response))
	assert.EqualError(t, err, "could not unmarshal response")
}
*/

/*
func TestMockHandler_ErrorMarshalingRequest(t *testing.T) {
	method := "grpc_method_1"

	// Setup mock dependencies
	mockStubsMatcher := new(MockStubsMatcher)

	_, err := MockHandler(context.Background(), mockStubsMatcher, method, math.Inf(1), new(Response))
	assert.EqualError(t, err, "could not marshal the request to JSON: json: unsupported value: +Inf")
}
*/
