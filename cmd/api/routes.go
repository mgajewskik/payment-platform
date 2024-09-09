package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func (app *application) routes() http.Handler {
	mux := chi.NewRouter()

	mux.NotFound(app.notFound)
	mux.MethodNotAllowed(app.methodNotAllowed)

	mux.Use(app.logAccess)
	mux.Use(app.recoverPanic)
	mux.Use(app.authenticate)

	mux.Get("/status", app.status)
	mux.Get("/token", app.generateToken)

	mux.Group(func(mux chi.Router) {
		mux.Use(app.requireAuthenticatedMerchant)

		mux.Post("/payments", app.createPayment)
		mux.Get("/payments/{paymentID}", app.getPayment)
		mux.Patch("/payments/{paymentID}/refund", app.refundPayment)
	})

	return mux
}
