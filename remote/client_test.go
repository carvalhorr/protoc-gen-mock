//+build unit all

package remote

import (
	"fmt"
	httputils "github.com/carvalhorr/goutils/http"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"
)

func TestDeleteAllStubs_Success(t *testing.T) {
	mockHttpClient := new(httputils.MockClient)
	mockHttpClient.On("Do",
		mock.AnythingOfType("*http.Request")).Return(&http.Response{
		Status:     "OK",
		StatusCode: 200,
		Body:       ioutil.NopCloser(strings.NewReader("hello world")),
	}, nil)
	client := &client{
		HttpClient: mockHttpClient,
	}
	err := client.DeleteAllStubs()
	assert.Nil(t, err)
}

func TestDeleteAllStubs_HttpError(t *testing.T) {
	mockHttpClient := new(httputils.MockClient)
	mockHttpClient.On("Do", mock.AnythingOfType("*http.Request")).Return(nil, fmt.Errorf("error 123"))
	client := &client{
		HttpClient: mockHttpClient,
	}
	err := client.DeleteAllStubs()
	assert.EqualError(t, err, "error 123")
}

func TestDeleteAllStubs_StatusNot200(t *testing.T) {
	mockHttpClient := new(httputils.MockClient)
	mockHttpClient.On("Do",
		mock.AnythingOfType("*http.Request")).Return(&http.Response{
		Status:     "400 - mocked status",
		StatusCode: 400,
		Body:       ioutil.NopCloser(strings.NewReader("")),
	}, nil)
	client := &client{
		HttpClient: mockHttpClient,
	}
	err := client.DeleteAllStubs()
	assert.EqualError(t, err, "error: status 400 - mocked status")
}
