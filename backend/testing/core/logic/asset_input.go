package logic

// ResourceInput carries resource metadata and a definition into business transactions.
type ResourceInput struct {
	OrgId     string         `json:"org_id"`
	ProjectId string         `json:"project_id"`
	Id        string         `json:"id"`
	Kind      string         `json:"kind"`
	Name      string         `json:"name"`
	FolderID  string         `json:"folder_id"`
	Revision  int64          `json:"revision"`
	Document  map[string]any `json:"document"`
}
