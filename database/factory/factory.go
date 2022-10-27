package factory

import (
	"github.com/chofnar/release-bot/database"
	"github.com/chofnar/release-bot/database/dynamodb"
	"go.uber.org/zap"
)

var driverFactories = map[string]DriverFactory{
	"dynamodb": &dynamodb.DriverFactory{},
}

type DriverFactory interface {
	Create(parameters interface{}, logger zap.SugaredLogger) (database.Database)
}

func Create(dbtype string, parameters interface{}, logger zap.SugaredLogger) (database.Database) {
	driverFactory, ok := driverFactories[dbtype]
	if !ok {
		return nil
	}

	return driverFactory.Create(parameters, logger)
}
