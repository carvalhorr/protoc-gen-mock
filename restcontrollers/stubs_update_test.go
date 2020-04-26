package restcontrollers

import (
	"github.com/carvalhorr/protoc-gen-mock/stub"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestStubsController_updateStubHandler(t *testing.T) {
	stubsStore := stub.NewInMemoryStubsStore()
	stubsStore.Add(&stub.Stub{
		FullMethod: "method1",
		Request: stub.StubRequest{
			Match:   "exact",
			Content: "{\"name\":\"Rodrigo\"}",
		},
		Response: stub.StubResponse{
			Type:    "sccess",
			Content: "response1",
			Error:   "error1",
		},
	})
	ctrl := StubsController{
		StubsStore:       stubsStore,
		SupportedMethods: []string{"method1"},
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
	request := httptest.NewRequest(http.MethodPut, "/stubs", strings.NewReader(payload))
	findHandler(ctrl.GetHandlers(), "UpdateStub").Handler(response, request)
	assert.Equal(t, 1, len(stubsStore.GetAllStubs()))
	assert.Equal(t, "OK", response.Body.String())
	assert.Equal(t, 200, response.Code)
	assert.Equal(t, "{\"name\":\"Rodrigo de Carvalho UPDATED\"}", string(stubsStore.GetAllStubs()[0].Response.Content))
}
