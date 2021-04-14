package restcontrollers

import (
	"github.com/thebaasco/protoc-gen-mock/stub"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestStubsController_deleteStubHandler(t *testing.T) {
	stubsStore := stub.NewInMemoryStubsStore()
	stubsStore.Add(&stub.Stub{
		FullMethod: "method1",
		Request: &stub.StubRequest{
			Match:   "exact",
			Content: "{\"name\":\"Rodrigo\"}",
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
	payload := `{
    "fullMethod": "method1",
    "request": {
    	"match": "exact",
    	"content": {
    		"name":"Rodrigo"
    	}
    },
    "response": {
    	"type": "success",
    	"content": {
    		"name":"Rodrigo de Carvalho UPDATED"
    	},
    	"error": "an unfortunate error"
    }
}`
	request := httptest.NewRequest(http.MethodDelete, "/stubs", strings.NewReader(payload))
	findHandler(ctrl.GetHandlers(), "DeleteStub").Handler(response, request)
	assert.Equal(t, 0, len(stubsStore.GetAllStubs()))
	assert.Equal(t, "OK", response.Body.String())
	assert.Equal(t, 200, response.Code)
}
