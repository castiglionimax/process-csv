package domain

type (
	AccountID string
	Account   struct {
		ID    AccountID `json:"account_id"`
		Name  string    `json:"name"`
		Email string    `json:"email"`
	}
)

func (a AccountID) String() string {
	return string(a)
}
