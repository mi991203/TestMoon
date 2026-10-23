// Package domain contains the version-one low-code document and worker IPC contracts.
package domain

type Variable struct {
	Value  string `json:"value"`
	Secret bool   `json:"secret"`
}
type Environment struct {
	Variables map[string]Variable `json:"variables"`
}
type Auth struct {
	Type     string `json:"type"`
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
	Token    string `json:"token,omitempty"`
	Key      string `json:"key,omitempty"`
	Value    string `json:"value,omitempty"`
	In       string `json:"in,omitempty"`
}
type HTTPRequest struct {
	Method    string            `json:"method"`
	URL       string            `json:"url"`
	Headers   map[string]string `json:"headers"`
	Query     map[string]string `json:"query"`
	BodyType  string            `json:"body_type"`
	Body      string            `json:"body"`
	Auth      Auth              `json:"auth"`
	TimeoutMS int               `json:"timeout_ms"`
}
type RequestPatch struct {
	Method   *string           `json:"method,omitempty"`
	URL      *string           `json:"url,omitempty"`
	Headers  map[string]string `json:"headers,omitempty"`
	Query    map[string]string `json:"query,omitempty"`
	BodyType *string           `json:"body_type,omitempty"`
	Body     *string           `json:"body,omitempty"`
	Auth     *Auth             `json:"auth,omitempty"`
}

func Merge(r HTTPRequest, p RequestPatch) HTTPRequest {
	if p.Method != nil {
		r.Method = *p.Method
	}
	if p.URL != nil {
		r.URL = *p.URL
	}
	if p.Body != nil {
		r.Body = *p.Body
	}
	if p.BodyType != nil {
		r.BodyType = *p.BodyType
	}
	if p.Auth != nil {
		r.Auth = *p.Auth
	}
	h := map[string]string{}
	for k, v := range r.Headers {
		h[k] = v
	}
	for k, v := range p.Headers {
		h[k] = v
	}
	r.Headers = h
	q := map[string]string{}
	for k, v := range r.Query {
		q[k] = v
	}
	for k, v := range p.Query {
		q[k] = v
	}
	r.Query = q
	return r
}

type API struct {
	Protocol      string      `json:"protocol"`
	Request       HTTPRequest `json:"request"`
	Description   string      `json:"description"`
	Schema        any         `json:"schema,omitempty"`
	EnvironmentID string      `json:"environment_id"`
	RunnerID      string      `json:"runner_id"`
}
type Operation struct {
	Type    string `json:"type"`
	Version int    `json:"version"`
	Script  string `json:"script"`
	Source  string `json:"source,omitempty"`
}
type Assertion struct {
	Source   string `json:"source"`
	Path     string `json:"path"`
	Operator string `json:"operator"`
	Expected any    `json:"expected"`
}
type Extraction struct {
	Source   string `json:"source"`
	Path     string `json:"path"`
	Variable string `json:"variable"`
	Secret   bool   `json:"secret"`
}
type Case struct {
	APIID               string            `json:"api_id"`
	Name                string            `json:"name"`
	Request             HTTPRequest       `json:"request"`
	Override            RequestPatch      `json:"override"`
	Variables           map[string]string `json:"variables"`
	CollectionVariables map[string]string `json:"collection_variables"`
	Before              []Operation       `json:"before"`
	After               []Operation       `json:"after"`
	Assertions          []Assertion       `json:"assertions"`
	Extractions         []Extraction      `json:"extractions"`
	EnvironmentID       string            `json:"environment_id"`
	RunnerID            string            `json:"runner_id"`
}
type Step struct {
	Kind              string            `json:"kind,omitempty"`
	Children          []Step            `json:"children,omitempty"`
	ID                string            `json:"id"`
	Name              string            `json:"name"`
	CaseID            string            `json:"case_id"`
	Inline            *Case             `json:"inline,omitempty"`
	Override          RequestPatch      `json:"override"`
	Variables         map[string]string `json:"variables"`
	ContinueOnFailure bool              `json:"continue_on_failure"`
}
type Scenario struct {
	Name                string            `json:"name"`
	Steps               []Step            `json:"steps"`
	CollectionVariables map[string]string `json:"collection_variables"`
}
type Selection struct {
	Kind     string      `json:"kind"`
	ID       string      `json:"id"`
	NodeID   string      `json:"node_id,omitempty"`
	Name     string      `json:"name,omitempty"`
	Children []Selection `json:"children,omitempty"`
}
type Task struct {
	Items          []Selection `json:"items"`
	EnvironmentID  string      `json:"environment_id"`
	RunnerID       string      `json:"runner_id"`
	TimeoutSeconds int         `json:"timeout_seconds"`
}
type CompiledStep struct {
	ID                string `json:"id"`
	Name              string `json:"name"`
	Case              Case   `json:"case"`
	ContinueOnFailure bool   `json:"continue_on_failure"`
}
type CompiledItem struct {
	ID                  string            `json:"id"`
	Name                string            `json:"name"`
	Steps               []CompiledStep    `json:"steps"`
	CollectionVariables map[string]string `json:"collection_variables"`
}
type Snapshot struct {
	Version        int              `json:"version"`
	OrgID          string           `json:"org_id"`
	ProjectID      string           `json:"project_id"`
	Environment    Environment      `json:"environment"`
	Items          []CompiledItem   `json:"items"`
	Revisions      map[string]int64 `json:"revisions"`
	TimeoutSeconds int              `json:"timeout_seconds"`
}
type Response struct {
	Status     int                 `json:"status"`
	Headers    map[string][]string `json:"headers"`
	Body       string              `json:"body"`
	DurationMS int64               `json:"duration_ms"`
	Error      string              `json:"error,omitempty"`
	Truncated  bool                `json:"truncated"`
}
type AssertionResult struct {
	Name    string `json:"name"`
	Passed  bool   `json:"passed"`
	Message string `json:"message,omitempty"`
}
type StepResult struct {
	ID         string            `json:"id"`
	Name       string            `json:"name"`
	Status     string            `json:"status"`
	Request    HTTPRequest       `json:"request"`
	Response   Response          `json:"response"`
	Assertions []AssertionResult `json:"assertions"`
	Logs       []string          `json:"logs"`
	Error      string            `json:"error,omitempty"`
	DurationMS int64             `json:"duration_ms"`
}
type ItemResult struct {
	ID     string       `json:"id"`
	Name   string       `json:"name"`
	Status string       `json:"status"`
	Steps  []StepResult `json:"steps"`
}
type Result struct {
	Status     string       `json:"status"`
	Items      []ItemResult `json:"items"`
	Error      string       `json:"error,omitempty"`
	DurationMS int64        `json:"duration_ms"`
}
type Frame struct {
	Type     string       `json:"type"`
	ID       string       `json:"id,omitempty"`
	ItemID   string       `json:"item_id,omitempty"`
	Request  *HTTPRequest `json:"request,omitempty"`
	Response *Response    `json:"response,omitempty"`
	Step     *StepResult  `json:"step,omitempty"`
	Result   *Result      `json:"result,omitempty"`
}
