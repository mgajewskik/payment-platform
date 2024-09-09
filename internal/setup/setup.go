package setup

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"

	"github.com/mgajewskik/payment-platform/internal/storage"
)

type DBSetup struct {
	client    *dynamodb.Client
	tableName string
}

func NewDBSetup(config aws.Config, tableName string) *DBSetup {
	return &DBSetup{
		client:    dynamodb.NewFromConfig(config),
		tableName: tableName,
	}
}

func (s *DBSetup) Setup() error {
	exists, err := s.TableExists()
	if err != nil {
		return err
	}

	if !exists {
		err := s.CreateTable()
		if err != nil {
			return err
		}

		time.Sleep(10 * time.Second) // wait for table to be created
	}

	err = s.InsertTestData()
	if err != nil {
		return err
	}

	return nil
}

func (s *DBSetup) TableExists() (bool, error) {
	_, err := s.client.DescribeTable(context.TODO(), &dynamodb.DescribeTableInput{
		TableName: aws.String(s.tableName),
	})
	if err != nil {
		var notFound *types.ResourceNotFoundException
		if errors.As(err, &notFound) {
			return false, nil
		}

		return false, err
	}

	return true, nil
}

func (s *DBSetup) CreateTable() error {
	input := &dynamodb.CreateTableInput{
		AttributeDefinitions: []types.AttributeDefinition{
			{
				AttributeName: aws.String("PK"),
				AttributeType: types.ScalarAttributeTypeS,
			},
			{
				AttributeName: aws.String("SK"),
				AttributeType: types.ScalarAttributeTypeS,
			},
		},
		KeySchema: []types.KeySchemaElement{
			{
				AttributeName: aws.String("PK"),
				KeyType:       types.KeyTypeHash,
			},
			{
				AttributeName: aws.String("SK"),
				KeyType:       types.KeyTypeRange,
			},
		},
		TableName:   aws.String(s.tableName),
		BillingMode: types.BillingModePayPerRequest,
		SSESpecification: &types.SSESpecification{
			Enabled: aws.Bool(true),
		},
	}

	_, err := s.client.CreateTable(context.TODO(), input)
	if err != nil {
		log.Fatalf("Got error calling CreateTable: %s", err)
	}

	return nil
}

func (s *DBSetup) InsertTestData() error {
	item := storage.MerchantItem{
		PK: "test@merchant",
		SK: "MERCHANT",
		AccountDetails: storage.AccountDetails{
			Name:     "Test Merchant",
			IBAN:     "DE89370400440532013000",
			BIC:      "COBADEFFXXX",
			Currency: "EUR",
		},
	}

	av, err := attributevalue.MarshalMap(item)
	if err != nil {
		return err
	}

	input := &dynamodb.PutItemInput{
		Item:      av,
		TableName: aws.String(s.tableName),
	}

	_, err = s.client.PutItem(context.TODO(), input)
	if err != nil {
		return err
	}

	return nil
}

func (s *DBSetup) Teardown() error {
	_, err := s.client.DeleteTable(context.TODO(), &dynamodb.DeleteTableInput{
		TableName: aws.String(s.tableName),
	})
	if err != nil {
		return err
	}

	return nil
}
