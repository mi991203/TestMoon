package logic

import (
	m "github.com/mi991203/TestMoon/libs/go/testing/domain"
)

type RunInput struct {
	OrgId          string        `json:"org_id"`
	ProjectId      string        `json:"project_id"`
	Id             string        `json:"id"`
	Items          []m.Selection `json:"items"`
	EnvironmentID  string        `json:"environment_id"`
	RunnerID       string        `json:"runner_id"`
	TimeoutSeconds int           `json:"timeout_seconds"`
	Debug          *m.Case       `json:"debug"`
	IdempotencyKey string        `json:"idempotency_key" validate:"required,max=128"`
}
