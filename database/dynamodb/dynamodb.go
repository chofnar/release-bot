package dynamodb

import (
	"context"
	"encoding/json"
	"net/http"
	"os"

	"github.com/chofnar/release-bot/api/repo"
	"github.com/chofnar/release-bot/database"
	"github.com/chofnar/release-bot/errors"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
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

	if value := os.Getenv("BOT_ENDPOINT"); value != "" {
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

func (db *Driver) AddRepo(chatID, details *repo.Repo) error {
	resp, err := http.Get("https://api.github.com/repos/" + details.Owner + "/" + details.Name + "/releases")
	if err != nil {
		return err
	}
	resultingRepo := repo.HelperRepo{}

	err = json.NewDecoder(resp.Body).Decode(&resultingRepo)
	if err != nil {
		return err
	}

	return nil
}

func (db *Driver) RemoveRepo(chatID, nameHash string) error {
	_, err := db.client.DeleteItem(context.TODO(), &dynamodb.DeleteItemInput{
		Key: map[string]types.AttributeValue{
			"chatID": &types.AttributeValueMemberS{
				Value: chatID,
			},
			"nameHash": &types.AttributeValueMemberS{
				Value: nameHash,
			},
		},
		TableName: &db.tableName,
	})

	return err
}

func (db *Driver) AllRepos() (*[]repo.HelperRepo, error) {
	return nil, nil
}

func (db *Driver) UpdateEntry(chatID, nameHash, newTagName string) error {
	return nil
}
