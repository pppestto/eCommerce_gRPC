package domain

import "errors"

var (
	ErrInvalidAmount   = errors.New("amount must be non-negative")
	ErrInvalidCurrency = errors.New("currency cannot be empty")
)

type Money struct {
	Amount   int64
	Currency string
}

func NewMoney(amount int64, currency string) (*Money, error) {
	if amount < 0 {
		return nil, ErrInvalidAmount
	}
	if currency == "" {
		return nil, ErrInvalidCurrency
	}
	return &Money{Amount: amount, Currency: currency}, nil
}
