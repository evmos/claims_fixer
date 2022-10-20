package internal

import (
	"database/sql"
	"fmt"
	"math/big"
	"sync"
)

type Process struct {
	mutex                  sync.Mutex
	tx                     *sql.Tx
	insertClaims           *sql.Stmt
	insertBalances         *sql.Stmt
	totalProcessedAccounts int
}

func NewProcess() Process {
	db := CreateDatabase("./results.db")
	fmt.Println("Database initialized")
	txClaims, insertClaims := CreateInsertClaimsQuery(db)
	insertBalances := CreateInsertBalanceQuery(db, txClaims)

	return Process{
		mutex:                  sync.Mutex{},
		tx:                     txClaims,
		insertClaims:           insertClaims,
		insertBalances:         insertBalances,
		totalProcessedAccounts: 0,
	}
}

func (p *Process) TxCommit() {
	err := p.tx.Commit()
	if err != nil {
		fmt.Printf("Error commiting transaction: %q", err)
		panic("Failed to commit tx")
	}
	fmt.Println("All addresses added")
}

func (p *Process) InsertClaims(address string, preBalance string, postBalance string, preSequence string, postSequence string, affected bool) {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	// Store the data
	_, err := p.insertClaims.Exec(address, preBalance, postBalance, preSequence, postSequence, affected)
	if err != nil {
		fmt.Println("Error adding:", err)
		panic("Stop processing")
	}
	fmt.Println("Processed accounts: ", p.totalProcessedAccounts)
	p.totalProcessedAccounts = p.totalProcessedAccounts + 1
}

func (p *Process) InsertBalances(address string, denom string, pre string, post string) {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	_, err := p.insertBalances.Exec(address, denom, pre, post)
	if err != nil {
		fmt.Println("Error adding ibc:", err)
		panic("Stop processing")
	}
}

func (p *Process) DoWork(accounts []string) {
	for _, address := range accounts {
		fmt.Println("Processing account", address)
		// Get the sequence number, if it's 0 continue processing
		accountPreRes := GetAccount(address, PreHeight)
		if accountPreRes.Account.BaseAccount.Sequence == "0" {
			balancePreRes := GetBalances(address, PreHeight)
			balancePostRes := GetBalances(address, PostHeight)
			// Make sure that the account still exists
			accountPostRes := GetAccount(address, PostHeight)

			// Data to store
			ibcBalance := make(map[string][]string)
			preBalance := "-1"
			postBalance := "-1"
			affected := false

			for _, k := range balancePreRes.Balances {
				if k.Denom == "aevmos" {
					preBalance = k.Amount
				} else {
					ibcBalance[k.Denom] = []string{k.Amount, "0"}
				}
			}

			for _, k := range balancePostRes.Balances {
				if k.Denom == "aevmos" {
					postBalance = k.Amount
				} else {
					ibcBalance[k.Denom][1] = k.Amount
				}
			}

			old := new(big.Int)
			old, ok := old.SetString(preBalance, 10)
			if !ok {
				fmt.Println("Error parsing the balance", preBalance, ok)
				return
			}

			newBalance := new(big.Int)
			newBalance, ok = newBalance.SetString(postBalance, 10)
			if !ok {
				fmt.Println("Error parsing the balance", postBalance, ok)
				return
			}

			if newBalance.Cmp(old) == -1 {
				affected = true
			}
			p.InsertClaims(address, preBalance, postBalance, accountPreRes.Account.BaseAccount.Sequence, accountPostRes.Account.BaseAccount.Sequence, affected)
			for k, v := range ibcBalance {
				p.InsertBalances(address, k, v[0], v[1])
			}
		}
	}
}
