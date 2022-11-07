package repo

type Release struct {
	CurrentReleaseTagName string `dynamodbav:"currentReleaseTagName,string" json:"tag_name"`
	CurrentReleaseID      string `dynamodbav:"currentReleaseID,string" json:"id"`
}

type Repo struct {
	RepoID string `dynamobav:"repoID,string"`
	Name   string `dynamodbav:"repoName,string"`
	Owner  string `dynamodbav:"repoOwner,string"`
	Link   string `dynamodbav:"repoLink,string"`
	Release
}
