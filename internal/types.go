package internal

type Balance struct {
	Amount string `json:"amount"`
	Denom  string `json:"denom"`
}
type BalanceResponse struct {
	Balance Balance `json:"balance"`
}

type BalancesResponse struct {
	Balances []Balance `json:"balances"`
}

// Note only sequece is getting unmarshalled
type BaseAccount struct {
	Sequence string `json:"sequence"`
}

type Account struct {
	BaseAccount BaseAccount `json:"base_account"`
}

type AccountResponse struct {
	Account Account `json:"account"`
}
