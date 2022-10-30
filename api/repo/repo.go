package repo

type HelperRepo struct {
	RepoID   string `json:"id"`
	Name     string `json:"name"`
	NestedOwner struct{
		Owner string `json:"login"`
	} `json:"owner"` 
	Link     string `json:"html_url"`
}

type Release struct {
	CurrentReleaseTagName string `dynamodbav:"currentReleaseTagName,string" json:"tag_name"`
	CurrentReleaseID      string `dynamodbav:"currentReleaseID,string" json:"id"`
}

// TODO: get rid of NameHash, use RepoID instead
type Repo struct {
	RepoID   string `dynamobav:"repoID,string"`
	ChatID   string `dynamodbav:"chatID,string"`
	NameHash string `dynamodbav:"nameHash,string"`
	Name     string `dynamodbav:"repoName,string"`
	Owner    string `dynamodbav:"repoOwner,string"`
	Link     string `dynamodbav:"repoLink,string"`
	Release
}
