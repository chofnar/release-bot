package dynamodb

import (
	"context"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/chofnar/release-bot/internal/database"
	"github.com/chofnar/release-bot/internal/errors"
	"github.com/chofnar/release-bot/internal/server/repo"
	"go.uber.org/zap"
)

type Driver struct {
	client    *dynamodb.Client
	logger    zap.SugaredLogger
	tableName string
}

type DriverFactory struct{}

type dynamoDBparams struct {
	endpoint  string
	region    string
	tableName string
}

const (
	defaultEndpoint  = "http://localhost:4566"
	defaultRegion    = "eu-central-1"
	defaultTableName = "ReleasesBot"
)

func (params *dynamoDBparams) fillDefaults() {
	params.endpoint = defaultEndpoint
	params.region = defaultRegion
	params.tableName = defaultTableName
}

func loadConfig() dynamoDBparams {
	var params dynamoDBparams
	params.fillDefaults()

	if value := os.Getenv("BOT_DYNAMODB_ENDPOINT"); value != "" {
		params.endpoint = value
	}
	if value := os.Getenv("BOT_REGION"); value != "" {
		params.region = value
	}
	if value := os.Getenv("BOT_TABLE_NAME"); value != "" {
		params.tableName = value
	}

	return params
}

func (factory *DriverFactory) Create(logger zap.SugaredLogger) database.Database {
	params := loadConfig()

	customResolver := aws.EndpointResolverWithOptionsFunc(
		func(service, region string, options ...interface{}) (aws.Endpoint, error) {
			if service == dynamodb.ServiceID {
				return aws.Endpoint{
					PartitionID:   "aws",
					URL:           params.endpoint,
					SigningRegion: params.region,
				}, nil
			}
			return aws.Endpoint{}, errors.ErrInvalidDynamoDBEndpoint
		})

	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(params.region), config.WithEndpointResolverWithOptions(customResolver))
	if err != nil {
		logger.Error(err)
	}

	return &Driver{
		client:    dynamodb.NewFromConfig(cfg),
		tableName: params.tableName,
		logger:    logger,
	}
}

func (db *Driver) GetRepos(chatID string) ([]repo.Repo, error) {
	filterExp := "chatID = :chatid"
	filterField := types.AttributeValueMemberS{Value: chatID}

	resp, err := db.client.Query(context.TODO(), &dynamodb.QueryInput{
		TableName:              &db.tableName,
		KeyConditionExpression: &filterExp,
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":chatid": &filterField,
		},
	})
	if err != nil {
		return nil, err
	}

	if resp == nil {
		return nil, errors.ErrChatIDNotFound
	}

	repos := make([]repo.Repo, len(resp.Items))

	err = attributevalue.UnmarshalListOfMaps(resp.Items, &repos)
	if err != nil {
		return nil, err
	}

	return repos, nil
}

func (db *Driver) AddRepo(chatID string, details *repo.Repo) error {
	// TODO: Contexts, mate
	_, err := db.client.PutItem(context.TODO(), &dynamodb.PutItemInput{
		TableName: &db.tableName,
		Item: map[string]types.AttributeValue{
			"chatID":                &types.AttributeValueMemberS{Value: chatID},
			"repoID":                &types.AttributeValueMemberS{Value: details.RepoID},
			"repoName":              &types.AttributeValueMemberS{Value: details.Name},
			"repoOwner":             &types.AttributeValueMemberS{Value: details.Owner},
			"repoLink":              &types.AttributeValueMemberS{Value: details.Link},
			"currentReleaseTagName": &types.AttributeValueMemberS{Value: details.Release.CurrentReleaseTagName},
			"currentReleaseID":      &types.AttributeValueMemberS{Value: details.Release.CurrentReleaseID},
			"shouldPre":             &types.AttributeValueMemberBOOL{Value: false},
		},
	})
	if err != nil {
		return err
	}
	return nil
}

func (db *Driver) RemoveRepo(chatID, repoID string) error {
	_, err := db.client.DeleteItem(context.TODO(), &dynamodb.DeleteItemInput{
		Key: map[string]types.AttributeValue{
			"chatID": &types.AttributeValueMemberS{
				Value: chatID,
			},
			"repoID": &types.AttributeValueMemberS{
				Value: repoID,
			},
		},
		TableName: &db.tableName,
	})

	return err
}

func (db *Driver) AllRepos() ([]repo.RepoWithChatID, error) {
	// TODO: may need to implement pagination
	result, err := db.client.Scan(context.TODO(), &dynamodb.ScanInput{
		TableName: &db.tableName,
	})

	if err != nil {
		return nil, err
	}

	repos := make([]repo.RepoWithChatID, len(result.Items))

	err = attributevalue.UnmarshalListOfMaps(result.Items, &repos)
	if err != nil {
		return nil, err
	}

	return repos, nil
}

func (db *Driver) UpdateEntry(repo repo.RepoWithChatID) error {
	_, err := db.client.UpdateItem(context.TODO(), &dynamodb.UpdateItemInput{
		Key: map[string]types.AttributeValue{
			"chatID": &types.AttributeValueMemberS{Value: fmt.Sprint(repo.ChatID)},
			"repoID": &types.AttributeValueMemberS{Value: repo.RepoID},
		},
		UpdateExpression: aws.String("set currentReleaseID = :releaseID, currentReleaseTagName = :releaseTagName"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":releaseID":      &types.AttributeValueMemberS{Value: repo.CurrentReleaseID},
			":releaseTagName": &types.AttributeValueMemberS{Value: repo.CurrentReleaseTagName},
		},
		TableName: &db.tableName,
	})

	return err
}

func (db *Driver) SetPreReleaseRetrieve(chatID, repoID string, newValue bool) error {
	_, err := db.client.UpdateItem(context.TODO(), &dynamodb.UpdateItemInput{
		Key: map[string]types.AttributeValue{
			"chatID": &types.AttributeValueMemberS{Value: fmt.Sprint(chatID)},
			"repoID": &types.AttributeValueMemberS{Value: repoID},
		},
		UpdateExpression: aws.String("set shouldPre = :newValuePreReleaseRetrieve"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":newValuePreReleaseRetrieve": &types.AttributeValueMemberBOOL{Value: newValue},
		},
		TableName: &db.tableName,
	})

	return err
}

func (db *Driver) CheckExisting(chatID, repoID string) (bool, error) {
	output, err := db.client.GetItem(context.TODO(), &dynamodb.GetItemInput{
		TableName: &db.tableName,
		Key: map[string]types.AttributeValue{
			"chatID": &types.AttributeValueMemberS{Value: chatID},
			"repoID": &types.AttributeValueMemberS{Value: repoID},
		},
	})
	if err != nil {
		return false, err
	}

	if output.Item != nil {
		return true, nil
	}

	return false, nil
}
