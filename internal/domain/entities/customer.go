package entities

type Customer struct {
	ID          string
	CardDetails CardDetails
}

type CardDetails struct {
	Name           string
	Number         string
	SecurityCode   int
	ExpirationDate string
}
