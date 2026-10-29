package platform

import (
	"encoding/json"
	"fmt"

	iam "github.com/mi991203/TestMoon/protocols/platform/identity/v1"
	core "github.com/mi991203/TestMoon/protocols/testing/core/v1"
	executionpb "github.com/mi991203/TestMoon/protocols/testing/execution/v1"
	"github.com/mi991203/TestMoon/tools/apimoon/clierror"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/dynamicpb"
)

// Resolve contracts and HTTP routes from the same descriptors as the platform.
func Method(service, name string) protoreflect.MethodDescriptor {
	var s protoreflect.ServiceDescriptor
	switch service {
	case "identity":
		s = iam.File_protocols_platform_identity_v1_service_proto.Services().ByName("Identity")
	case "core":
		s = core.File_protocols_testing_core_v1_service_proto.Services().ByName("Core")
	case "execution":
		s = executionpb.File_protocols_testing_execution_v1_service_proto.Services().ByName("Execution")
	}
	if s == nil {
		return nil
	}
	return s.Methods().ByName(protoreflect.Name(name))
}
func CanonicalRequest(service, name string, input map[string]any) ([]byte, error) {
	m := Method(service, name)
	if m == nil {
		return nil, clierror.New("local", "未知平台操作: "+service+"."+name)
	}
	b, err := json.Marshal(input)
	if err != nil {
		return nil, err
	}
	p := dynamicpb.NewMessage(m.Input())
	if err = protojson.Unmarshal(b, p); err != nil {
		return nil, clierror.New("validation", fmt.Sprintf("%s 请求不符合平台协议: %s", name, err))
	}
	return (protojson.MarshalOptions{UseProtoNames: true}).Marshal(p)
}
