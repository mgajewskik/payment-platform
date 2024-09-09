package entities

type Payment struct {
	ID                string
	Merchant          Merchant
	Customer          Customer
	Price             Money
	BankTransactionID string
	Timestamp         int64
	Refunded          bool
	RefundTimestamp   int64
}

// PaymentDetails without sensitive information
type PaymentDetails struct {
	ID              string
	MerchantID      string
	CustomerID      string
	Price           Money
	Timestamp       int64
	Refunded        bool
	RefundTimestamp int64
}

func NewPaymentDetailsFromPayment(payment Payment) PaymentDetails {
	return PaymentDetails{
		ID:              payment.ID,
		MerchantID:      payment.Merchant.ID,
		CustomerID:      payment.Customer.ID,
		Price:           payment.Price,
		Timestamp:       payment.Timestamp,
		Refunded:        payment.Refunded,
		RefundTimestamp: payment.RefundTimestamp,
	}
}
