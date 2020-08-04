package restcontrollers

import (
	"github.com/carvalhorr/protoc-gen-mock/stub"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestStubsController_addStubHandler(t *testing.T) {
	stubsStore := stub.NewInMemoryStubsStore(false)
	ctrl := StubsController{
		StubsStore: stubsStore,
	}
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/stubs", strings.NewReader(`{
    "fullMethod": "method1",
    "request": {
    	"match": "exact",
    	"content": {
    		"name":"Rodrigo"
    	},
    	"metadata": {
    		"key1": "value1",
    		"key2": "value2",
    		"key3": 1
    	}
    },
    "response": {
    	"type": "success",
    	"content": {
    		"name":"Rodrigo de Carvalho"
    	},
    	"error": "an unfortunate error"
    }
}`))
	findHandler(ctrl.GetHandlers(), "AddStub").Handler(response, request)
	assert.Equal(t, 1, len(stubsStore.GetAllStubs()))
	assert.Equal(t, "OK", response.Body.String())
	assert.Equal(t, 200, response.Code)
}

func TestStubsController_addStubHandler_MethodNotSupportedError(t *testing.T) {
	stubsStore := stub.NewInMemoryStubsStore(false)
	ctrl := StubsController{
		StubsStore: stubsStore,
	}
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/stubs", strings.NewReader(`{
    "fullMethod": "NOT_SUPPORTED_METHOD"
}`))
	findHandler(ctrl.GetHandlers(), "AddStub").Handler(response, request)
	assert.Equal(t, 0, len(stubsStore.GetAllStubs()))
	assert.Equal(t, "Method NOT_SUPPORTED_METHOD is not supported", response.Body.String())
	assert.Equal(t, 400, response.Code)
}
