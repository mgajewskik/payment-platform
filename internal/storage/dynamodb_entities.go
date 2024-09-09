package storage

import (
	"strconv"

	"github.com/mgajewskik/payment-platform/internal/domain/entities"
)

type PaymentsItem struct {
	PK              string `dynamodbav:"PK"` // merchantID
	SK              string `dynamodbav:"SK"` // PAYMENTS#paymentID
	DATA            string `dynamodbav:"DATA"`
	CustomerID      string `dynamodbav:"CustomerID"`
	CardDetails     CardDetails
	Timestamp       int64 `dynamodbav:"Timestamp"`
	Refunded        bool  `dynamodbav:"Refunded"`
	RefundTimestamp int64 `dynamodbav:"RefundTimestamp"`
}

func NewPaymentsItemFromPayment(payment entities.Payment) PaymentsItem {
	return PaymentsItem{
		PK:         payment.Merchant.ID,
		SK:         "PAYMENT#" + payment.ID,
		DATA:       payment.Price.Currency + "#" + strconv.Itoa(int(payment.Price.Amount)),
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
	}
}

// CardDetails NOTE: duplicating this model as it can differ from the business model in the future
type CardDetails struct {
	Name           string `dynamodbav:"name"`
	Number         string `dynamodbav:"number"`
	SecurityCode   int    `dynamodbav:"securityCode"`
	ExpirationDate string `dynamodbav:"expirationDate"`
}

type MerchantItem struct {
	PK             string `dynamodbav:"PK"` // merchantID
	SK             string `dynamodbav:"SK"` // MERCHANT
	AccountDetails AccountDetails
}

type AccountDetails struct {
	Name     string `dynamodbav:"name"`
	IBAN     string `dynamodbav:"iban"`
	BIC      string `dynamodbav:"bic"`
	Currency string `dynamodbav:"currency"`
}
