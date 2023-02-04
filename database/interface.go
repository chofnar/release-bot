package database

import (
	"github.com/chofnar/release-bot/server/repo"
)

type Database interface {
	GetRepos(chatID string) ([]repo.Repo, error)
	AddRepo(chatID string, details *repo.Repo) error
	RemoveRepo(chatID, repoID string) error
	AllRepos() (*[]repo.Repo, error)
	UpdateEntry(chatID, repoID, newTagName string) error
	CheckExisting(chatID, repoID string) (bool, error)
}
