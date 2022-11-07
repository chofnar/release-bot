package messages

import (
	"fmt"

	"github.com/chofnar/release-bot/api/consts"
	"github.com/chofnar/release-bot/database"

	"github.com/mymmrac/telego"
	tu "github.com/mymmrac/telego/telegoutil"
)

func StartMessage(chatID int64) *telego.SendMessageParams {
	return tu.Message(tu.ID(chatID), consts.StartMessage).WithReplyMarkup(consts.StartKeyboard)
}

func EditedStartMessage(chatID int64, messageID int) *telego.EditMessageTextParams {
	return &telego.EditMessageTextParams{
		ChatID:      tu.ID(chatID),
		MessageID:   messageID,
		Text:        consts.StartMessage,
		ReplyMarkup: consts.StartKeyboard,
	}
}

func AboutMessage(chatID int64) *telego.SendMessageParams {
	return tu.Message(tu.ID(chatID), consts.AboutMessage)
}

func UnknownCommandMessage(chatID int64) *telego.SendMessageParams {
	return tu.Message(tu.ID(chatID), consts.UnknownCommandMessage)
}

func SeeAllReposMessage(chatID int64, messageID int) *telego.EditMessageTextParams {
	return &telego.EditMessageTextParams{
		ChatID:    tu.ID(chatID),
		MessageID: int(messageID),
		Text:      consts.ShowingAllReposMessage,
	}
}

func SeeAllReposMarkup(chatID int64, messageID int, database *database.Database) (*telego.EditMessageReplyMarkupParams, error) {
	repoList, err := (*database).GetRepos(fmt.Sprint(chatID))
	if err != nil {
		return nil, err
	}

	rows := make([][]telego.InlineKeyboardButton, len(repoList))

	for index, repo := range repoList {
		currentRow := make([]telego.InlineKeyboardButton, 3)

		repoNameButton := telego.InlineKeyboardButton{
			Text: repo.Name,
			URL:  repo.Link,
		}
		currentRow[0] = repoNameButton

		releaseButton := telego.InlineKeyboardButton{
			Text: repo.CurrentReleaseTagName,
			URL:  repo.Link + "/releases/" + repo.CurrentReleaseTagName,
		}
		currentRow[1] = releaseButton

		deleteButton := telego.InlineKeyboardButton{
			Text:         consts.DelteRepoEmoji,
			CallbackData: repo.RepoID,
		}
		currentRow[2] = deleteButton

		rows[index] = currentRow
	}

	replyMarkup := tu.InlineKeyboard(rows...)

	return &telego.EditMessageReplyMarkupParams{
		ChatID:      tu.ID(chatID),
		MessageID:   messageID,
		ReplyMarkup: replyMarkup,
	}, nil
}

func AddRepoMessage(chatID int64, messageID int) *telego.EditMessageTextParams {
	return &telego.EditMessageTextParams{
		ChatID:      tu.ID(chatID),
		MessageID:   messageID,
		Text:        consts.ShowingAddRepoMessage,
		ReplyMarkup: nil,
	}
}

func SuccesfullyAddedRepoMessage(chatID int64, messageID int) *telego.SendMessageParams {
	return &telego.SendMessageParams{
		ChatID:      tu.ID(chatID),
		Text:        consts.AddedRepoSuccesfully,
		ReplyMarkup: consts.AddAnotherRepoKeyboard,
	}
}

func DeleteRepo(chatID int64, repoID string, database *database.Database) error {
	err := (*database).RemoveRepo(fmt.Sprint(chatID), repoID)
	return err
}
