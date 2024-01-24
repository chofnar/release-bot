package messages

import (
	"fmt"
	"strconv"

	"github.com/chofnar/release-bot/internal/database"
	"github.com/chofnar/release-bot/internal/errors"
	"github.com/chofnar/release-bot/internal/server/consts"
	"github.com/chofnar/release-bot/internal/server/repo"
	"github.com/mymmrac/telego"
	tu "github.com/mymmrac/telego/telegoutil"
)

func StartMessage(chatID int64) *telego.SendMessageParams {
	return tu.Message(tu.ID(chatID), consts.StartMessage).WithReplyMarkup(consts.StartKeyboard)
}

func UpdateMessage(repository repo.RepoWithChatID, isPre bool) *telego.SendMessageParams {
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

	pre := ""
	if isPre {
		pre = "pre"
	}
	return tu.Message(tu.ID(int64(intID)), "New "+pre+"release: "+repository.Name+" : "+repository.CurrentReleaseTagName).WithReplyMarkup(&kbd)
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

func SeeReposMarkup(chatID int64, messageID, limit, page int, database *database.Database) (*telego.EditMessageReplyMarkupParams, error) {
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

	start := limit * page
	var end int
	if limit*page+limit < len(repoList) {
		end = limit
	} else {
		end = len(repoList) - limit*page
	}

	rows := make([][]telego.InlineKeyboardButton, end)

	for index, repo := range repoList[start : start+end] {
		currentRow := make([]telego.InlineKeyboardButton, 4)

		repoNameButton := telego.InlineKeyboardButton{
			Text: repo.Name,
			URL:  repo.Link,
		}
		currentRow[0] = repoNameButton

		var releaseButton telego.InlineKeyboardButton
		if repo.CurrentReleaseTagName != "" {
			releaseButton = telego.InlineKeyboardButton{
				Text: repo.CurrentReleaseTagName,
				URL:  repo.Link + "/releases/" + repo.CurrentReleaseTagName,
			}
		} else {
			releaseButton = telego.InlineKeyboardButton{
				Text: "N/A",
				URL:  repo.Link + "/releases",
			}
		}
		currentRow[1] = releaseButton

		notifyPre := consts.PreReleasesInactive
		newVal := "T"
		if repo.ShouldNotifyPrerelease {
			notifyPre = consts.PreReleasesActive
			newVal = "F"
		}
		preReleaseNotifyButton := telego.InlineKeyboardButton{
			Text:         notifyPre,
			CallbackData: consts.FlipOperationPrefix + newVal + "_" + repo.RepoID,
		}
		currentRow[2] = preReleaseNotifyButton

		deleteButton := telego.InlineKeyboardButton{
			Text:         consts.DelteRepoEmoji,
			CallbackData: repo.RepoID,
		}
		currentRow[3] = deleteButton

		rows[index] = currentRow
	}

	// prev/forward

	paginationRow := []telego.InlineKeyboardButton{}

	if page > 0 {
		paginationRow = append(paginationRow, telego.InlineKeyboardButton{
			Text:         "Previous",
			CallbackData: consts.PreviousOperationPrefix + strconv.Itoa(page-1),
		})
	}

	// more pages left
	if page*limit+limit < len(repoList) {
		paginationRow = append(paginationRow, telego.InlineKeyboardButton{
			Text:         "Next",
			CallbackData: consts.ForwardOperationPrefix + strconv.Itoa(page+1),
		})
	}

	rows = append(rows, paginationRow)

	// back

	backToMenuRow := []telego.InlineKeyboardButton{}
	backToMenuRow = append(backToMenuRow, telego.InlineKeyboardButton{
		Text:         "Back to Menu",
		CallbackData: consts.MenuCallback,
	})

	rows = append(rows, backToMenuRow)

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

func SuccessfullyAddedRepoMessage(chatID int64) *telego.SendMessageParams {
	return &telego.SendMessageParams{
		ChatID:      tu.ID(chatID),
		Text:        consts.AddedRepoSuccessfully,
		ReplyMarkup: consts.AddAnotherRepoKeyboard,
	}
}

func SuccessfullyAddedRepoWithoutReleasesMessage(chatID int64) *telego.SendMessageParams {
	return &telego.SendMessageParams{
		ChatID:      tu.ID(chatID),
		Text:        consts.AddedRepoSuccessfullyNoReleases,
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

func SetPreReleaseRetrieve(chatID int64, repoID string, newValue bool, database *database.Database) error {
	err := (*database).SetPreReleaseRetrieve(fmt.Sprint(chatID), repoID, newValue)
	return err
}
