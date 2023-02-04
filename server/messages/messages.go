package messages

import (
	"fmt"
	"strconv"

	"github.com/chofnar/release-bot/database"
	"github.com/chofnar/release-bot/errors"
	"github.com/chofnar/release-bot/server/consts"
	"github.com/chofnar/release-bot/server/repo"
	"github.com/mymmrac/telego"
	tu "github.com/mymmrac/telego/telegoutil"
)

func StartMessage(chatID int64) *telego.SendMessageParams {
	return tu.Message(tu.ID(chatID), consts.StartMessage).WithReplyMarkup(consts.StartKeyboard)
}

func UpdateMessage(repository repo.RepoWithChatID) *telego.SendMessageParams {
	currentRow := make([]telego.InlineKeyboardButton, 1)

	kbd := telego.InlineKeyboardMarkup{
		InlineKeyboard: [][]telego.InlineKeyboardButton{
			{
				telego.InlineKeyboardButton{
					Text: consts.CheckRepo,
					URL:  repository.Link + "/releases/" + repository.CurrentReleaseTagName,
				},
			},
		},
	}
	releaseButton := telego.InlineKeyboardButton{
		Text: consts.CheckRepo,
		URL:  repository.Link + "/releases/" + repository.CurrentReleaseTagName,
	}
	currentRow[0] = releaseButton

	intID, _ := strconv.Atoi(repository.ChatID)

	return tu.Message(tu.ID(int64(intID)), "New release:"+repository.Name+" : "+repository.CurrentReleaseTagName).WithReplyMarkup(&kbd)
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

func SeeAllReposMessage(chatID int64, messageID int, markup telego.InlineKeyboardMarkup) *telego.EditMessageTextParams {
	return &telego.EditMessageTextParams{
		ChatID:      tu.ID(chatID),
		MessageID:   int(messageID),
		Text:        consts.ShowingAllReposMessage,
		ReplyMarkup: &markup,
	}
}

func SeeAllReposButNoneFoundMessage(chatID int64, messageID int, markup telego.InlineKeyboardMarkup) *telego.EditMessageTextParams {
	return &telego.EditMessageTextParams{
		ChatID:      tu.ID(chatID),
		MessageID:   int(messageID),
		Text:        consts.ShowingAllReposButNoneFoundMessage,
		ReplyMarkup: &markup,
	}
}

func SeeAllReposMarkup(chatID int64, messageID int, database *database.Database) (*telego.EditMessageReplyMarkupParams, error) {
	repoList, err := (*database).GetRepos(fmt.Sprint(chatID))
	if err != nil {
		return nil, err
	}

	if len(repoList) == 0 {

		return &telego.EditMessageReplyMarkupParams{
			ChatID:      tu.ID(chatID),
			MessageID:   messageID,
			ReplyMarkup: consts.AddAnotherRepoKeyboard,
		}, errors.ErrNoRepos
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
		ReplyMarkup: consts.CancelAddKeyboard,
	}
}

func InvalidRepoMessage(chatID int64) *telego.SendMessageParams {
	return &telego.SendMessageParams{
		ChatID:      tu.ID(chatID),
		Text:        consts.InvalidRepoMessage,
		ReplyMarkup: nil,
	}
}

func SuccesfullyAddedRepoMessage(chatID int64) *telego.SendMessageParams {
	return &telego.SendMessageParams{
		ChatID:      tu.ID(chatID),
		Text:        consts.AddedRepoSuccesfully,
		ReplyMarkup: consts.AddAnotherRepoKeyboard,
	}
}

func SuccesfullyAddedRepoWithoutReleasesMessage(chatID int64) *telego.SendMessageParams {
	return &telego.SendMessageParams{
		ChatID:      tu.ID(chatID),
		Text:        consts.AddedRepoSuccesfullyNoReleases,
		ReplyMarkup: consts.AddAnotherRepoKeyboard,
	}
}

func AlreadyExistsMessage(chatID int64, messageID int) *telego.SendMessageParams {
	return &telego.SendMessageParams{
		ChatID:      tu.ID(chatID),
		Text:        consts.RepoExists,
		ReplyMarkup: consts.AddAnotherRepoKeyboard,
	}
}

func RepoNotFoundMessage(chatID int64) *telego.SendMessageParams {
	return &telego.SendMessageParams{
		ChatID:      tu.ID(chatID),
		Text:        consts.RepoNotFound,
		ReplyMarkup: consts.AddAnotherRepoKeyboard,
	}
}

func DeleteRepo(chatID int64, repoID string, database *database.Database) error {
	err := (*database).RemoveRepo(fmt.Sprint(chatID), repoID)
	return err
}
