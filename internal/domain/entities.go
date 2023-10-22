package domain

type (
	AccountID string
	Account   struct {
		ID     AccountID `json:"account_id"`
		Name   string    `json:"name"`
		Number int64     `json:"account_number"`
	}
)

func (a AccountID) String() string {
	return string(a)
}
