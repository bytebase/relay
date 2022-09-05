package payload

type GitHubAuthor struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

type GitHubCommit struct {
	Message   string       `json:"message"`
	Timestamp string       `json:"timestamp"`
	Author    GitHubAuthor `json:"author"`
}

type GitHubPushEvent struct {
	Ref        string       `json:"ref"`
	Compare    string       `json:"compare"`
	HeadCommit GitHubCommit `json:"head_commit"`
}
