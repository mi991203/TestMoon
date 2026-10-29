package plan

import (
	"path/filepath"
	"testing"

	"github.com/mi991203/TestMoon/tools/apimoon/config"
	"github.com/mi991203/TestMoon/tools/apimoon/platform"
)

func TestStateLock(t *testing.T) {
	s := Syncer{Client: platform.New(config.Config{BaseURL: "http://localhost", Auth: config.AuthConfig{UserID: "user"}}), ConfigFile: filepath.Join(t.TempDir(), "config.yaml")}
	unlock, err := s.lock("plan")
	if err != nil {
		t.Fatal(err)
	}
	defer unlock()
	if _, err = s.lock("plan"); err == nil {
		t.Fatal("concurrent plan lock acquired")
	}
}
