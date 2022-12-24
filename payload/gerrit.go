package payload

type GerritEventType string

const (
	GerritEventChangeMerged GerritEventType = "change-merged"
)

type GerritChange struct {
	Project string `json:"project"`
	Branch  string `json:"branch"`
}

type GerritChangeKey struct {
	Key string `json:"key"`
}

type GerritPatchSet struct {
	Revision string `json:"revision"`
}

type GerritEvent struct {
	Change    *GerritChange    `json:"change"`
	ChangeKey *GerritChangeKey `json:"changeKey"`
	Type      GerritEventType  `json:"type"`
	PatchSet  *GerritPatchSet  `json:"patchSet"`
}
