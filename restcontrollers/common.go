package restcontrollers

import (
	"encoding/json"
	"fmt"
	log "github.com/sirupsen/logrus"
	"net/http"
)

type RESTController interface {
	GetHandlers() []RESTHandler
	GetPath() string
}

type RESTHandler struct {
	Name    string
	Path    string
	Methods []string
	Handler func(writer http.ResponseWriter, request *http.Request)
}

func getQueryParam(request *http.Request, paramName string) string {
	values, ok := request.URL.Query()[paramName]
	if !ok || len(values) == 0 {
		return emptyString
	}
	return values[0]
}

func writeResponse(writer http.ResponseWriter, respponse interface{}) error {
	return writeResponseWithCode(writer, respponse, http.StatusOK)
	return nil
}

func writeResponseWithCode(writer http.ResponseWriter, respponse interface{}, code int) error {
	responseJSONBytes, err := json.Marshal(respponse)
	if err != nil {
		log.Errorf("Unexpected error while writing response in JSON. Error %s", err.Error())
		return fmt.Errorf("error writing response to JSON")
	}

	writer.Header().Add(contentType, contentTypeApplicationJson)
	writer.WriteHeader(code)
	_, writeErr := writer.Write(responseJSONBytes)
	if writeErr != nil {
		return writeErr
	}

	return nil
}

func writeSuccessResponse(writer http.ResponseWriter) {
	writer.WriteHeader(http.StatusOK)
	_, writeErr := writer.Write([]byte(http.StatusText(http.StatusOK)))
	if writeErr != nil {
		log.Errorf("Error writing http response: Error %s", writeErr.Error())
	}
}

func writeErrorResponse(writer http.ResponseWriter, code int, message string) {
	log.Warn(message)
	writer.WriteHeader(code)
	_, writeErr := writer.Write([]byte(message))
	if writeErr != nil {
		log.Errorf("Error writing http response: Error %s", writeErr.Error())
	}
}
