package plan

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"strings"

	"github.com/mi991203/TestMoon/tools/apimoon/jsonutil"
	"github.com/mi991203/TestMoon/tools/apimoon/platform"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/dynamicpb"
)

func digest(v any) string {
	b, _ := json.Marshal(v)
	h := sha256.Sum256(b)
	return hex.EncodeToString(h[:])
}

func assetHash(kind string, in map[string]any) string {
	out := jsonutil.Clone(in)
	delete(out, "id")
	delete(out, "revision")
	delete(out, "org_id")
	delete(out, "project_id")
	svc, m, _ := platform.ResourceMethod(kind, "create")
	b, err := platform.CanonicalRequest(svc, m, out)
	if err != nil {
		return digest(out)
	}
	p := dynamicpb.NewMessage(platform.Method(svc, m).Input())
	_ = protojson.Unmarshal(b, p)
	pruneEmptyMessages(p.ProtoReflect())
	b, _ = (protojson.MarshalOptions{UseProtoNames: true}).Marshal(p)
	var normalized map[string]any
	_ = jsonutil.Decode(b, &normalized)
	return digest(normalized)
}

// Storage round-trips may materialize empty request/auth messages. Normalize
// those defaults while preserving override authentication, optional scalar
// presence and arbitrary JSON values.
func pruneEmptyMessages(p protoreflect.Message) {
	p.Range(func(f protoreflect.FieldDescriptor, v protoreflect.Value) bool {
		if string(p.Descriptor().Name()) == "RequestPatch" && string(f.Name()) == "auth" {
			return true // An explicit empty auth clears inherited API authentication.
		}
		if f.IsMap() || f.Kind() != protoreflect.MessageKind {
			return true
		}
		prune := func(m protoreflect.Message) bool {
			if strings.HasPrefix(string(m.Descriptor().FullName()), "google.protobuf.") {
				return false
			}
			pruneEmptyMessages(m)
			empty := true
			m.Range(func(protoreflect.FieldDescriptor, protoreflect.Value) bool { empty = false; return false })
			return empty
		}
		if f.IsList() {
			list := v.List()
			for i := 0; i < list.Len(); i++ {
				prune(list.Get(i).Message())
			}
		} else if prune(v.Message()) {
			p.Clear(f)
		}
		return true
	})
}
