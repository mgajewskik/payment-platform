package service

import (
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/mgajewskik/payment-platform/internal/domain/entities"
	"github.com/mgajewskik/payment-platform/internal/domain/simulator"
	"github.com/mgajewskik/payment-platform/internal/storage"
)

var (
	now     = time.Now
	newUUID = uuid.New
)

type Service struct {
	storage    storage.DBRepository
	bankClient simulator.BankClient
	logger     *slog.Logger
}

func NewService(
	storage storage.DBRepository,
	bankClient simulator.BankClient,
	logger *slog.Logger,
) *Service {
	return &Service{
		storage:    storage,
		bankClient: bankClient,
		logger:     logger,
	}
}

func (s *Service) CreateNewPayment(payment entities.Payment) (string, error) {
	err := s.bankClient.ValidateCardInformation(payment.Customer.CardDetails)
	if err != nil {
		s.logger.Error("error validating card information", "error", err)
		return "", err
	}

	merchant, err := s.storage.GetMerchantDetails(payment.Merchant.ID)
	if err != nil {
		s.logger.Error("error getting merchant details", "error", err)
		return "", err
	}

	transactionID, err := s.bankClient.ProcessTransaction(
		merchant.AccountDetails,
		payment.Customer.CardDetails,
		payment.Price,
	)
	if err != nil {
		s.logger.Error("error processing transaction", "error", err)
		return "", err
	}

	payment.ID = newUUID().String()
	payment.BankTransactionID = transactionID
	payment.Merchant.AccountDetails = merchant.AccountDetails
	payment.Timestamp = now().UnixNano() / int64(time.Millisecond)

	err = s.storage.CreateNewPayment(payment)
	if err != nil {
		s.logger.Error("error creating new payment", "error", err)
		return "", err
	}

	return payment.ID, nil
}

func (s *Service) GetPaymentDetails(merchantID, paymentID string) (entities.PaymentDetails, error) {
	payment, err := s.storage.GetPayment(merchantID, paymentID)
	if err != nil {
		s.logger.Error("error getting payment", "error", err)
		return entities.PaymentDetails{}, err
	}

	s.logger.Info("payment details retrieved", "paymentID", payment.ID)

	return entities.NewPaymentDetailsFromPayment(payment), nil
}

func (s *Service) RefundPayment(merchantID, paymentID string) error {
	payment, err := s.storage.GetPayment(merchantID, paymentID)
	if err != nil {
		s.logger.Error("error getting payment", "error", err)
		return err
	}

	err = s.bankClient.RevertTransaction(payment.BankTransactionID)
	if err != nil {
		s.logger.Error("error reverting transaction", "error", err)
		return err
	}

	payment.Refunded = true
	payment.RefundTimestamp = now().UnixNano() / int64(time.Millisecond)

	err = s.storage.UpdatePayment(payment)
	if err != nil {
		s.logger.Error("error updating payment", "error", err)
		return err
	}

	s.logger.Info("payment refunded", "paymentID", payment.ID)

	return nil
}
