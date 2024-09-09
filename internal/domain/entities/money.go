package entities

type Money struct {
	Amount   int64 // storing as int to avoid floating point precision issues
	Currency string
}
