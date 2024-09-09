package main

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/mgajewskik/payment-platform/internal/domain/entities"
	"github.com/mgajewskik/payment-platform/internal/domain/service"
	"github.com/mgajewskik/payment-platform/internal/domain/simulator"
	"github.com/mgajewskik/payment-platform/internal/storage"
	"github.com/pascaldekloe/jwt"
	"github.com/stretchr/testify/assert"
)

func TestGenerateToken(t *testing.T) {
	app := &application{
		config: config{
			merchantID: "testMerchant",
			baseURL:    "http://localhost",
			jwt: struct {
				secretKey string
			}{
				secretKey: "testSecret",
			},
		},
	}

	req, err := http.NewRequest("GET", "/token", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(app.generateToken)
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	responseMap := make(map[string]string)
	err = json.Unmarshal(rr.Body.Bytes(), &responseMap)
	if err != nil {
		t.Fatal(err)
	}

	tokenString := responseMap["AuthenticationToken"]
	claims, err := jwt.HMACCheck([]byte(tokenString), []byte(app.config.jwt.secretKey))
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, app.config.merchantID, claims.Subject)

	assert.NotNil(t, claims)

	expiry, err := time.Parse(time.RFC3339, responseMap["AuthenticationTokenExpiry"])
	assert.Nil(t, err)
	assert.True(t, expiry.After(time.Now()))
}

func TestCreatePayment(t *testing.T) {
	logger := slog.Default()
	bank := simulator.NewBankSimulator(logger)
	storage := storage.NewMemoryRepository()
	app := &application{
		config: config{
			merchantID: "testMerchant",
			baseURL:    "http://localhost",
			jwt: struct {
				secretKey string
			}{
				secretKey: "testSecret",
			},
		},
		service: service.NewService(storage, bank, logger),
		logger:  logger,
	}

	paymentRequest := map[string]interface{}{
		"CustomerID":     "testCustomer",
		"CustomerName":   "Test Customer",
		"CardNumber":     "1234123412341234",
		"CardCVV":        123,
		"CardExpiryDate": "12/23",
		"Price":          1000,
		"Currency":       "USD",
	}

	jsonValue, _ := json.Marshal(paymentRequest)

	t.Run("should create payment", func(t *testing.T) {
		r := chi.NewRouter()
		r.Post("/payments", app.createPayment)

		req, err := http.NewRequest("POST", "/payments", bytes.NewBuffer(jsonValue))
		if err != nil {
			t.Fatal(err)
		}

		rr := httptest.NewRecorder()

		r.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		responseMap := make(map[string]string)
		err = json.Unmarshal(rr.Body.Bytes(), &responseMap)
		if err != nil {
			t.Fatal(err)
		}

		assert.NotEmpty(t, responseMap["paymentID"])
	})

	t.Run("should create payment with middleware", func(t *testing.T) {
		r := app.routes()

		req, err := http.NewRequest("GET", "/token", nil)
		if err != nil {
			t.Fatal(err)
		}

		rr := httptest.NewRecorder()

		r.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		responseMap := make(map[string]string)
		err = json.Unmarshal(rr.Body.Bytes(), &responseMap)
		if err != nil {
			t.Fatal(err)
		}

		tokenString := responseMap["AuthenticationToken"]

		req, err = http.NewRequest("POST", "/payments", bytes.NewBuffer(jsonValue))
		if err != nil {
			t.Fatal(err)
		}

		req.Header.Set("Authorization", "Bearer "+tokenString)

		rr = httptest.NewRecorder()

		r.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		responseMap = make(map[string]string)
		err = json.Unmarshal(rr.Body.Bytes(), &responseMap)
		if err != nil {
			t.Fatal(err)
		}

		assert.NotEmpty(t, responseMap["paymentID"])
	})
}

func TestGetPayment(t *testing.T) {
	logger := slog.Default()
	bank := simulator.NewBankSimulator(logger)
	storage := storage.NewMemoryRepository()
	app := &application{
		config: config{
			merchantID: "testMerchant",
			baseURL:    "http://localhost",
			jwt: struct {
				secretKey string
			}{
				secretKey: "testSecret",
			},
		},
		service: service.NewService(storage, bank, logger),
		logger:  logger,
	}

	t.Run("should get payment", func(t *testing.T) {
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

		_ = storage.CreateNewPayment(input)

		r := chi.NewRouter()
		r.Get("/payments/{paymentID}", app.getPayment)

		req, err := http.NewRequest("GET", "/payments/00000000-0000-0000-0000-000000000000", nil)
		if err != nil {
			t.Fatal(err)
		}

		rr := httptest.NewRecorder()

		r.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		responseMap := make(map[string]string)
		err = json.Unmarshal(rr.Body.Bytes(), &responseMap)
		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "00000000-0000-0000-0000-000000000000", responseMap["PaymentID"])
		assert.Equal(t, "testMerchantID", responseMap["MerchantID"])
		assert.Equal(t, "testCustomerID", responseMap["CustomerID"])
		assert.Equal(t, "100", responseMap["Price"])
		assert.Equal(t, "USD", responseMap["Currency"])
		assert.Equal(t, "123", responseMap["Timestamp"])
	})

	t.Run("should get payment with middleware", func(t *testing.T) {
		input := entities.Payment{
			ID: "11111111-1111-1111-1111-111111111111",
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

		_ = storage.CreateNewPayment(input)

		r := app.routes()

		req, err := http.NewRequest("GET", "/token", nil)
		if err != nil {
			t.Fatal(err)
		}

		rr := httptest.NewRecorder()

		r.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		responseMap := make(map[string]string)
		err = json.Unmarshal(rr.Body.Bytes(), &responseMap)
		if err != nil {
			t.Fatal(err)
		}

		tokenString := responseMap["AuthenticationToken"]

		req, err = http.NewRequest("GET", "/payments/11111111-1111-1111-1111-111111111111", nil)
		if err != nil {
			t.Fatal(err)
		}

		req.Header.Set("Authorization", "Bearer "+tokenString)

		rr = httptest.NewRecorder()

		r.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		responseMap = make(map[string]string)
		err = json.Unmarshal(rr.Body.Bytes(), &responseMap)
		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "11111111-1111-1111-1111-111111111111", responseMap["PaymentID"])
		assert.Equal(t, "testMerchantID", responseMap["MerchantID"])
		assert.Equal(t, "testCustomerID", responseMap["CustomerID"])
		assert.Equal(t, "100", responseMap["Price"])
		assert.Equal(t, "USD", responseMap["Currency"])
		assert.Equal(t, "123", responseMap["Timestamp"])
	})
}

func TestRefundPayment(t *testing.T) {
	logger := slog.Default()
	bank := simulator.NewBankSimulator(logger)
	storage := storage.NewMemoryRepository()
	app := &application{
		config: config{
			merchantID: "testMerchant",
			baseURL:    "http://localhost",
			jwt: struct {
				secretKey string
			}{
				secretKey: "testSecret",
			},
		},
		service: service.NewService(storage, bank, logger),
		logger:  logger,
	}

	t.Run("should refund payment", func(t *testing.T) {
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

		_ = storage.CreateNewPayment(input)

		r := chi.NewRouter()
		r.Patch("/payments/{paymentID}/refund", app.refundPayment)

		req, err := http.NewRequest(
			"PATCH",
			"/payments/00000000-0000-0000-0000-000000000000/refund",
			nil,
		)
		if err != nil {
			t.Fatal(err)
		}

		rr := httptest.NewRecorder()

		r.ServeHTTP(rr, req)

		got, _ := storage.GetPayment("testMerchantID", "00000000-0000-0000-0000-000000000000")

		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Equal(t, true, got.Refunded)
	})

	t.Run("should refund payment with middleware", func(t *testing.T) {
		input := entities.Payment{
			ID: "11111111-1111-1111-1111-111111111111",
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

		_ = storage.CreateNewPayment(input)

		r := app.routes()

		req, err := http.NewRequest("GET", "/token", nil)
		if err != nil {
			t.Fatal(err)
		}

		rr := httptest.NewRecorder()

		r.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		responseMap := make(map[string]string)
		err = json.Unmarshal(rr.Body.Bytes(), &responseMap)
		if err != nil {
			t.Fatal(err)
		}

		tokenString := responseMap["AuthenticationToken"]

		req, err = http.NewRequest(
			"PATCH",
			"/payments/11111111-1111-1111-1111-111111111111/refund",
			nil,
		)
		if err != nil {
			t.Fatal(err)
		}

		req.Header.Set("Authorization", "Bearer "+tokenString)

		rr = httptest.NewRecorder()

		r.ServeHTTP(rr, req)

		got, _ := storage.GetPayment("testMerchantID", "11111111-1111-1111-1111-111111111111")

		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Equal(t, true, got.Refunded)
	})
}
