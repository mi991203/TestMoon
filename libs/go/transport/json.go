// Package transport converts typed domain records to protobuf JSON contracts.
package transport

import (
	"encoding/json"
	"fmt"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Fields is for boundary serialization, never for operation dispatch. It retains
// optional field presence, native integer revisions and RFC3339 timestamps.
func Fields(value any) map[string]any {
	if message, ok := value.(proto.Message); ok {
		out, _ := messageJSON(message).(map[string]any)
		return out
	}
	raw, _ := json.Marshal(value)
	var out map[string]any
	_ = json.Unmarshal(raw, &out)
	return out
}
func messageJSON(message proto.Message) any {
	switch v := message.(type) {
	case *structpb.Struct:
		return v.AsMap()
	case *structpb.Value:
		return v.AsInterface()
	case *timestamppb.Timestamp:
		return v.AsTime().UTC().Format("2006-01-02T15:04:05.999999999Z07:00")
	}
	if message.ProtoReflect().Descriptor().FullName() == "testmoon.testing.contracts.v1.HeaderValues" {
		values := message.ProtoReflect().Get(message.ProtoReflect().Descriptor().Fields().ByName("values")).List()
		out := []any{}
		for i := 0; i < values.Len(); i++ {
			out = append(out, values.Get(i).String())
		}
		return out
	}
	result := map[string]any{}
	message.ProtoReflect().Range(func(field protoreflect.FieldDescriptor, value protoreflect.Value) bool {
		decode := func(v protoreflect.Value, d protoreflect.FieldDescriptor) any {
			if d.Kind() == protoreflect.MessageKind {
				return messageJSON(v.Message().Interface())
			}
			if d.Kind() == protoreflect.EnumKind {
				return string(d.Enum().Values().ByNumber(v.Enum()).Name())
			}
			return v.Interface()
		}
		if field.IsMap() {
			out := map[string]any{}
			value.Map().Range(func(k protoreflect.MapKey, v protoreflect.Value) bool {
				out[k.String()] = decode(v, field.MapValue())
				return true
			})
			result[string(field.Name())] = out
		} else if field.IsList() {
			out := []any{}
			for i := 0; i < value.List().Len(); i++ {
				out = append(out, decode(value.List().Get(i), field))
			}
			result[string(field.Name())] = out
		} else {
			result[string(field.Name())] = decode(value, field)
		}
		return true
	})
	return result
}
func Encode(value any, out proto.Message) error {
	raw, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("serialize domain response: %w", err)
	}
	var data any
	if err = json.Unmarshal(raw, &data); err != nil {
		return err
	}
	data = normalize(data, out.ProtoReflect().Descriptor())
	raw, err = json.Marshal(data)
	if err != nil {
		return err
	}
	return (protojson.UnmarshalOptions{DiscardUnknown: true}).Unmarshal(raw, out)
}
func normalize(value any, desc protoreflect.MessageDescriptor) any {
	if desc.FullName() == "testmoon.testing.contracts.v1.HeaderValues" {
		if values, ok := value.([]any); ok {
			return map[string]any{"values": values}
		}
	}
	data, ok := value.(map[string]any)
	if !ok {
		return value
	}
	// Import entries have a finite set of typed variants rather than an opaque document.
	if desc.FullName() == "testmoon.testing.core.v1.ImportEntry" {
		if doc, ok := data["document"]; ok {
			kind, _ := data["kind"].(string)
			key := map[string]string{"api": "api", "case": "test_case", "scenario": "scenario", "environment": "environment", "folder": "folder_definition"}[kind]
			if key != "" {
				data[key] = doc
			}
			delete(data, "document")
		}
	}
	for key, v := range data {
		field := desc.Fields().ByName(protoreflect.Name(key))
		if field == nil {
			continue
		}
		if field.IsMap() {
			if field.MapValue().Kind() == protoreflect.MessageKind {
				if entries, ok := v.(map[string]any); ok {
					for k, item := range entries {
						entries[k] = normalize(item, field.MapValue().Message())
					}
				}
			}
		} else if field.Kind() == protoreflect.MessageKind {
			if field.IsList() {
				if entries, ok := v.([]any); ok {
					for i, item := range entries {
						entries[i] = normalize(item, field.Message())
					}
				}
			} else {
				data[key] = normalize(v, field.Message())
			}
		}
	}
	return data
}
