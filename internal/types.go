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

// Genesis parser
type Genesis struct {
	AppState AppState `json:"app_state"`
}

type AppState struct {
	Claims Claims `json:"claims"`
}

type Claims struct {
	ClaimsRecords []ClaimsRecord `json:"claims_records"`
}

type ClaimsRecord struct {
	Address string `json:"address"`
}

// BlockResult
type BlockResult struct {
	Result Result `json:"result"`
}

type Result struct {
	EndBlockEvents []EndBlockEvent `json:"end_block_events"`
}

type EndBlockEvent struct {
	Type       string      `json:"type"`
	Attributes []Attribute `json:"attributes"`
}

type Attribute struct {
	Value string `json:"value"`
}
