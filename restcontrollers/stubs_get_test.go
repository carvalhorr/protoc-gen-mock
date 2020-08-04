package restcontrollers

import (
	"github.com/carvalhorr/protoc-gen-mock/stub"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestStubsController_getStubsHandler(t *testing.T) {
	stubsStore := stub.NewInMemoryStubsStore(false)
	stubsStore.Add(&stub.Stub{
		FullMethod: "method1",
		Request: &stub.StubRequest{
			Match:   "exact",
			Content: "request1",
			Metadata: map[string][]string{
				"key1": []string{"value1"},
				"key2": []string{"2"},
			},
		},
		Response: &stub.StubResponse{
			Type:    "sccess",
			Content: "response1",
			Error: &stub.ErrorResponse{
				Code:    0,
				Message: "",
				Details: nil,
			},
		},
	})
	ctrl := StubsController{
		StubsStore: stubsStore,
	}
	response := httptest.NewRecorder()
	request := &http.Request{
		Method: http.MethodGet,
		URL:    &url.URL{},
	}
	findHandler(ctrl.GetHandlers(), "GetStubs").Handler(response, request)
	expectedBody := "[{\"fullMethod\":\"method1\",\"request\":{\"match\":\"exact\",\"content\":\"request1\",\"metadata\":{\"key1\":\"value1\",\"key2\":2}},\"response\":{\"type\":\"sccess\",\"content\":\"response1\",\"error\":\"error1\"}}]"
	assert.Equal(t, expectedBody, response.Body.String())
	assert.Equal(t, 200, response.Code)
	assert.Equal(t, "application/json", strings.Join(response.Header().Values("Content-Type"), ""))
}

func TestStubsController_getStubsHandler_MethodNotSupportedError(t *testing.T) {
	ctrl := StubsController{}
	response := httptest.NewRecorder()
	request := &http.Request{
		Method: http.MethodGet,
		URL: &url.URL{
			RawQuery: "method=test123",
		},
	}
	findHandler(ctrl.GetHandlers(), "GetStubs").Handler(response, request)
	expectedBody := "Unsupported method: test123"
	assert.Equal(t, expectedBody, response.Body.String())
	assert.Equal(t, 400, response.Code)
}
