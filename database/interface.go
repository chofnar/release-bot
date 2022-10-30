package database

import (
	"github.com/chofnar/release-bot/api/repo"
)

type Database interface {
	GetRepos(chatID string) ([]repo.Repo, error)
	AddRepo(chatID, details *repo.Repo) error
	RemoveRepo(chatID, nameHash string) error
	AllRepos() (*[]repo.HelperRepo, error)
	UpdateEntry(chatID, nameHash, newTagName string) error
}
