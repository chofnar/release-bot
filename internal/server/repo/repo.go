package repo

type Release struct {
	CurrentReleaseTagName string `dynamodbav:"currentReleaseTagName,string" json:"tag_name"`
	CurrentReleaseID      string `dynamodbav:"currentReleaseID,string" json:"id"`
	IsPrerelease          bool   `json:"isPrerelease"`
}

type Repo struct {
	RepoID                 string `dynamobav:"repoID,string" json:"repo_id,omitempty"`
	Name                   string `dynamodbav:"repoName,string" json:"name,omitempty"`
	Owner                  string `dynamodbav:"repoOwner,string" json:"owner,omitempty"`
	Link                   string `dynamodbav:"repoLink,string" json:"link,omitempty"`
	ShouldNotifyPrerelease bool   `dynamodbav:"shouldPre,bool" json:"shouldPre,omitempty"`
	Release
}

type RepoWithChatID struct {
	Repo
	ChatID string `dynamobav:"chatID,string"`
}
