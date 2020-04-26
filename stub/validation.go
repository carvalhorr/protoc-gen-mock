package stub

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
)

type StubsValidator interface {
	IsValid(stub *Stub) (isValid bool, errorMessages []string)
}

func IsStubValid(stub *Stub, request, response reflect.Type) (isValid bool, errorMessages []string) {
	valid, errorMessages := stub.IsValid()
	if !valid {
		return valid, errorMessages
	}
	reqValid, reqErrorMessages := stub.Request.Content.isJsonValid(request, "request.content")
	respValid, respErrorMessages := stub.Response.Content.isJsonValid(response, "response.content")
	errorMessages = append(errorMessages, reqErrorMessages...)
	errorMessages = append(errorMessages, respErrorMessages...)
	return reqValid && respValid, errorMessages
}

func (j JsonString) isJsonValid(t reflect.Type, baseName string) (isValid bool, errorMessages []string) {
	jsonResult := new(map[string]interface{})
	err := json.Unmarshal([]byte(string(j)), jsonResult)
	if err != nil {
		return false, []string{"Invalid JSON."}
	}
	return isJsonValid(t, *jsonResult, baseName)
}

func isJsonValid(t reflect.Type, json map[string]interface{}, baseName string) (isValid bool, errorMessages []string) {
	errorMessages = make([]string, 0)
	reverseFields := make(map[string]reflect.StructField, 0)
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		jsonTag := field.Tag.Get("json")
		if jsonTag == "-" || jsonTag == "" {
			continue
		}
		jsonTag = strings.ReplaceAll(jsonTag, ",omitempty", "")
		reverseFields[jsonTag] = field
	}
	for jsonName, fieldValue := range json {
		field, ok := reverseFields[jsonName]
		if !ok {
			errorMessages = append(errorMessages, fmt.Sprintf("Field '%s.%s' does not exist", baseName, jsonName))
			continue
		}
		if fieldValue == nil {
			continue
		}
		switch {
		case field.Type.Kind() == reflect.String:
			switch fieldValue.(type) {
			case string:
			default:
				errorMessages = append(errorMessages, fmt.Sprintf("Field '%s.%s' is expected to be a string.", baseName, jsonName))
			}
		case field.Type.Kind() == reflect.Ptr:
			{
				_, subTypeErrorMessages := isJsonValid(field.Type.Elem(), fieldValue.(map[string]interface{}), baseName+"."+jsonName)
				errorMessages = append(errorMessages, subTypeErrorMessages...)
			}
		case isEnum(field.Type):
			enum := reflect.New(field.Type).Interface()
			values := getEnumValues(enum.(EnumType))
			found := false
			for _, value := range values {
				if value == fieldValue {
					found = true
				}
			}
			if !found {
				errorMessages = append(errorMessages, fmt.Sprintf("Value '%s' is not valid for field '%s.%s'. Possible values are '%s'.", fieldValue, baseName, jsonName, strings.Join(values, ", ")))
			}
		}
	}
	return len(errorMessages) == 0, errorMessages
}

func (stub *Stub) IsValid() (isValid bool, errMsgs []string) {
	if stub.FullMethod == "" {
		errMsgs = append(errMsgs, "Method can't be empty.")
	}
	// Validate request
	if stub.Request == nil {
		errMsgs = append(errMsgs, "Request can't be empty.")
	}
	if stub.Request.Content == "" {
		errMsgs = append(errMsgs, "Request content can't be empty.")
	}
	if stub.Request.Match != "exact" && stub.Request.Match != "partial" {
		errMsgs = append(errMsgs, "Request matching type can only be either 'exact' or 'partial'.")
	}
	// Validate response
	if stub.Response == nil {
		errMsgs = append(errMsgs, "Response can't be empty.")
	}
	if stub.Response.Content == "" {
		errMsgs = append(errMsgs, "Response content can't be empty.")
	}
	if stub.Response.Type != "error" && stub.Response.Type != "success" {
		errMsgs = append(errMsgs, "Response type can only be either 'error' or 'success'.")
	}
	if stub.Response.Type == "success" && stub.Response.Content == "" {
		errMsgs = append(errMsgs, "Response content is mandatory when the response type is 'success'.")
	}
	if stub.Response.Type == "error" && stub.Response.Error == "" {
		errMsgs = append(errMsgs, "Response error is mandatory when the response type ir 'error'.")
	}

	return len(errMsgs) == 0, errMsgs
}
