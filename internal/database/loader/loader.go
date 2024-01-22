package loader

import (
	"github.com/chofnar/release-bot/internal/database"
	"github.com/chofnar/release-bot/internal/database/factory"
	"go.uber.org/zap"
)

func GetDatabase(logger zap.SugaredLogger) database.Database {
	return factory.Create("dynamodb", logger)
}
