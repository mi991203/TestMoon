package plan

import (
	"github.com/mi991203/TestMoon/tools/apimoon/platform"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// Schema shares field types with the platform protobuf descriptors, including
// recursive orchestration and protobuf's JSON representation of int64 fields.
func Schema() map[string]any {
	defs := map[string]any{}
	var message func(protoreflect.MessageDescriptor) map[string]any
	var field func(protoreflect.FieldDescriptor) map[string]any
	field = func(f protoreflect.FieldDescriptor) map[string]any {
		if f.IsMap() {
			return map[string]any{"type": "object", "additionalProperties": field(f.MapValue())}
		}
		var s map[string]any
		switch f.Kind() {
		case protoreflect.MessageKind:
			s = message(f.Message())
		case protoreflect.StringKind:
			s = map[string]any{"type": "string"}
		case protoreflect.BoolKind:
			s = map[string]any{"type": "boolean"}
		case protoreflect.Int64Kind, protoreflect.Uint64Kind:
			s = map[string]any{"oneOf": []any{map[string]any{"type": "integer"}, map[string]any{"type": "string", "pattern": "^-?[0-9]+$"}}}
		default:
			s = map[string]any{"type": "integer"}
		}
		if f.IsList() {
			return map[string]any{"type": "array", "items": s}
		}
		return s
	}
	message = func(d protoreflect.MessageDescriptor) map[string]any {
		name := string(d.FullName())
		if name == "google.protobuf.Struct" {
			return map[string]any{"type": "object"}
		}
		if name == "google.protobuf.Value" {
			return map[string]any{}
		}
		ref := map[string]any{"$ref": "#/$defs/" + name}
		if _, ok := defs[name]; ok {
			return ref
		}
		defs[name] = map[string]any{}
		properties := map[string]any{}
		for i := 0; i < d.Fields().Len(); i++ {
			f := d.Fields().Get(i)
			properties[string(f.Name())] = field(f)
		}
		s := map[string]any{"type": "object", "additionalProperties": false, "properties": properties}
		switch string(d.Name()) {
		case "APIDefinition":
			s["required"] = []string{"protocol", "request"}
			properties["protocol"] = map[string]any{"const": "http"}
		case "HTTPRequest":
			s["required"] = []string{"method", "url"}
		case "ScenarioDefinition":
			s["required"] = []string{"steps"}
		case "TaskDefinition":
			s["required"] = []string{"items"}
		case "Operation":
			s["required"] = []string{"type", "version", "script"}
		case "Assertion":
			s["required"] = []string{"source", "operator"}
		case "Extraction":
			s["required"] = []string{"source", "variable"}
		}
		defs[name] = s
		return ref
	}
	variants := []any{}
	for _, kind := range []string{"api", "case", "scenario", "task", "environment", "folder"} {
		svc, m, _ := platform.ResourceMethod(kind, "create")
		doc := platform.Method(svc, m).Input().Fields().ByName("document").Message()
		props := map[string]any{"key": map[string]any{"type": "string", "minLength": 1, "maxLength": 128}, "kind": map[string]any{"const": kind}, "id": map[string]any{"type": "string", "minLength": 1}, "revision": map[string]any{"oneOf": []any{map[string]any{"type": "integer", "minimum": 1}, map[string]any{"type": "string", "pattern": "^[1-9][0-9]*$"}}}, "name": map[string]any{"type": "string", "minLength": 1, "maxLength": 128}, "document": message(doc)}
		if kind != "task" {
			props["folder_id"] = map[string]any{"type": "string"}
		}
		variants = append(variants, map[string]any{"type": "object", "additionalProperties": false, "properties": props, "required": []string{"key", "kind", "name", "document"}, "dependentRequired": map[string]any{"id": []string{"revision"}, "revision": []string{"id"}}})
	}
	return map[string]any{"$schema": "https://json-schema.org/draft/2020-12/schema", "title": "APIMoon TestPlan v1", "type": "object", "additionalProperties": false, "required": []string{"version", "plan_id", "entries"}, "properties": map[string]any{"version": map[string]any{"const": 1}, "plan_id": map[string]any{"type": "string", "minLength": 1, "maxLength": 128}, "source": map[string]any{"type": "object", "additionalProperties": false, "properties": map[string]any{"requirement": map[string]any{"type": "string"}, "code_ref": map[string]any{"type": "string"}}}, "entries": map[string]any{"type": "array", "minItems": 1, "maxItems": 2000, "items": map[string]any{"oneOf": variants}}}, "$defs": defs}
}
