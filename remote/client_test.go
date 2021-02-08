//+build unit all

package remote

import (
	"context"
	"fmt"
	httputils "github.com/carvalhorr/goutils/http"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"
)

func TestClient_DeleteAllStubs_Success(t *testing.T) {
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

func TestClient_DeleteAllStubs_HttpError(t *testing.T) {
	mockHttpClient := new(httputils.MockClient)
	mockHttpClient.On("Do", mock.AnythingOfType("*http.Request")).Return(nil, fmt.Errorf("error 123"))
	client := &client{
		HttpClient: mockHttpClient,
	}
	err := client.DeleteAllStubs()
	assert.EqualError(t, err, "error 123")
}

func TestClient_DeleteAllStubs_StatusNot200(t *testing.T) {
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

func TestClient_AddStub_Success(t *testing.T) {
	mockHttpClient := new(httputils.MockClient)
	mockHttpClient.On("Post",
		mock.Anything, mock.Anything, mock.Anything).Return(&http.Response{
		Status:     "OK",
		StatusCode: 201,
		Body:       ioutil.NopCloser(strings.NewReader("hello world")),
	}, nil)
	client := &client{
		HttpClient: mockHttpClient,
	}
	err := client.AddStub("", context.Background(), &Request{}, &Response{}, nil)
	assert.Nil(t, err)
}

func TestClient_AddStub_NilRequest_Success(t *testing.T) {
	mockHttpClient := new(httputils.MockClient)
	mockHttpClient.On("Post",
		mock.Anything, mock.Anything, mock.Anything).Return(&http.Response{
		Status:     "OK",
		StatusCode: 201,
		Body:       ioutil.NopCloser(strings.NewReader("hello world")),
	}, nil)
	client := &client{
		HttpClient: mockHttpClient,
	}
	err := client.AddStub("", context.Background(), nil, &Response{}, nil)
	assert.Nil(t, err)
}

func TestClient_AddStub_NilResponse_Success(t *testing.T) {
	mockHttpClient := new(httputils.MockClient)
	mockHttpClient.On("Post",
		mock.Anything, mock.Anything, mock.Anything).Return(&http.Response{
		Status:     "OK",
		StatusCode: 201,
		Body:       ioutil.NopCloser(strings.NewReader("hello world")),
	}, nil)
	client := &client{
		HttpClient: mockHttpClient,
	}
	err := client.AddStub("", context.Background(), &Request{}, nil, nil)
	assert.Nil(t, err)
}

func TestClient_AddStub_ErrorResponse_Success(t *testing.T) {
	mockHttpClient := new(httputils.MockClient)
	mockHttpClient.On("Post",
		mock.Anything, mock.Anything, mock.Anything).Return(&http.Response{
		Status:     "OK",
		StatusCode: 201,
		Body:       ioutil.NopCloser(strings.NewReader("hello world")),
	}, nil)
	client := &client{
		HttpClient: mockHttpClient,
	}
	err := client.AddStub("", context.Background(), &Request{}, nil, status.New(codes.AlreadyExists, "error"))
	assert.Nil(t, err)
}

func TestClient_AddStub_HttpError(t *testing.T) {
	mockHttpClient := new(httputils.MockClient)
	mockHttpClient.On("Post",
		mock.Anything, mock.Anything, mock.Anything).Return(nil, fmt.Errorf("http error"))
	client := &client{
		HttpClient: mockHttpClient,
	}
	err := client.AddStub("", context.Background(), &Request{}, nil, status.New(codes.AlreadyExists, "error"))
	assert.EqualError(t, err, "http error")
}
