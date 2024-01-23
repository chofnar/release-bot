package behaviors

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/chofnar/release-bot/internal/database"
	"github.com/chofnar/release-bot/internal/errors"
	"github.com/chofnar/release-bot/internal/server/consts"
	"github.com/chofnar/release-bot/internal/server/messages"
	"github.com/chofnar/release-bot/internal/server/repo"
	"github.com/hasura/go-graphql-client"
	"github.com/mymmrac/telego"
	"go.uber.org/zap"
)

type BehaviorHandler struct {
	Bot                    *telego.Bot
	LinkRegex, DirectRegex *regexp.Regexp
	GQLClient              *graphql.Client
	DB                     database.Database
}

func (bh BehaviorHandler) About(chatID int64) error {
	_, err := bh.Bot.SendMessage(messages.AboutMessage(chatID))
	return err
}

func (bh BehaviorHandler) Start(chatID int64) error {
	_, err := bh.Bot.SendMessage(messages.StartMessage(chatID))
	return err
}

func (bh BehaviorHandler) UnknownCommand(chatID int64) error {
	_, err := bh.Bot.SendMessage(messages.UnknownCommandMessage(chatID))
	if err != nil {
		return err
	}

	_, err = bh.Bot.SendMessage(messages.StartMessage(chatID))
	return err
}

func (bh BehaviorHandler) SentRepo(messageText string, messageID int, chatID int64) error {
	owner, repoName, valid := bh.validateInput(messageText)
	if valid {
		repoToAdd, err := bh.validateAndRetrieveRepo(owner, repoName)
		hasReleases := true
		if err != nil {
			if err == errors.ErrNoReleases {
				hasReleases = false
			} else {
				_, _ = bh.Bot.SendMessage(messages.RepoNotFoundMessage(chatID))

				return err
			}
		}
		exists, err := bh.DB.CheckExisting(fmt.Sprint(chatID), repoToAdd.RepoID)
		if err != nil {
			return err
		}

		if exists {
			_, err = bh.Bot.SendMessage(messages.AlreadyExistsMessage(chatID, messageID))
			if err != nil {
				return err
			}
			return nil
		}

		err = bh.DB.AddRepo(fmt.Sprint(chatID), &repoToAdd)
		if err != nil {
			return err
		}

		if hasReleases {
			_, err = bh.Bot.SendMessage(messages.SuccessfullyAddedRepoMessage(chatID))
			if err != nil {
				return err
			}
		} else {
			_, err = bh.Bot.SendMessage(messages.SuccessfullyAddedRepoWithoutReleasesMessage(chatID))
			if err != nil {
				return err
			}
		}
	} else {
		_, err := bh.Bot.SendMessage(messages.InvalidRepoMessage(chatID))
		if err != nil {
			return err
		}
	}
	return nil
}

func (bh BehaviorHandler) validateInput(message string) (owner, repo string, isValid bool) {
	if result := bh.LinkRegex.FindStringSubmatch(message); result != nil {
		owner, repo = result[1], result[2]
		isValid = true
		return
	}

	if result := bh.DirectRegex.FindString(message); result != "" {
		splitResult := strings.Split(message, "/")
		owner, repo = splitResult[0], splitResult[1]
		isValid = true
		return
	}

	return
}

func (bh BehaviorHandler) validateAndRetrieveRepo(owner, name string) (repo.Repo, error) {
	variables := map[string]interface{}{
		"name":  name,
		"owner": owner,
	}

	var getRepoQuery struct {
		Repository struct {
			ID    string
			URL   string
			Name  string
			Owner struct {
				Login string
			}
			Releases struct {
				Nodes []struct {
					TagName      string
					ID           string
					IsPrerelease bool
				}
			} `graphql:"releases(first: 1)"`
		} `graphql:"repository(name: $name, owner: $owner)"`
	}

	err := bh.GQLClient.Query(context.TODO(), &getRepoQuery, variables)
	if err != nil {
		return repo.Repo{}, err
	}

	if len(getRepoQuery.Repository.Releases.Nodes) != 0 {
		return repo.Repo{
			RepoID: getRepoQuery.Repository.ID,
			Name:   getRepoQuery.Repository.Name,
			Owner:  getRepoQuery.Repository.Owner.Login,
			Link:   getRepoQuery.Repository.URL,
			Release: repo.Release{
				CurrentReleaseTagName: getRepoQuery.Repository.Releases.Nodes[0].TagName,
				CurrentReleaseID:      getRepoQuery.Repository.Releases.Nodes[0].ID,
				IsPrerelease:          getRepoQuery.Repository.Releases.Nodes[0].IsPrerelease,
			},
		}, nil
	}

	return repo.Repo{
		RepoID: getRepoQuery.Repository.ID,
		Name:   getRepoQuery.Repository.Name,
		Owner:  getRepoQuery.Repository.Owner.Login,
		Link:   getRepoQuery.Repository.URL,
		Release: repo.Release{
			CurrentReleaseTagName: "",
			CurrentReleaseID:      "",
		},
	}, errors.ErrNoReleases
}

func (bh BehaviorHandler) SeeRepos(chatID int64, messageID, limit, page int) error {
	markup, err := messages.SeeReposMarkup(chatID, messageID, limit, page, &bh.DB)
	if err != errors.ErrNoRepos && err != nil {
		return err
	}
	if err == errors.ErrNoRepos {
		_, err = bh.Bot.EditMessageText(messages.SeeAllReposButNoneFoundMessage(chatID, messageID, *markup.ReplyMarkup))
		if err != nil {
			return err
		}
	} else {
		_, err = bh.Bot.EditMessageText(messages.SeeAllReposMessage(chatID, messageID, *markup.ReplyMarkup))
		if err != nil {
			return err
		}
	}
	return nil
}

func (bh BehaviorHandler) Add(chatID int64, messageID int) error {
	_, err := bh.Bot.EditMessageText(messages.AddRepoMessage(chatID, messageID))
	return err
}

func (bh BehaviorHandler) Menu(chatID int64, messageID int) error {
	_, err := bh.Bot.EditMessageText(messages.EditedStartMessage(chatID, messageID))
	return err
}

func (bh BehaviorHandler) DeleteRepo(chatID int64, messageID int, data string) error {
	err := messages.DeleteRepo(chatID, data, &bh.DB)
	if err != nil {
		return err
	}

	return bh.Menu(chatID, messageID)
}

func (bh BehaviorHandler) FlipPreRelease(chatID int64, messageID int, repoIDwithOP string) error {
	repoIDwithNewVal := strings.TrimPrefix(repoIDwithOP, consts.FlipOperationPrefix)

	newValStr := repoIDwithNewVal[0]
	newVal := false
	if newValStr == byte('T') {
		newVal = true
	}

	repoID := repoIDwithNewVal[2:]

	err := messages.SetPreReleaseRetrieve(chatID, repoID, newVal, &bh.DB)
	if err != nil {
		return err
	}

	return bh.Menu(chatID, messageID)
}

func (bh BehaviorHandler) newUpdate(repository repo.RepoWithChatID, isPre bool) error {
	_, err := bh.Bot.SendMessage(messages.UpdateMessage(repository, isPre))
	return err
}

type erroredRepo struct {
	Err  error     `json:"err,omitempty"`
	Repo repo.Repo `json:"repo,omitempty"`
}

func (bh BehaviorHandler) UpdateRepos(logger zap.SugaredLogger) []erroredRepo {
	repos, err := bh.DB.AllRepos()
	failedRepos := []erroredRepo{}
	if err != nil {
		failedRepos = append(failedRepos, erroredRepo{Err: err})
		return failedRepos
	}

	for _, repository := range repos {
		newlyRetrievedRepo, err := bh.validateAndRetrieveRepo(repository.Owner, repository.Name)
		if err != nil {
			failedRepos = append(failedRepos, erroredRepo{Err: err, Repo: newlyRetrievedRepo})
			// Could not resolve
			if strings.Contains(err.Error(), "Could not resolve to a Repository with the name") {
				errdb := bh.DB.RemoveRepo(repository.ChatID, repository.RepoID)
				if errdb != nil {
					logger.Error(errdb)
				}
			}
			continue
		}

		if newlyRetrievedRepo.CurrentReleaseID != repository.CurrentReleaseID && (!newlyRetrievedRepo.IsPrerelease || (newlyRetrievedRepo.IsPrerelease && repository.ShouldNotifyPrerelease)) {
			withChatID := repo.RepoWithChatID{
				Repo:   newlyRetrievedRepo,
				ChatID: repository.ChatID,
			}

			err = bh.DB.UpdateEntry(withChatID)
			if err != nil {
				failedRepos = append(failedRepos, erroredRepo{Err: err, Repo: newlyRetrievedRepo})
				continue
			}

			err = bh.newUpdate(withChatID, newlyRetrievedRepo.IsPrerelease)
			if err != nil {
				// clean up orphaned repos:
				// 400 chat not found, 403 user blocked the bot
				if strings.Contains(err.Error(), "Forbidden: bot was blocked by the user") || strings.Contains(err.Error(), "Bad Request: chat not found") {
					errdb := bh.DB.RemoveRepo(repository.ChatID, repository.RepoID)
					if errdb != nil {
						logger.Error(errdb)
					}
					continue
				}

				// other
				failedRepos = append(failedRepos, erroredRepo{Err: err, Repo: newlyRetrievedRepo})
				continue
			}

		}
	}

	return failedRepos
}
