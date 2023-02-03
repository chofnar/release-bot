package errors

import "errors"

var (
	ErrInvalidDynamoDBEndpoint = errors.New("dynamodb: invalid endpoint")
	ErrChatIDNotFound          = errors.New("dynamodb: specified chatID does not exist in db")
	ErrNoReleases              = errors.New("repository has no release")
	ErrNoRepos                 = errors.New("no repos for current user")
)
