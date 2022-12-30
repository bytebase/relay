package payload

type GerritEventType string

const (
	GerritEventChangeMerged GerritEventType = "change-merged"
)

type GerritChange struct {
	Project string `json:"project"`
	Branch  string `json:"branch"`
	ID      string `json:"id"`
}

type GerritPatchSet struct {
	Revision string `json:"revision"`
}

// GerritEvent is the API message for Gerrit webhook.
type GerritEvent struct {
	Change   *GerritChange   `json:"change"`
	Type     GerritEventType `json:"type"`
	PatchSet *GerritPatchSet `json:"patchSet"`
}

type GerritFileChangeMessage struct {
	Files []*GerritChangedFile
}

type GerritChangedFile struct {
	FileName string
	Content  string
}
