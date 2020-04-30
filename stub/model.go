package stub

import (
	"bytes"
	"encoding/json"
	log "github.com/sirupsen/logrus"
	"google.golang.org/protobuf/reflect/protoreflect"
	"reflect"
)

type JsonString string

type EnumType interface {
	Descriptor() protoreflect.EnumDescriptor
}

func isEnum(t reflect.Type) bool {
	inter := reflect.TypeOf((*EnumType)(nil)).Elem()
	return t.Implements(inter)
}

type Stub struct {
	FullMethod string        `json:"fullMethod"`
	Request    *StubRequest  `json:"request"`
	Response   *StubResponse `json:"response"`
}

type StubRequest struct {
	Match    string            `json:"match"`
	Content  JsonString        `json:"content"`
	Metadata map[string]string `json:"metadata"`
}

func (s StubRequest) String() string {
	data, _ := json.Marshal(s)
	return string(data)
}

type StubResponse struct {
	Type    string         `json:"type"`
	Content JsonString     `json:"content"`
	Error   *ErrorResponse `json:"error"`
}

type ErrorResponse struct {
	Code    int32         `json:"code"`
	Message string        `json:"message"`
	Details *ErrorDetails `json:"details"`
}

type ErrorDetails struct {
	Spec   *ErrorDetailsSpec   `json:"spec"`
	Values []ErrorDetailsValue `json:"values"`
}

type ErrorDetailsValue struct {
	SpecOverride ErrorDetailsSpec `json:"specOverride"`
	Value        JsonString       `json:"value"`
}

type ErrorDetailsSpec struct {
	Import string `json:"import"`
	Type   string `json:"type"`
}

func (j JsonString) String() string {
	return string(j)
}

func (j *JsonString) UnmarshalJSON(data []byte) error {
	buffer := new(bytes.Buffer)
	err := json.Compact(buffer, data)
	if err != nil {
		log.Errorf("error compacting json: %s", string(data))
	}
	result := JsonString(buffer.String())
	*j = result
	return nil
}

func (j *JsonString) MarshalJSON() ([]byte, error) {
	val := string(*j)
	return []byte(val), nil
}

type InvalidStubMessage struct {
	Errors  []string `json:"errors"`
	Example Stub     `json:"example"`
}
