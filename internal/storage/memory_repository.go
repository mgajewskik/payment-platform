package storage

import (
	"fmt"

	"github.com/mgajewskik/payment-platform/internal/domain/entities"
)

type MemoryRepository struct {
	payments map[string]entities.Payment
}

func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{
		payments: make(map[string]entities.Payment),
	}
}

func (r *MemoryRepository) GetMerchantDetails(merchantID string) (entities.Merchant, error) {
	return entities.Merchant{ID: merchantID, AccountDetails: entities.AccountDetails{
		Name:     "Test Merchant",
		IBAN:     "DE89370400440532013000",
		BIC:      "COBADEFFXXX",
		Currency: "EUR",
	}}, nil
}

func (r *MemoryRepository) CreateNewPayment(payment entities.Payment) error {
	r.payments[payment.ID] = payment

	return nil
}

func (r *MemoryRepository) UpdatePayment(payment entities.Payment) error {
	r.payments[payment.ID] = payment

	return nil
}

func (r *MemoryRepository) GetPayment(_, paymentID string) (entities.Payment, error) {
	if _, ok := r.payments[paymentID]; !ok {
		return entities.Payment{}, fmt.Errorf("payment with ID %s not found", paymentID)
	}

	return r.payments[paymentID], nil
}
