package restcontrollers

import (
	"github.com/thebaasco/protoc-gen-mock/stub"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestExamplesController_GetPath(t *testing.T) {
	ctrl := ExamplesController{}

	assert.Equal(t, "/examples", ctrl.GetPath())
}

func TestExamplesController_GetHandlers(t *testing.T) {
	ctrl := ExamplesController{}

	assert.Equal(t, 1, len(ctrl.GetHandlers()))
	assert.Equal(t, http.MethodGet, strings.Join(ctrl.GetHandlers()[0].Methods, ""))
	assert.Equal(t, "", ctrl.GetHandlers()[0].Path)
	assert.Equal(t, "GetExamples", ctrl.GetHandlers()[0].Name)
}

func TestExamplesController_getExamplesHandler(t *testing.T) {
	ctrl := ExamplesController{
		StubExamples: []stub.Stub{
			{
				FullMethod: "method1",
				Request: &stub.StubRequest{
					Match:    "exact",
					Content:  "request1",
					Metadata: map[string][]string{"key1": []string{"value1"}, "key2": []string{"2"}},
				},
				Response: &stub.StubResponse{
					Type:    "sccess",
					Content: "response1",
					Error: &stub.ErrorResponse{
						Code:    0,
						Message: "erro1",
						Details: nil,
					},
				},
			},
		},
	}
	response := httptest.NewRecorder()
	ctrl.GetHandlers()[0].Handler(response, nil)
	expectedBody := "[{\"fullMethod\":\"method1\",\"request\":{\"match\":\"exact\",\"content\":\"request1\",\"metadata\":{\"key1\":\"value1\",\"key2\":2}},\"response\":{\"type\":\"sccess\",\"content\":\"response1\",\"error\":\"error1\"}}]"
	assert.Equal(t, expectedBody, response.Body.String())
	assert.Equal(t, 200, response.Code)
	assert.Equal(t, "application/json", strings.Join(response.Header().Values("Content-Type"), ""))
}
