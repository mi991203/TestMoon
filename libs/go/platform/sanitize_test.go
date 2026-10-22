package platform_test

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	p "github.com/mi991203/TestMoon/libs/go/platform"
	m "github.com/mi991203/TestMoon/libs/go/testing/domain"
)

func TestSanitizePreservesResponseHeaderValues(t *testing.T) {
	response := m.Response{Status: 200, Headers: map[string][]string{
		"Set-Cookie":   {"session=private", "tracking=private"},
		"X-Api-Key":    {"private-key"},
		"Content-Type": {"text/html"},
	}}
	raw, err := json.Marshal(p.Sanitize(response, nil))
	if err != nil {
		t.Fatal(err)
	}
	var safe m.Response
	if err := json.Unmarshal(raw, &safe); err != nil {
		t.Fatalf("sanitized response must retain its domain types: %v", err)
	}
	want := map[string][]string{"Set-Cookie": {"***", "***"}, "X-Api-Key": {"***"}, "Content-Type": {"text/html"}}
	if safe.Status != 200 || !reflect.DeepEqual(safe.Headers, want) {
		t.Fatalf("unexpected sanitized response: %#v", safe)
	}
	if strings.Contains(string(raw), "private") {
		t.Fatal("sanitized response contains a secret")
	}
	if response.Headers["Set-Cookie"][0] != "session=private" {
		t.Fatal("sanitization mutated the original response")
	}
}

func TestSanitizeRetainsScalarRedaction(t *testing.T) {
	input := map[string]any{
		"authorization": "Bearer private-token",
		"variable":      m.Variable{Secret: true, Value: "private-value"},
		"auth":          m.Auth{Type: "api_key", Key: "X-Key", Value: "private-key"},
	}
	safe := p.Sanitize(input, nil).(map[string]any)
	if safe["authorization"] != "***" {
		t.Fatal("authorization was not redacted")
	}
	variable := safe["variable"].(map[string]any)
	if variable["secret"] != true || variable["value"] != "***" {
		t.Fatalf("secret variable changed shape or leaked: %#v", variable)
	}
	if safe["auth"].(map[string]any)["value"] != "***" {
		t.Fatal("API key was not redacted")
	}
}
