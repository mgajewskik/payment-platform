package simulator

import (
	"log/slog"

	"github.com/mgajewskik/payment-platform/internal/domain/entities"
)

// NOTE: custom errors could be returned from the simulator

type BankClient interface {
	ValidateCardInformation(card entities.CardDetails) error
	ProcessTransaction(
		account entities.AccountDetails,
		card entities.CardDetails,
		money entities.Money,
	) (string, error)
	RevertTransaction(transactionID string) error
}

type BankSimulator struct {
	logger *slog.Logger
}

func NewBankSimulator(logger *slog.Logger) *BankSimulator {
	return &BankSimulator{logger: logger}
}

func (b *BankSimulator) ValidateCardInformation(_ entities.CardDetails) error {
	b.logger.Info("requesting bank to validate card information")
	return nil
}

func (b *BankSimulator) ProcessTransaction(
	_ entities.AccountDetails,
	_ entities.CardDetails,
	_ entities.Money,
) (string, error) {
	// ask a bank to charge the card with the given amount
	// ask a bank to transfer the money to the given account
	b.logger.Info("requesting bank to process transaction")
	return "simulatedTransactionID", nil
}

// RevertTransaction NOTE: assumes that the transaction can be reverted by ID withput passing in the exact details of a transaction
func (b *BankSimulator) RevertTransaction(
	_ string,
) error {
	b.logger.Info("requesting bank to revert transaction")
	return nil
}
