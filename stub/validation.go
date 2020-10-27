package stub

import (
	"encoding/json"
	"fmt"
	"google.golang.org/protobuf/reflect/protoreflect"
)

type StubsValidator interface {
	IsValid(stub *Stub) (isValid bool, errorMessages []string)
}

func NewCompositeStubsValidator(validators []StubsValidator) StubsValidator {
	return compositeStubsValidator{
		stubsValidators: validators,
	}
}

type compositeStubsValidator struct {
	stubsValidators []StubsValidator
}

func (c compositeStubsValidator) IsValid(stub *Stub) (isValid bool, errorMessages []string) {
	for _, validator := range c.stubsValidators {
		isValid, errorMessages = validator.IsValid(stub)
		if !isValid {
			return isValid, errorMessages
		}
	}
	return true, nil
}

func IsStubValid(stub *Stub, request, response protoreflect.MessageDescriptor) (isValid bool, errorMessages []string) {
	valid, errorMessages := stub.IsValid()
	if !valid {
		return valid, errorMessages
	}
	reqValid, reqErrorMessages := stub.Request.Content.isJsonValid(request, "request.content")
	respValid := true
	respErrorMessages := make([]string, 0)
	if stub.Type == "mock" && stub.Response.Type == "success" {
		respValid, respErrorMessages = stub.Response.Content.isJsonValid(response, "response.content")
	}
	errorMessages = append(errorMessages, reqErrorMessages...)
	errorMessages = append(errorMessages, respErrorMessages...)
	return reqValid && respValid, errorMessages
}

func (j JsonString) isJsonValid(t protoreflect.MessageDescriptor, baseName string) (isValid bool, errorMessages []string) {
	jsonResult := new(map[string]interface{})
	err := json.Unmarshal([]byte(string(j)), jsonResult)
	if err != nil {
		return false, []string{fmt.Sprintf("%s: invalid JSON", baseName)}
	}
	return isJsonValid(t, *jsonResult, baseName)
}

func isJsonValid(t protoreflect.MessageDescriptor, json map[string]interface{}, baseName string) (isValid bool, errorMessages []string) {
	errorMessages = make([]string, 0)
	reverseFields := make(map[string]protoreflect.FieldDescriptor, 0)
	for i := 0; i < t.Fields().Len(); i++ {
		field := t.Fields().Get(i)
		if !field.HasJSONName() {
			continue
		}

		jsonTag := field.JSONName()
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
		case field.Kind() == protoreflect.StringKind:
			switch fieldValue.(type) {
			case string:
			default:
				errorMessages = append(errorMessages, fmt.Sprintf("Field '%s.%s' is expected to be a string.", baseName, jsonName))
			}
		case field.Kind() == protoreflect.MessageKind:
			_, subTypeErrorMessages := isJsonValid(field.Message(), fieldValue.(map[string]interface{}), baseName+"."+jsonName)
			errorMessages = append(errorMessages, subTypeErrorMessages...)
		case field.Kind() == protoreflect.EnumKind:
			found := false
			for i := 0; i < field.Enum().Values().Len(); i++ {
				value := field.Enum().Values().Get(i)
				if value == fieldValue {
					found = true
					break // TODO make sure break is leaving the for loop
				}
			}
			if !found {
				// TODO implement the correct names
				errorMessages = append(errorMessages, fmt.Sprintf("Value '%s' is not valid for field '%s.%s'. Possible values are '%s'.", fieldValue, baseName, jsonName, field.Enum().ReservedNames()))
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
	requestValid, requestErrMsgs := stub.isValidRequest()
	isValid = isValid && requestValid
	errMsgs = append(errMsgs, requestErrMsgs...)

	// Response type can be either 'mock' or 'forward'
	if stub.Type != "mock" && stub.Type != "forward" {
		errMsgs = append(errMsgs, "Stub type must be either 'mock' or 'forward'")
	}

	// Validate response
	responseValid, responseErrMsgs := stub.isValidResponse()
	isValid = isValid && responseValid
	errMsgs = append(errMsgs, responseErrMsgs...)

	// Validate forward
	forwardValid, forwardErrMsgs := stub.isValidForward()
	isValid = isValid && forwardValid
	errMsgs = append(errMsgs, forwardErrMsgs...)

	return len(errMsgs) == 0, errMsgs
}

func (stub *Stub) isValidRequest() (isValid bool, errMsgs []string) {
	if stub.Request == nil {
		errMsgs = append(errMsgs, "Request can't be empty.")
	}
	if stub.Request.Content == "" {
		errMsgs = append(errMsgs, "Request content can't be empty.")
	}
	if stub.Request.Match != "exact" && stub.Request.Match != "partial" {
		errMsgs = append(errMsgs, "Request matching type can only be either 'exact' or 'partial'.")
	}
	return len(errMsgs) == 0, errMsgs
}

func (stub *Stub) isValidResponse() (isValid bool, errMsgs []string) {
	if stub.Type != "mock" {
		return true, nil
	}

	if stub.Forward != nil {
		errMsgs = append(errMsgs, "Stub must not contain a forward definition if it's type is 'mock'")
	}

	if stub.Response == nil {
		errMsgs = append(errMsgs, "Response can't be empty when stub's type is 'mock'.")
		return false, errMsgs
	}
	if stub.Response.Type != "error" && stub.Response.Type != "success" {
		errMsgs = append(errMsgs, "Response type can only be either 'error' or 'success'.")
	}
	if stub.Response.Type == "success" && stub.Response.Content == "" {
		errMsgs = append(errMsgs, "Response content is mandatory when the response type is 'success'.")
	}
	if stub.Response.Type == "error" && stub.Response.Error == nil {
		errMsgs = append(errMsgs, "Response error is mandatory when the response type ir 'error'.")
	}
	return len(errMsgs) == 0, errMsgs
}

func (stub *Stub) isValidForward() (isValid bool, errMsgs []string) {
	if stub.Type != "forward" {
		return true, nil
	}

	if stub.Response != nil {
		errMsgs = append(errMsgs, "Stub must not contain a response definition if it's type is 'forward'")
	}

	if stub.Forward == nil {
		errMsgs = append(errMsgs, "Stub must not contain a forward definition if it's type is 'forward'")
		return false, errMsgs
	}

	if stub.Forward.ServerAddress == "" {
		errMsgs = append(errMsgs, "You must provide a server address for forwarding stub types.")
	}
	return len(errMsgs) == 0, errMsgs
}
