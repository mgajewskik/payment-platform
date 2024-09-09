package entities

type Merchant struct {
	ID             string
	AccountDetails AccountDetails
}

type AccountDetails struct {
	Name     string
	IBAN     string
	BIC      string
	Currency string
}
