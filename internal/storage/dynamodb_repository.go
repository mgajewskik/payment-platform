package storage

import (
	"context"
	"log/slog"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"

	"github.com/mgajewskik/payment-platform/internal/domain/entities"
)

type DBRepository interface {
	GetMerchantDetails(merchantID string) (entities.Merchant, error)
	CreateNewPayment(payment entities.Payment) error
	UpdatePayment(payment entities.Payment) error
	GetPayment(merchantID, paymentID string) (entities.Payment, error)
}

type DynamoDBClient interface {
	PutItem(
		ctx context.Context,
		params *dynamodb.PutItemInput,
		optFns ...func(*dynamodb.Options),
	) (*dynamodb.PutItemOutput, error)
	UpdateItem(
		ctx context.Context,
		params *dynamodb.UpdateItemInput,
		optFns ...func(*dynamodb.Options),
	) (*dynamodb.UpdateItemOutput, error)
	GetItem(
		ctx context.Context,
		params *dynamodb.GetItemInput,
		optFns ...func(*dynamodb.Options),
	) (*dynamodb.GetItemOutput, error)
}

type DynamoDBRepository struct {
	db        DynamoDBClient
	tableName string
	logger    *slog.Logger
}

func NewDynamoDBRepository(
	tableName string,
	config aws.Config,
	logger *slog.Logger,
) *DynamoDBRepository {
	return &DynamoDBRepository{
		db:        dynamodb.NewFromConfig(config),
		tableName: tableName,
		logger:    logger,
	}
}

func (r *DynamoDBRepository) GetMerchantDetails(merchantID string) (entities.Merchant, error) {
	input := &dynamodb.GetItemInput{
		TableName: aws.String(r.tableName),
		Key: map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{Value: merchantID},
			"SK": &types.AttributeValueMemberS{Value: "MERCHANT"},
		},
	}

	result, err := r.db.GetItem(context.TODO(), input)
	if err != nil {
		return entities.Merchant{}, err
	}

	var item MerchantItem

	err = attributevalue.UnmarshalMap(result.Item, &item)
	if err != nil {
		return entities.Merchant{}, err
	}

	return entities.Merchant{
		ID: item.PK,
		AccountDetails: entities.AccountDetails{
			Name:     item.AccountDetails.Name,
			IBAN:     item.AccountDetails.IBAN,
			BIC:      item.AccountDetails.BIC,
			Currency: item.AccountDetails.Currency,
		},
	}, nil
}

func (r *DynamoDBRepository) CreateNewPayment(payment entities.Payment) error {
	item := NewPaymentsItemFromPayment(payment)

	av, err := attributevalue.MarshalMap(item)
	if err != nil {
		return err
	}

	input := &dynamodb.PutItemInput{
		Item:      av,
		TableName: aws.String(r.tableName),
	}

	_, err = r.db.PutItem(context.TODO(), input)
	if err != nil {
		return err
	}

	return nil
}

func (r *DynamoDBRepository) UpdatePayment(payment entities.Payment) error {
	// NOTE: DynamoDB PutItem operation also updates the item if it exists
	err := r.CreateNewPayment(payment)
	if err != nil {
		return err
	}

	// NOTE: this UpdateItem operation could be used to update specific fields only
	// input := &dynamodb.UpdateItemInput{
	// 	TableName: aws.String(r.tableName),
	// 	Key: map[string]types.AttributeValue{
	// 		"PK": &types.AttributeValueMemberS{Value: payment.Merchant.ID},
	// 		"SK": &types.AttributeValueMemberS{Value: "PAYMENT#" + payment.ID},
	// 	},
	// 	UpdateExpression: aws.String("SET Refunded = :r, RefundTimestamp = :rt"),
	// 	ExpressionAttributeValues: map[string]types.AttributeValue{
	// 		":r": &types.AttributeValueMemberBOOL{Value: true},
	// 		":rt": &types.AttributeValueMemberN{
	// 			Value: strconv.Itoa(int(now().UnixNano() / int64(time.Millisecond))),
	// 		},
	// 	},
	// }
	//
	// _, err := r.db.UpdateItem(context.TODO(), input)
	// if err != nil {
	// 	return err
	// }

	return nil
}

func (r *DynamoDBRepository) GetPayment(merchantID, paymentID string) (entities.Payment, error) {
	input := &dynamodb.GetItemInput{
		TableName: aws.String(r.tableName),
		Key: map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{Value: merchantID},
			"SK": &types.AttributeValueMemberS{Value: "PAYMENT#" + paymentID},
		},
	}

	result, err := r.db.GetItem(context.TODO(), input)
	if err != nil {
		return entities.Payment{}, err
	}

	var item PaymentsItem

	err = attributevalue.UnmarshalMap(result.Item, &item)
	if err != nil {
		return entities.Payment{}, err
	}

	amount, err := strconv.Atoi(strings.Split(item.DATA, "#")[1])
	if err != nil {
		return entities.Payment{}, err
	}

	return entities.Payment{
		ID:       strings.Split(item.SK, "#")[1],
		Merchant: entities.Merchant{ID: item.PK},
		Customer: entities.Customer{
			ID: item.CustomerID,
			CardDetails: entities.CardDetails{
				Name:           item.CardDetails.Name,
				Number:         item.CardDetails.Number,
				SecurityCode:   item.CardDetails.SecurityCode,
				ExpirationDate: item.CardDetails.ExpirationDate,
			},
		},
		Price: entities.Money{
			Amount:   int64(amount),
			Currency: strings.Split(item.DATA, "#")[0],
		},
		Timestamp:       item.Timestamp,
		Refunded:        item.Refunded,
		RefundTimestamp: item.RefundTimestamp,
	}, nil
}
