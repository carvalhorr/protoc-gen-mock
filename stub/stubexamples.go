package stub

import (
	"bytes"
	"fmt"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func CreateStubExample(req proto.Message) string {
	// TODO make marshal work with child structs
	stack := make(map[string]bool)
	return generateJSONForType(req.ProtoReflect().Descriptor(), &bytes.Buffer{}, stack).String()
}

func generateJSONForType(t protoreflect.MessageDescriptor, writer *bytes.Buffer, stack map[string]bool) *bytes.Buffer {
	typeName := string(t.FullName())
	if stack[typeName] {
		writer.WriteString("{}")
		return writer
	}
	stack[typeName] = true
	writer.WriteString("{")
	generateJSONForField(t.Fields(), writer, stack, false)
	writer.WriteString("}")
	return writer
}

func generateJSONForField(fields protoreflect.FieldDescriptors, writer *bytes.Buffer, stack map[string]bool, isOneOf bool) *bytes.Buffer {
	first := true
	for i := 0; i < fields.Len(); i++ {
		field := fields.Get(i)
		if !field.HasJSONName() {
			continue
		}
		if !first {
			if !isOneOf && field.ContainingOneof() != nil {
				oneOfName := string(field.ContainingOneof().Name())
				if stack[oneOfName] {
					continue
				}
			}
			writer.WriteString(", ")
		}
		first = false
		if !isOneOf && field.ContainingOneof() != nil {
			generateJSONForoneOf(field.ContainingOneof(), writer, stack)
			first = false
			continue
		}
		writer.WriteString(fmt.Sprintf("\"%s\": ", field.JSONName()))
		if field.Cardinality() == protoreflect.Repeated {
			writer.WriteString(" [")
		}
		switch field.Kind() {
		case protoreflect.MessageKind:
			generateJSONForType(field.Message(), writer, stack)
		case protoreflect.StringKind:
			writer.WriteString("\"\"")
		case protoreflect.BoolKind:
			writer.WriteString(" true")
		case protoreflect.EnumKind:
			writer.WriteString(fmt.Sprintf("\"%s\"", field.Enum().Values()))
		default:
			writer.WriteString(" 0")
		}
		if field.Cardinality() == protoreflect.Repeated {
			writer.WriteString("]")
		}
	}
	return writer
}

func generateJSONForoneOf(t protoreflect.OneofDescriptor, writer *bytes.Buffer, stack map[string]bool) *bytes.Buffer {
	typeName := string(t.Name())
	if stack[typeName] {
		return writer
	}
	stack[typeName] = true
	writer.WriteString(fmt.Sprintf("\"%s\": { \"oneof\": {", typeName))
	generateJSONForField(t.Fields(), writer, stack, true)
	writer.WriteString("}}")
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
