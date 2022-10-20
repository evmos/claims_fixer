package internal

import (
	"database/sql"
	"fmt"
	"strconv"
	"sync"
)

var dust = "1000000000000000"

type Process struct {
	mutex                 sync.Mutex
	tx                    *sql.Tx
	insertBalances        *sql.Stmt
	totalAffectedAccounts int64
}

func NewProcess() Process {
	db := CreateDatabase("./results.db")
	fmt.Println("Database initialized")
	tx, insertBalances := CreateInsertBalanceQuery(db)

	return Process{
		mutex:                 sync.Mutex{},
		tx:                    tx,
		insertBalances:        insertBalances,
		totalAffectedAccounts: 0,
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

func (p *Process) InsertBalances(address string, amount string) {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	_, err := p.insertBalances.Exec(address, amount)
	if err != nil {
		fmt.Println("Error adding ibc:", err)
		panic("Stop processing")
	}
	p.totalAffectedAccounts = p.totalAffectedAccounts + 1
	fmt.Printf(`Accounts affected:` + strconv.FormatInt(p.totalAffectedAccounts, 10))
}

func (p *Process) DoWork(accounts []string) {
	for _, address := range accounts {
		// fmt.Println("Processing account", address)
		fmt.Printf(".")
		// Get the sequence number, if it's 0 continue processing
		accountPreRes := GetAccount(address, PreHeight)
		if accountPreRes.Account.BaseAccount.Sequence == "0" {
			balancePreRes := GetEvmosBalance(address, PreHeight)
			// If the evmos balance is dust we can ignore this account because it a valid case for the clawback
			// If the evmos balance is 0 we can ignore because there is nothing to clawback
			if balancePreRes.Balance.Amount == dust || balancePreRes.Balance.Amount == "0" {
				continue
			}

			// Make sure that the balance was moved
			balancePostRes := GetEvmosBalance(address, PostHeight)
			if balancePostRes.Balance.Amount != "0" {
				fmt.Println("This account didnt get its balance moved!", address)
				panic("Stop execution")
			}

			p.InsertBalances(address, balancePreRes.Balance.Amount)
		}
	}
}
