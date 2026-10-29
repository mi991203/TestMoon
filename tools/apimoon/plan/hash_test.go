package plan

import (
	"testing"

	"github.com/mi991203/TestMoon/tools/apimoon/jsonutil"
)

func TestAssetHashNormalizesDefaultsButPreservesOptionalPatch(t *testing.T) {
	a := map[string]any{"name": "api", "document": map[string]any{"protocol": "http", "request": map[string]any{"method": "GET", "url": "https://example.test"}}}
	b := jsonutil.Clone(a)
	jsonutil.Object(jsonutil.Object(b["document"])["request"])["auth"] = map[string]any{}
	jsonutil.Object(b["document"])["schema"] = nil
	if assetHash("api", a) != assetHash("api", b) {
		t.Fatal("empty messages affect identity")
	}
	x := map[string]any{"name": "case", "document": map[string]any{"api_id": "api"}}
	y := jsonutil.Clone(x)
	jsonutil.Object(y["document"])["override"] = map[string]any{"body": ""}
	if assetHash("case", x) == assetHash("case", y) {
		t.Fatal("explicit empty patch collapsed")
	}
}

func TestAssetHashPreservesEmptyOverrideAuth(t *testing.T) {
	a := map[string]any{"name": "case", "document": map[string]any{"api_id": "api"}}
	b := jsonutil.Clone(a)
	jsonutil.Object(b["document"])["override"] = map[string]any{"auth": map[string]any{}}
	if assetHash("case", a) == assetHash("case", b) {
		t.Fatal("explicit empty authentication patch collapsed")
	}
}
