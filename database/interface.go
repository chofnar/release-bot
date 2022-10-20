package database

import (
	"github.com/chofnar/release-bot/api/repo"
)

type database interface {
	GetRepos(chatID string) error
	AddRepo(chatID, repoPath string) error
	RemoveRepo(chatID, nameHash string) error
	AllRepos() (*[]repo.Repo, error)
	UpdateEntry(chatID, nameHash, newTagName string) error
}
