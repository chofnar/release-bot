package behaviors

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/chofnar/release-bot/database"
	"github.com/chofnar/release-bot/errors"
	"github.com/chofnar/release-bot/server/messages"
	"github.com/chofnar/release-bot/server/repo"
	"github.com/hasura/go-graphql-client"
	"github.com/mymmrac/telego"
)

type BehaviorHandler struct {
	Bot                    *telego.Bot
	LinkRegex, DirectRegex *regexp.Regexp
	GQLClient              *graphql.Client
	DB                     database.Database
}

func (bh BehaviorHandler) About(chatID int64) error {
	_, err := bh.Bot.SendMessage(messages.StartMessage(chatID))
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
		repoToAdd, err := bh.validateAndBuildRepo(owner, repoName)
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
			_, err = bh.Bot.SendMessage(messages.SuccesfullyAddedRepoMessage(chatID))
			if err != nil {
				return err
			}
		} else {
			_, err = bh.Bot.SendMessage(messages.SuccesfullyAddedRepoWithoutReleasesMessage(chatID))
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

func (bh BehaviorHandler) validateAndBuildRepo(owner, name string) (repo.Repo, error) {
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
					TagName string
					ID      string
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

func (bh BehaviorHandler) SeeAll(chatID int64, messageID int) error {
	markup, err := messages.SeeAllReposMarkup(chatID, messageID, &bh.DB)
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

	return bh.SeeAll(chatID, messageID)
}

func (bh BehaviorHandler) UpdateRepos() error {
	//TODO: use bh.DB.AllRepos()

	return nil
}