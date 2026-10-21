package api

import (
	"google.golang.org/protobuf/proto"

	"github.com/mi991203/TestMoon/backend/testing/core/logic"
	p "github.com/mi991203/TestMoon/libs/go/platform"
)

func resourceWriteInput(request proto.Message, kind string) (*logic.ResourceInput, error) {
	input := &logic.ResourceInput{}
	if err := p.Decode(request, input); err != nil {
		return nil, err
	}
	input.Kind = kind
	fields := struct {
		Name     string         `validate:"required,max=128"`
		Document map[string]any `validate:"required"`
	}{Name: input.Name, Document: input.Document}
	if err := p.Check(fields); err != nil {
		return nil, err
	}
	return input, nil
}
