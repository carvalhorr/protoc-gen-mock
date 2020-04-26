package restcontrollers

import (
	"github.com/stretchr/testify/assert"
	"net/http"
	"strings"
	"testing"
)

func TestStubsController_GetPath_GetPath(t *testing.T) {
	ctrl := StubsController{}

	assert.Equal(t, "/stubs", ctrl.GetPath())
}

func TestStubsController_GetHandlers(t *testing.T) {
	ctrl := StubsController{}

	assert.Equal(t, 4, len(ctrl.GetHandlers()))
	validateHandler(t, findHandler(ctrl.GetHandlers(), "GetStubs"), http.MethodGet)
	validateHandler(t, findHandler(ctrl.GetHandlers(), "AddStub"), http.MethodPost)
	validateHandler(t, findHandler(ctrl.GetHandlers(), "UpdateStub"), http.MethodPut)
	validateHandler(t, findHandler(ctrl.GetHandlers(), "DeleteStub"), http.MethodDelete)
}

func validateHandler(t *testing.T, handler *RESTHandler, method string) {
	t.Run(handler.Name, func(t *testing.T) {
		assert.Equal(t, method, strings.Join(handler.Methods, ""))
		assert.Equal(t, "", handler.Path)
	})
}

func findHandler(handlers []RESTHandler, name string) *RESTHandler {
	for _, handler := range handlers {
		if handler.Name == name {
			return &handler
		}
	}
	return nil
}
