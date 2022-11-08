package errors

import "errors"

var (
	ErrInvalidDynamoDBEndpoint = errors.New("dynamodb: invalid endpoint")
	ErrChatIDNotFound          = errors.New("dynamodb: specified chatID does not exist in db")
	ErrNoReleases              = errors.New("Repository has no release!")
)
