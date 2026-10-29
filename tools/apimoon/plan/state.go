package plan

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/mi991203/TestMoon/tools/apimoon/clierror"
	"github.com/mi991203/TestMoon/tools/apimoon/config"
	"github.com/mi991203/TestMoon/tools/apimoon/fileio"
	"github.com/mi991203/TestMoon/tools/apimoon/jsonutil"
	"github.com/mi991203/TestMoon/tools/apimoon/platform"
)

type Binding struct {
	Kind        string `json:"kind"`
	ID          string `json:"id,omitempty"`
	Revision    string `json:"revision,omitempty"`
	Hash        string `json:"hash,omitempty"`
	State       string `json:"state"`
	PendingHash string `json:"pending_hash,omitempty"`
}
type PlanState struct {
	PlanID    string             `json:"plan_id"`
	UpdatedAt time.Time          `json:"updated_at"`
	Entries   map[string]Binding `json:"entries"`
}
type Syncer struct {
	Client     *platform.Client
	ConfigFile string
}

func (s Syncer) stateFile(planID string) string {
	c := s.Client.Config
	base, _ := config.NormalizeURL(c.BaseURL)
	key := digest([]string{base, c.Auth.UserID, c.Context.OrgID, c.Context.ProjectID, planID})
	return filepath.Join(filepath.Dir(s.ConfigFile), "state", key+".json")
}
func (s Syncer) Status(planID string) (PlanState, error) {
	state := PlanState{PlanID: planID, Entries: map[string]Binding{}}
	b, err := os.ReadFile(s.stateFile(planID))
	if errors.Is(err, os.ErrNotExist) {
		return state, nil
	}
	if err != nil {
		return state, err
	}
	err = jsonutil.Decode(b, &state)
	if state.Entries == nil {
		state.Entries = map[string]Binding{}
	}
	if err == nil && state.PlanID != planID {
		err = clierror.New("conflict", "计划状态标识不一致")
	}
	return state, err
}
func (s Syncer) save(state *PlanState) error {
	state.UpdatedAt = time.Now().UTC()
	return fileio.WriteJSON(s.stateFile(state.PlanID), state)
}
func (s Syncer) lock(planID string) (func(), error) {
	file := s.stateFile(planID) + ".lock"
	if err := os.MkdirAll(filepath.Dir(file), 0700); err != nil {
		return nil, err
	}
	f, err := os.OpenFile(file, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		return nil, clierror.New("conflict", "计划正在同步；若进程异常退出，请确认无同步进程后移除锁文件: "+file)
	}
	_, err = fmt.Fprintf(f, "pid=%d\n", os.Getpid())
	_ = f.Close()
	if err != nil {
		_ = os.Remove(file)
		return nil, err
	}
	return func() { _ = os.Remove(file) }, nil
}
