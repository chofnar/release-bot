package messages

import (
	"github.com/chofnar/release-bot/api/consts"

	"github.com/mymmrac/telego"
	tu "github.com/mymmrac/telego/telegoutil"
)

func StartMessage(chatID int64) *telego.SendMessageParams {
	return tu.Message(tu.ID(chatID), consts.StartMessage).WithReplyMarkup(consts.StartKeyboard)
}

func AboutMessage(chatID int64) *telego.SendMessageParams {
	return tu.Message(tu.ID(chatID), consts.AboutMessage)
}

func UnknownCommandMessage(chatID int64) *telego.SendMessageParams {
	return tu.Message(tu.ID(chatID), consts.UnknownCommandMessage)
}

func SeeAllReposMessage(chatID int64, messageID int) *telego.EditMessageTextParams {
	return &telego.EditMessageTextParams{
		ChatID: tu.ID(chatID),
		MessageID: int(messageID),
		Text: consts.ShowingAllReposMessage,
	}
}

func SeeAllReposMarkup(chatID int64, messageID int) *telego.EditMessageReplyMarkupParams {
	return &telego.EditMessageReplyMarkupParams{

	}
}

func AddRepoMessage(chatID int64, messageID int) *telego.EditMessageTextParams {
	return &telego.EditMessageTextParams{
		ChatID: tu.ID(chatID),
		MessageID: int(messageID),
		Text: consts.ShowingAddRepoMessage,
	}
}

func AddRepoMarkup(chatID int64, messageID int) *telego.EditMessageReplyMarkupParams {
	return &telego.EditMessageReplyMarkupParams{
		
	}
}
