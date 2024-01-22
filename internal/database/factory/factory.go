package factory

import (
	"github.com/chofnar/release-bot/internal/database"
	"github.com/chofnar/release-bot/internal/database/dynamodb"
	"go.uber.org/zap"
)

var driverFactories = map[string]DriverFactory{
	"dynamodb": &dynamodb.DriverFactory{},
}

type DriverFactory interface {
	Create(logger zap.SugaredLogger) database.Database
}

func Create(dbtype string, logger zap.SugaredLogger) database.Database {
	driverFactory, ok := driverFactories[dbtype]
	if !ok {
		return nil
	}

	return driverFactory.Create(logger)
}
