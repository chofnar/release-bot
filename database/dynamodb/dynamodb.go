package dynamodb

import (
	"github.com/chofnar/release-bot/api/repo"
	"github.com/chofnar/release-bot/database"
	"go.uber.org/zap"
)

type Dynamodb struct {}

type DriverFactory struct{}

func (factory *DriverFactory) Create(parameters interface{}, logger zap.SugaredLogger) (database.Database) {
	return &Dynamodb{
		// TODO: implement
	}
}

func (db *Dynamodb) GetRepos(chatID string) error {
	return nil
}

func (db *Dynamodb) AddRepo(chatID, repoPath string) error {
	return nil
}

func (db *Dynamodb) RemoveRepo(chatID, nameHash string) error {
	return nil
}

func (db *Dynamodb) AllRepos() (*[]repo.Repo, error) {
	return nil, nil
}

func (db *Dynamodb) UpdateEntry(chatID, nameHash, newTagName string) error {
	return nil
}
