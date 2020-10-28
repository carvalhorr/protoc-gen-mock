package restcontrollers

import (
	"github.com/carvalhorr/protoc-gen-mock/stub"
	log "github.com/sirupsen/logrus"
	"net/http"
)

type RecordingsController struct {
	RecordingsStore stub.RecordingsStore
}

func (c RecordingsController) GetHandlers() []RESTHandler {
	return []RESTHandler{
		{
			Name:    "GetRecordings",
			Path:    "",
			Methods: []string{http.MethodGet},
			Handler: c.getRecordingsHandler,
		},
	}
}

func (c RecordingsController) GetPath() string {
	return "/recordings"
}

func (c RecordingsController) getRecordingsHandler(writer http.ResponseWriter, request *http.Request) {
	log.Info("REST: received call to get recordings")

	recordings := c.RecordingsStore.GetAllStubs()
	writeErr := writeResponse(writer, recordings)
	if writeErr != nil {
		writeErrorResponse(writer, http.StatusInternalServerError, writeErr.Error())
	}
}
