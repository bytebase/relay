package payload

type GitHubPushEvent struct {
	Ref     string `json:"ref"`
	Compare string `json:"compare"`
}
