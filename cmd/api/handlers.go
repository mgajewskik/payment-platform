package main

import (
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/pascaldekloe/jwt"

	"github.com/mgajewskik/payment-platform/internal/domain/entities"
	"github.com/mgajewskik/payment-platform/internal/request"
	"github.com/mgajewskik/payment-platform/internal/response"
	"github.com/mgajewskik/payment-platform/internal/validator"
)

func (app *application) status(w http.ResponseWriter, r *http.Request) {
	data := map[string]string{
		"Status": "OK",
	}

	err := response.JSON(w, http.StatusOK, data)
	if err != nil {
		app.serverError(w, r, err)
	}
}

func (app *application) generateToken(w http.ResponseWriter, r *http.Request) {
	var claims jwt.Claims
	claims.Subject = app.config.merchantID

	expiry := time.Now().Add(24 * time.Hour)
	claims.Issued = jwt.NewNumericTime(time.Now())
	claims.NotBefore = jwt.NewNumericTime(time.Now())
	claims.Expires = jwt.NewNumericTime(expiry)

	claims.Issuer = app.config.baseURL
	claims.Audiences = []string{app.config.baseURL}

	jwtBytes, err := claims.HMACSign(jwt.HS256, []byte(app.config.jwt.secretKey))
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	data := map[string]string{
		"AuthenticationToken":       string(jwtBytes),
		"AuthenticationTokenExpiry": expiry.Format(time.RFC3339),
	}

	err = response.JSON(w, http.StatusOK, data)
	if err != nil {
		app.serverError(w, r, err)
	}
}

func (app *application) createPayment(w http.ResponseWriter, r *http.Request) {
	merchantID := contextGetAuthenticatedMerchantID(r)

	var input struct {
		CustomerID     string              `json:"CustomerID"`
		CustomerName   string              `json:"CustomerName"`
		CardNumber     string              `json:"CardNumber"`
		CardCVV        int                 `json:"CardCVV"`
		CardExpiryDate string              `json:"CardExpiryDate"`
		Price          int64               `json:"Price"`
		Currency       string              `json:"Currency"`
		Validator      validator.Validator `json:"-"`
	}

	err := request.DecodeJSON(w, r, &input)
	if err != nil {
		app.badRequest(w, r, err)
		return
	}

	input.Validator.CheckField(input.CustomerID != "", "CustomerID", "CustomerID is required")
	input.Validator.CheckField(input.CustomerName != "", "CustomerName", "CustomerName is required")
	input.Validator.CheckField(input.CardNumber != "", "CardNumber", "CardNumber is required")
	input.Validator.CheckField(
		len([]rune(input.CardNumber)) == 16,
		"CardNumber",
		"CardNumber must be exactly 16 digits",
	)
	input.Validator.CheckField(input.CardCVV != 0, "CardCVV", "CardCVV is required")
	input.Validator.CheckField(
		input.CardCVV >= 100 && input.CardCVV <= 999,
		"CardCVV",
		"CardCVV must be exactly 3 digits",
	)
	input.Validator.CheckField(
		input.CardExpiryDate != "",
		"CardExpiryDate",
		"CardExpiryDate is required",
	)
	input.Validator.CheckField(input.Price != 0, "Price", "Price is required and cannot be zero")
	input.Validator.CheckField(input.Currency != "", "Currency", "Currency is required")

	if input.Validator.HasErrors() {
		app.failedValidation(w, r, input.Validator)
		return
	}

	payment := entities.Payment{
		Merchant: entities.Merchant{ID: merchantID},
		Customer: entities.Customer{ID: input.CustomerID, CardDetails: entities.CardDetails{
			Name:           input.CustomerName,
			Number:         input.CardNumber,
			SecurityCode:   input.CardCVV,
			ExpirationDate: input.CardExpiryDate,
		}},
		Price: entities.Money{Amount: input.Price, Currency: input.Currency},
	}

	paymentID, err := app.service.CreateNewPayment(payment)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	data := map[string]string{
		"paymentID": paymentID,
	}

	err = response.JSON(w, http.StatusOK, data)
	if err != nil {
		app.serverError(w, r, err)
	}
}

func (app *application) getPayment(w http.ResponseWriter, r *http.Request) {
	paymentID := chi.URLParam(r, "paymentID")
	merchantID := contextGetAuthenticatedMerchantID(r)

	paymentDetails, err := app.service.GetPaymentDetails(merchantID, paymentID)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	data := map[string]string{
		"PaymentID":  paymentDetails.ID,
		"MerchantID": paymentDetails.MerchantID,
		"CustomerID": paymentDetails.CustomerID,
		"Price":      strconv.Itoa(int(paymentDetails.Price.Amount)),
		"Currency":   paymentDetails.Price.Currency,
		"Timestamp":  strconv.Itoa(int(paymentDetails.Timestamp)),
	}

	if paymentDetails.Refunded {
		data["Refunded"] = strconv.FormatBool(paymentDetails.Refunded)
		data["RefundTimestamp"] = strconv.Itoa(int(paymentDetails.RefundTimestamp))
	}

	err = response.JSON(w, http.StatusOK, data)
	if err != nil {
		app.serverError(w, r, err)
	}
}

func (app *application) refundPayment(w http.ResponseWriter, r *http.Request) {
	paymentID := chi.URLParam(r, "paymentID")
	merchantID := contextGetAuthenticatedMerchantID(r)

	err := app.service.RefundPayment(merchantID, paymentID)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}
