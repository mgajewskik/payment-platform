package main

import (
	"context"
	"net/http"
)

type contextKey string

const (
	authenticatedMerchantContextKey = contextKey("authenticatedMerchantID")
)

func contextSetAuthenticatedMerchantID(r *http.Request, merchantID string) *http.Request {
	ctx := context.WithValue(r.Context(), authenticatedMerchantContextKey, merchantID)
	return r.WithContext(ctx)
}

func contextGetAuthenticatedMerchantID(r *http.Request) string {
	merchantID, ok := r.Context().Value(authenticatedMerchantContextKey).(string)
	if !ok {
		return ""
	}

	return merchantID
}
