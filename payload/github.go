package payload

type GitHub struct {
	Ref     string `json:"ref"`
	Compare string `json:"compare"`
}
