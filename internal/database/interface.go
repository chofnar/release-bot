package database

import (
	"github.com/chofnar/release-bot/internal/server/repo"
)

type Database interface {
	GetRepos(chatID string) ([]repo.Repo, error)
	AddRepo(chatID string, details *repo.Repo) error
	RemoveRepo(chatID, repoID string) error
	SetPreReleaseRetrieve(chatID, repoID string, newValue bool) error
	AllRepos() ([]repo.RepoWithChatID, error)
	UpdateEntry(repo repo.RepoWithChatID) error
	CheckExisting(chatID, repoID string) (bool, error)
}
