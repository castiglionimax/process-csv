package domain

import "time"

type Transaction struct {
	AccountID AccountID `json:"account_id"`
	Date      time.Time `json:"date"`
	Amount    float64   `json:"amount"`
}
