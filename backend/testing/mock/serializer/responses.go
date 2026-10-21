package serializer

import (
	"github.com/mi991203/TestMoon/libs/go/transport"
	pb "github.com/mi991203/TestMoon/protocols/testing/mock/v1"
)

func ListEndpoints(value any) (*pb.ListEndpointsResponse, error) {
	out := &pb.ListEndpointsResponse{}
	if err := transport.Encode(value, out); err != nil {
		return nil, err
	}
	return out, nil
}

func ListRules(value any) (*pb.ListRulesResponse, error) {
	out := &pb.ListRulesResponse{}
	if err := transport.Encode(value, out); err != nil {
		return nil, err
	}
	return out, nil
}

func CreateEndpoint(value any) (*pb.CreateEndpointResponse, error) {
	out := &pb.CreateEndpointResponse{}
	if err := transport.Encode(value, out); err != nil {
		return nil, err
	}
	return out, nil
}

func UpdateEndpoint(value any) (*pb.UpdateEndpointResponse, error) {
	out := &pb.UpdateEndpointResponse{}
	if err := transport.Encode(value, out); err != nil {
		return nil, err
	}
	return out, nil
}

func DeleteEndpoint(value any) (*pb.DeleteEndpointResponse, error) {
	out := &pb.DeleteEndpointResponse{}
	if err := transport.Encode(value, out); err != nil {
		return nil, err
	}
	return out, nil
}

func CreateRule(value any) (*pb.CreateRuleResponse, error) {
	out := &pb.CreateRuleResponse{}
	if err := transport.Encode(value, out); err != nil {
		return nil, err
	}
	return out, nil
}

func UpdateRule(value any) (*pb.UpdateRuleResponse, error) {
	out := &pb.UpdateRuleResponse{}
	if err := transport.Encode(value, out); err != nil {
		return nil, err
	}
	return out, nil
}

func DeleteRule(value any) (*pb.DeleteRuleResponse, error) {
	out := &pb.DeleteRuleResponse{}
	if err := transport.Encode(value, out); err != nil {
		return nil, err
	}
	return out, nil
}
