package protoconv

import (
	"encoding/json"
	"fmt"
	"strings"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// JSONToProto converts JSON data to a protobuf message
func JSONToProto(msg proto.Message, data []byte) error {
	// First try direct unmarshal for simple cases
	if err := json.Unmarshal(data, msg); err == nil {
		return nil
	}

	// Fallback to field-by-field mapping
	var jsonData map[string]interface{}
	if err := json.Unmarshal(data, &jsonData); err != nil {
		return fmt.Errorf("failed to parse JSON: %v", err)
	}

	msgReflect := msg.ProtoReflect()
	fields := msgReflect.Descriptor().Fields()

	for i := 0; i < fields.Len(); i++ {
		field := fields.Get(i)
		jsonName := toJSONName(string(field.Name()))
		
		value, exists := jsonData[jsonName]
		if !exists {
			continue
		}

		switch field.Kind() {
		case protoreflect.BoolKind:
			if b, ok := value.(bool); ok {
				msgReflect.Set(field, protoreflect.ValueOfBool(b))
			}
		case protoreflect.Int32Kind, protoreflect.Int64Kind,
			protoreflect.Sint32Kind, protoreflect.Sint64Kind,
			protoreflect.Sfixed32Kind, protoreflect.Sfixed64Kind:
			if n, ok := value.(float64); ok {
				msgReflect.Set(field, protoreflect.ValueOfInt64(int64(n)))
			}
		case protoreflect.Uint32Kind, protoreflect.Uint64Kind,
			protoreflect.Fixed32Kind, protoreflect.Fixed64Kind:
			if n, ok := value.(float64); ok {
				msgReflect.Set(field, protoreflect.ValueOfUint64(uint64(n)))
			}
		case protoreflect.FloatKind, protoreflect.DoubleKind:
			if f, ok := value.(float64); ok {
				msgReflect.Set(field, protoreflect.ValueOfFloat64(f))
			}
		case protoreflect.StringKind:
			if s, ok := value.(string); ok {
				msgReflect.Set(field, protoreflect.ValueOfString(s))
			}
		case protoreflect.BytesKind:
			if b, ok := value.(string); ok {
				msgReflect.Set(field, protoreflect.ValueOfBytes([]byte(b)))
			}
		case protoreflect.EnumKind:
			if s, ok := value.(string); ok {
				enumVal := field.Enum().Values().ByName(protoreflect.Name(s))
				if enumVal != nil {
					msgReflect.Set(field, protoreflect.ValueOfEnum(enumVal.Number()))
				}
			}
		case protoreflect.MessageKind:
			if m, ok := value.(map[string]interface{}); ok {
				subMsg := msgReflect.NewField(field).Message().Interface()
				subData, _ := json.Marshal(m)
				if err := JSONToProto(subMsg, subData); err == nil {
					msgReflect.Set(field, protoreflect.ValueOfMessage(subMsg.ProtoReflect()))
				}
			}
		}
	}

	return nil
}

// toJSONName converts protobuf field name to common JSON naming conventions
func toJSONName(fieldName string) string {
	if len(fieldName) == 0 {
		return fieldName
	}
	return strings.ToLower(fieldName[:1]) + fieldName[1:]
}
