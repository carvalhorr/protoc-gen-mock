package stub

import (
	"bytes"
	"fmt"
	"github.com/gogo/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"reflect"
	"strings"
)

func CreateStubExample(req proto.Message) string {
	// TODO make marshal work with child structs
	return generateJSONForType(reflect.TypeOf(req).Elem(), &bytes.Buffer{}).String()
	/*
		marshaler := jsonpb.Marshaler{EmitDefaults: true}
		strJSON, err := marshaler.MarshalToString(req)
		if err != nil {
			return fmt.Sprintf("Error getting JSON: %s", err.Error())
		}
		return strJSON
	*/
}

func generateJSONForType(t reflect.Type, writer *bytes.Buffer) *bytes.Buffer {
	if t.Kind() != reflect.Struct || t.NumField() == 0 {
		return writer
	}
	writer.WriteString("{")
	first := true
	for i := 0; i < t.NumField(); i++ {
		json, ok := t.Field(i).Tag.Lookup("json")
		if !ok {
			continue
		}
		if json == "-" {
			continue
		}
		json = strings.Replace(json, ",omitempty", "", 1)
		if !first {
			writer.WriteString(", ")
		}
		first = false
		switch t.Field(i).Type.Kind() {
		case reflect.Ptr:
			writer.WriteString(fmt.Sprintf("\"%s\": ", json))
			generateJSONForType(t.Field(i).Type.Elem(), writer)
		case reflect.Struct:
			writer.WriteString(fmt.Sprintf("\"%s\": ", json))
			generateJSONForType(t.Field(i).Type, writer)
		case reflect.String:
			writer.WriteString(fmt.Sprintf("\"%s\": \"\"", json))
		case reflect.Array:
			writer.WriteString(fmt.Sprintf("\"%s\": [ARRAY]", json)) // Should not happen, leaving ARRAY to indicate to the consumer that it may need work
			generateJSONForType(t.Field(i).Type, writer)
		case reflect.Slice:
			writer.WriteString(fmt.Sprintf("\"%s\": [", json))
			switch t.Field(i).Type.Elem().Kind() {
			case reflect.Struct:
				generateJSONForType(t.Field(i).Type.Elem(), writer)
			case reflect.Ptr:
				generateJSONForType(t.Field(i).Type.Elem().Elem(), writer)
			}
			writer.WriteString("]")
		case reflect.Map:
			writer.WriteString(fmt.Sprintf("\"%s\": MAP", json)) // Should not happen, leaving MAP to indicate to the consumer that it may need work
		case reflect.Bool:
			writer.WriteString(fmt.Sprintf("\"%s\": true", json))
		default:
			if isEnum(t.Field(i).Type) {
				val := reflect.New(t.Field(i).Type).Interface()
				values := getEnumValues(val.(EnumType))
				writer.WriteString(fmt.Sprintf("\"%s\": \"%s\"", json, strings.Join(values, " | ")))
				continue
			}
			writer.WriteString(fmt.Sprintf("\"%s\": 0", json))
		}
	}
	writer.WriteString("}")
	return writer
}

func getEnumValues(enum EnumType) []string {
	values := make([]string, 0)
	for i := 0; i < enum.Descriptor().Values().Len(); i++ {
		val := enum.Descriptor().Values().ByNumber(protoreflect.EnumNumber(i))
		values = append(values, string(val.Name()))
	}
	return values
}
