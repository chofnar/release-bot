package database

import (
	"github.com/chofnar/release-bot/api/repo"
)

type Database interface {
	GetRepos(chatID string) ([]repo.Repo, error)
	AddRepo(chatID string, details *repo.Repo) error
	RemoveRepo(chatID, repoID string) error
	AllRepos() (*[]repo.Repo, error)
	UpdateEntry(chatID, repoID, newTagName string) error
}
