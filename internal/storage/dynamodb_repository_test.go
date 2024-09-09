package storage

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/mgajewskik/payment-platform/internal/domain/entities"
)

type MockDynamoDBClient struct {
	mock.Mock
}

func (m *MockDynamoDBClient) PutItem(
	ctx context.Context,
	params *dynamodb.PutItemInput,
	_ ...func(*dynamodb.Options),
) (*dynamodb.PutItemOutput, error) {
	args := m.Called(ctx, params)
	return &dynamodb.PutItemOutput{}, args.Error(0)
}

func (m *MockDynamoDBClient) UpdateItem(
	ctx context.Context,
	params *dynamodb.UpdateItemInput,
	_ ...func(*dynamodb.Options),
) (*dynamodb.UpdateItemOutput, error) {
	args := m.Called(ctx, params)
	return &dynamodb.UpdateItemOutput{}, args.Error(0)
}

func (m *MockDynamoDBClient) GetItem(
	ctx context.Context,
	params *dynamodb.GetItemInput,
	_ ...func(*dynamodb.Options),
) (*dynamodb.GetItemOutput, error) {
	args := m.Called(ctx, params)
	cast, _ := args.Get(0).(*dynamodb.GetItemOutput)
	return cast, args.Error(1)
}

func TestGetMerchantDetails(t *testing.T) {
	md := MockDynamoDBClient{}
	repo := DynamoDBRepository{
		db:        &md,
		tableName: "table",
		logger:    nil,
	}

	t.Run("should get merchant details", func(t *testing.T) {
		md.On("GetItem", mock.Anything, mock.Anything).Return(&dynamodb.GetItemOutput{
			Item: map[string]types.AttributeValue{
				"PK": &types.AttributeValueMemberS{Value: "merchantID"},
				"SK": &types.AttributeValueMemberS{Value: "MERCHANT"},
				"AccountDetails": &types.AttributeValueMemberM{
					Value: map[string]types.AttributeValue{
						"Name": &types.AttributeValueMemberS{Value: "Test Name"},
						"IBAN": &types.AttributeValueMemberS{
							Value: "PL61109010140000071219812874",
						},
						"BIC":      &types.AttributeValueMemberS{Value: "WBKPPLPP"},
						"Currency": &types.AttributeValueMemberS{Value: "PLN"},
					},
				},
			},
		}, nil)

		// tested function
		got, err := repo.GetMerchantDetails("merchantID")
		assert.NoError(t, err)

		want := entities.Merchant{
			ID: "merchantID",
			AccountDetails: entities.AccountDetails{
				Name:     "Test Name",
				IBAN:     "PL61109010140000071219812874",
				BIC:      "WBKPPLPP",
				Currency: "PLN",
			},
		}

		assert.Equal(t, got, want)
	})
}

func TestCreateNewPayment(t *testing.T) {
	md := MockDynamoDBClient{}
	repo := DynamoDBRepository{
		db:        &md,
		tableName: "table",
		logger:    nil,
	}

	t.Run("should create new payment", func(t *testing.T) {
		payment := entities.Payment{
			ID: "paymentID",
			Merchant: entities.Merchant{
				ID: "merchantID",
			},
			Price: entities.Money{
				Amount:   100,
				Currency: "USD",
			},
			Customer: entities.Customer{
				ID: "customerID",
				CardDetails: entities.CardDetails{
					Name:           "name",
					Number:         "number",
					SecurityCode:   123,
					ExpirationDate: "expirationDate",
				},
			},
			Timestamp:       123,
			Refunded:        false,
			RefundTimestamp: 0,
		}

		input, _ := attributevalue.MarshalMap(PaymentsItem{
			PK:         payment.Merchant.ID,
			SK:         "PAYMENT#paymentID",
			DATA:       "USD#100",
			CustomerID: payment.Customer.ID,
			CardDetails: CardDetails{
				Name:           payment.Customer.CardDetails.Name,
				Number:         payment.Customer.CardDetails.Number,
				SecurityCode:   payment.Customer.CardDetails.SecurityCode,
				ExpirationDate: payment.Customer.CardDetails.ExpirationDate,
			},
			Timestamp:       payment.Timestamp,
			Refunded:        payment.Refunded,
			RefundTimestamp: payment.RefundTimestamp,
		})

		md.On("PutItem", context.TODO(), &dynamodb.PutItemInput{
			TableName: aws.String("table"),
			Item:      input,
		}).Return(nil)

		// tested function
		err := repo.CreateNewPayment(payment)
		assert.NoError(t, err)
	})
}

func TestUpdatePayment(t *testing.T) {
	md := MockDynamoDBClient{}
	repo := DynamoDBRepository{
		db:        &md,
		tableName: "table",
		logger:    nil,
	}

	t.Run("should update payment", func(t *testing.T) {
		payment := entities.Payment{
			ID: "paymentID",
			Merchant: entities.Merchant{
				ID: "merchantID",
			},
			Price: entities.Money{
				Amount:   100,
				Currency: "USD",
			},
			Customer: entities.Customer{
				ID: "customerID",
				CardDetails: entities.CardDetails{
					Name:           "name",
					Number:         "number",
					SecurityCode:   123,
					ExpirationDate: "expirationDate",
				},
			},
			Timestamp:       123,
			Refunded:        false,
			RefundTimestamp: 0,
		}

		input, _ := attributevalue.MarshalMap(PaymentsItem{
			PK:         payment.Merchant.ID,
			SK:         "PAYMENT#paymentID",
			DATA:       "USD#100",
			CustomerID: payment.Customer.ID,
			CardDetails: CardDetails{
				Name:           payment.Customer.CardDetails.Name,
				Number:         payment.Customer.CardDetails.Number,
				SecurityCode:   payment.Customer.CardDetails.SecurityCode,
				ExpirationDate: payment.Customer.CardDetails.ExpirationDate,
			},
			Timestamp:       payment.Timestamp,
			Refunded:        payment.Refunded,
			RefundTimestamp: payment.RefundTimestamp,
		})

		md.On("PutItem", context.TODO(), &dynamodb.PutItemInput{
			TableName: aws.String("table"),
			Item:      input,
		}).Return(nil)

		// tested function
		err := repo.UpdatePayment(payment)
		assert.NoError(t, err)
	})

	// t.Run("should update payment with update action", func(t *testing.T) {
	// 	repo := DynamoDBRepository{
	// 		db:        &md,
	// 		tableName: "table",
	// 		logger:    nil,
	// 	}
	//
	// 	md.On("UpdateItem", context.TODO(), &dynamodb.UpdateItemInput{
	// 		TableName: aws.String("table"),
	// 		Key: map[string]types.AttributeValue{
	// 			"PK": &types.AttributeValueMemberS{Value: "merchantID"},
	// 			"SK": &types.AttributeValueMemberS{Value: "PAYMENT#paymentID"},
	// 		},
	// 		ExpressionAttributeValues: map[string]types.AttributeValue{
	// 			":r": &types.AttributeValueMemberBOOL{Value: true},
	// 			":rt": &types.AttributeValueMemberN{
	// 				Value: strconv.Itoa(int(now().Unix())),
	// 			},
	// 		},
	// 		UpdateExpression: aws.String("SET Refunded = :r, RefundTimestamp = :rt"),
	// 	}).Return(nil)
	//
	//  // tested function
	// 	err := repo.UpdatePayment("merchantID", "paymentID")
	// 	assert.NoError(t, err)
	// })
}

func TestGetPayment(t *testing.T) {
	md := MockDynamoDBClient{}
	repo := DynamoDBRepository{
		db:        &md,
		tableName: "table",
		logger:    nil,
	}

	t.Run("should get payment", func(t *testing.T) {
		md.On("GetItem", mock.Anything, mock.Anything).Return(&dynamodb.GetItemOutput{
			Item: map[string]types.AttributeValue{
				"PK":         &types.AttributeValueMemberS{Value: "merchantID"},
				"SK":         &types.AttributeValueMemberS{Value: "PAYMENT#paymentID"},
				"DATA":       &types.AttributeValueMemberS{Value: "USD#100"},
				"CustomerID": &types.AttributeValueMemberS{Value: "customerID"},
				"CardDetails": &types.AttributeValueMemberM{
					Value: map[string]types.AttributeValue{
						"Name":           &types.AttributeValueMemberS{Value: "Test Name"},
						"Number":         &types.AttributeValueMemberS{Value: "1234123412341234"},
						"SecurityCode":   &types.AttributeValueMemberN{Value: "123"},
						"ExpirationDate": &types.AttributeValueMemberS{Value: "12/23"},
					},
				},
				"Timestamp":       &types.AttributeValueMemberN{Value: "123"},
				"Refunded":        &types.AttributeValueMemberBOOL{Value: false},
				"RefundTimestamp": &types.AttributeValueMemberN{Value: "0"},
			},
		}, nil)

		// tested function
		got, err := repo.GetPayment("merchantID", "paymentID")
		assert.NoError(t, err)

		want := entities.Payment{
			ID: "paymentID",
			Merchant: entities.Merchant{
				ID: "merchantID",
			},
			Customer: entities.Customer{
				ID: "customerID",
				CardDetails: entities.CardDetails{
					Name:           "Test Name",
					Number:         "1234123412341234",
					SecurityCode:   123,
					ExpirationDate: "12/23",
				},
			},
			Price: entities.Money{
				Amount:   100,
				Currency: "USD",
			},
			Timestamp:       123,
			Refunded:        false,
			RefundTimestamp: 0,
		}

		assert.Equal(t, got, want)
	})
}
