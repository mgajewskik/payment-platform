package service

import (
	"log/slog"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/mgajewskik/payment-platform/internal/domain/entities"
	"github.com/mgajewskik/payment-platform/internal/domain/simulator"
	"github.com/mgajewskik/payment-platform/internal/storage"
	"github.com/stretchr/testify/assert"
)

// NOTE: business logic tests

func TestCreateNewPayment(t *testing.T) {
	logger := slog.Default()
	service := NewService(storage.NewMemoryRepository(), simulator.NewBankSimulator(logger), logger)
	now = func() time.Time {
		return time.Unix(100, 100)
	}

	newUUID = func() uuid.UUID {
		return uuid.MustParse("00000000-0000-0000-0000-000000000000")
	}

	input := entities.Payment{
		Merchant: entities.Merchant{
			ID: "testMerchantID",
		},
		Customer: entities.Customer{
			ID: "testCustomerID",
			CardDetails: entities.CardDetails{
				Number:         "1234567890123456",
				Name:           "Test Customer",
				SecurityCode:   123,
				ExpirationDate: "12/23",
			},
		},
		Price: entities.Money{
			Amount:   100,
			Currency: "USD",
		},
		Timestamp: now().UnixNano() / int64(time.Millisecond),
	}

	// tested function
	_, err := service.CreateNewPayment(input)
	if err != nil {
		t.Errorf("error creating new payment: %v", err)
	}

	got, err := service.storage.GetPayment(input.Merchant.ID, newUUID().String())
	if err != nil {
		t.Errorf("error getting payment: %v", err)
	}

	want := entities.Payment{
		ID: "00000000-0000-0000-0000-000000000000",
		Merchant: entities.Merchant{
			ID: "testMerchantID",
			AccountDetails: entities.AccountDetails{
				Name:     "Test Merchant",
				IBAN:     "DE89370400440532013000",
				BIC:      "COBADEFFXXX",
				Currency: "EUR",
			},
		},
		Customer: entities.Customer{
			ID: "testCustomerID",
			CardDetails: entities.CardDetails{
				Name:           "Test Customer",
				Number:         "1234567890123456",
				SecurityCode:   123,
				ExpirationDate: "12/23",
			},
		},
		Price:             entities.Money{Amount: 100, Currency: "USD"},
		BankTransactionID: "simulatedTransactionID",
		Timestamp:         100000,
		Refunded:          false,
		RefundTimestamp:   0,
	}

	assert.Equal(t, got, want)
}

func TestGetPaymentDetails(t *testing.T) {
	logger := slog.Default()
	service := NewService(storage.NewMemoryRepository(), simulator.NewBankSimulator(logger), logger)
	input := entities.Payment{
		ID: "00000000-0000-0000-0000-000000000000",
		Merchant: entities.Merchant{
			ID: "testMerchantID",
			AccountDetails: entities.AccountDetails{
				Name:     "Test Merchant",
				IBAN:     "DE89370400440532013000",
				BIC:      "COBADEFFXXX",
				Currency: "EUR",
			},
		},
		Customer: entities.Customer{
			ID: "testCustomerID",
			CardDetails: entities.CardDetails{
				Number:         "1234567890123456",
				Name:           "Test Customer",
				SecurityCode:   123,
				ExpirationDate: "12/23",
			},
		},
		Price: entities.Money{
			Amount:   100,
			Currency: "USD",
		},
		BankTransactionID: "simulatedTransactionID",
		Timestamp:         123,
	}

	_ = service.storage.CreateNewPayment(input)

	// tested function
	got, err := service.GetPaymentDetails("testMerchantID", "00000000-0000-0000-0000-000000000000")
	if err != nil {
		t.Errorf("error getting payment details: %v", err)
	}

	want := entities.PaymentDetails{
		ID:              "00000000-0000-0000-0000-000000000000",
		MerchantID:      "testMerchantID",
		CustomerID:      "testCustomerID",
		Price:           entities.Money{Amount: 100, Currency: "USD"},
		Timestamp:       123,
		Refunded:        false,
		RefundTimestamp: 0,
	}

	assert.Equal(t, got, want)
}

func TestRefundPayment(t *testing.T) {
	logger := slog.Default()
	service := NewService(storage.NewMemoryRepository(), simulator.NewBankSimulator(logger), logger)
	now = func() time.Time {
		return time.Unix(100, 100)
	}

	input := entities.Payment{
		ID: "00000000-0000-0000-0000-000000000000",
		Merchant: entities.Merchant{
			ID: "testMerchantID",
			AccountDetails: entities.AccountDetails{
				Name:     "Test Merchant",
				IBAN:     "DE89370400440532013000",
				BIC:      "COBADEFFXXX",
				Currency: "EUR",
			},
		},
		Customer: entities.Customer{
			ID: "testCustomerID",
			CardDetails: entities.CardDetails{
				Number:         "1234567890123456",
				Name:           "Test Customer",
				SecurityCode:   123,
				ExpirationDate: "12/23",
			},
		},
		Price: entities.Money{
			Amount:   100,
			Currency: "USD",
		},
		BankTransactionID: "simulatedTransactionID",
		Timestamp:         123,
	}

	_ = service.storage.CreateNewPayment(input)

	// tested function
	err := service.RefundPayment("testMerchantID", "00000000-0000-0000-0000-000000000000")
	if err != nil {
		t.Errorf("error refunding payment: %v", err)
	}

	got, _ := service.storage.GetPayment("testMerchantID", "00000000-0000-0000-0000-000000000000")

	want := entities.Payment{
		ID: "00000000-0000-0000-0000-000000000000",
		Merchant: entities.Merchant{
			ID: "testMerchantID",
			AccountDetails: entities.AccountDetails{
				Name:     "Test Merchant",
				IBAN:     "DE89370400440532013000",
				BIC:      "COBADEFFXXX",
				Currency: "EUR",
			},
		},
		Customer: entities.Customer{
			ID: "testCustomerID",
			CardDetails: entities.CardDetails{
				Name:           "Test Customer",
				Number:         "1234567890123456",
				SecurityCode:   123,
				ExpirationDate: "12/23",
			},
		},
		Price:             entities.Money{Amount: 100, Currency: "USD"},
		BankTransactionID: "simulatedTransactionID",
		Timestamp:         123,
		Refunded:          true,
		RefundTimestamp:   100000,
	}

	assert.Equal(t, got, want)
}
