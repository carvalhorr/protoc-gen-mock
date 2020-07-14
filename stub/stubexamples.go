package stub

import (
	"bytes"
	"fmt"
	"google.golang.org/protobuf/proto"
	"reflect"
	"strings"
)

func CreateStubExample(req proto.Message) string {
	// TODO make marshal work with child structs
	stack := make(map[string]bool)
	return generateJSONForType(reflect.TypeOf(req).Elem(), &bytes.Buffer{}, stack).String()
}

func generateJSONForType(t reflect.Type, writer *bytes.Buffer, stack map[string]bool) *bytes.Buffer {
	if t.Kind() != reflect.Struct || t.NumField() == 0 {
		return writer
	}
	typeName := t.String()
	if stack[typeName] {
		writer.WriteString("{}")
		return writer
	}
	stack[typeName] = true
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
			generateJSONForType(t.Field(i).Type.Elem(), writer, stack)
		case reflect.Struct:
			writer.WriteString(fmt.Sprintf("\"%s\": ", json))
			generateJSONForType(t.Field(i).Type, writer, stack)
		case reflect.String:
			writer.WriteString(fmt.Sprintf("\"%s\": \"\"", json))
		case reflect.Array:
			writer.WriteString(fmt.Sprintf("\"%s\": [ARRAY]", json)) // Should not happen, leaving ARRAY to indicate to the consumer that it may need work
			generateJSONForType(t.Field(i).Type, writer, stack)
		case reflect.Slice:
			writer.WriteString(fmt.Sprintf("\"%s\": [", json))
			switch t.Field(i).Type.Elem().Kind() {
			case reflect.Struct:
				generateJSONForType(t.Field(i).Type.Elem(), writer, stack)
			case reflect.Ptr:
				generateJSONForType(t.Field(i).Type.Elem().Elem(), writer, stack)
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
		val := enum.Descriptor().Values().Get(i)
		values = append(values, string(val.Name()))
	}
	return values
}
