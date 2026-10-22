package engine

import (
	"context"
	"reflect"
	"strings"
	"testing"

	p "github.com/mi991203/TestMoon/libs/go/platform"
	m "github.com/mi991203/TestMoon/libs/go/testing/domain"
	"github.com/mi991203/TestMoon/libs/go/transport"
	contracts "github.com/mi991203/TestMoon/protocols/testing/contracts/v1"
	pb "github.com/mi991203/TestMoon/protocols/testing/execution/v1"
	"google.golang.org/protobuf/encoding/protojson"
)

func TestCookieResponseCanBeReported(t *testing.T) {
	for _, withAssertion := range []bool{false, true} {
		name, wantStatus := "without assertion", "unverified"
		if withAssertion {
			name, wantStatus = "with assertion", "passed"
		}
		t.Run(name, func(t *testing.T) {
			c := m.Case{Request: m.HTTPRequest{Method: "GET", URL: "https://www.baidu.com"}}
			if withAssertion {
				c.Assertions = []m.Assertion{{Source: "status", Operator: "eq", Expected: 200}}
			}
			e := Engine{Transport: func(context.Context, string, m.HTTPRequest) m.Response {
				return m.Response{Status: 200, Headers: map[string][]string{
					"Set-Cookie":   {"session=private", "tracking=private"},
					"Content-Type": {"text/html"},
				}, Body: "<html>ok</html>"}
			}}
			result := e.Run(context.Background(), m.Snapshot{Version: 1, Items: []m.CompiledItem{
				{ID: "item", Steps: []m.CompiledStep{{ID: "step", Case: c}}},
			}})
			if result.Status != wantStatus || result.Items[0].Steps[0].Status != wantStatus {
				t.Fatalf("unexpected result status: %#v", result)
			}
			value := &contracts.ExecutionResult{}
			if err := transport.Encode(result, value); err != nil {
				t.Fatalf("cannot encode execution result for delivery: %v", err)
			}
			raw, err := (protojson.MarshalOptions{UseProtoNames: true}).Marshal(&pb.CompleteRunRequest{Id: "run", Lease: "lease", Result: value})
			if err != nil {
				t.Fatal(err)
			}
			if strings.Contains(string(raw), "private") {
				t.Fatal("execution report contains raw cookies")
			}
			request := &pb.CompleteRunRequest{}
			if err := protojson.Unmarshal(raw, request); err != nil {
				t.Fatal(err)
			}
			var received struct {
				Result m.Result `json:"result"`
			}
			if err := p.Decode(request, &received); err != nil {
				t.Fatalf("server cannot decode execution report: %v", err)
			}
			// The execution service sanitizes the result a second time before storage.
			stored := &contracts.ExecutionResult{}
			if err := transport.Encode(p.Sanitize(received.Result, nil), stored); err != nil {
				t.Fatalf("stored report cannot be returned to the client: %v", err)
			}
			response := stored.Items[0].Steps[0].Response
			if response.Status != 200 || !reflect.DeepEqual(response.Headers["Set-Cookie"].Values, []string{"***", "***"}) {
				t.Fatalf("response headers or status were corrupted: %#v", response)
			}
		})
	}
}
