package main

import (
	"fmt"
	"math/big"
	"os"

	"github.com/hanchon/claims-fixer/internal"
)

var dust = "1000000000000000"

func main() {
	if len(os.Args) == 1 {
		fmt.Println("Use init or process")
		return
	}
	if os.Args[1] == "process" {
		dbToRead := internal.OpenDatabase("./accounts_with_claims.go")
		defer dbToRead.Close()
		fmt.Println("Database opened")

		db := internal.CreateDatabase("./results.go")
		defer db.Close()
		fmt.Println("Database initialized")

		txClaims, insertClaims := internal.CreateInsertClaimsQuery(db)
		defer insertClaims.Close()

		txBalances, insertBalances := internal.CreateInsertBalanceQuery(db)
		defer insertBalances.Close()

		// For each account get its info
		rows, err := dbToRead.Query("select id, address from claims order by id")
		if err != nil {
			fmt.Println("Error reading addresses", err)
			return
		}
		defer rows.Close()

		for rows.Next() {
			var id int
			var address string
			err := rows.Scan(&id, &address)
			if err != nil {
				fmt.Println("Error getting row!", err)
				return
			}

			// Get the sequence number, if it's 0 continue processing
			accountPreRes := internal.GetAccount(address, internal.PreHeight)
			if accountPreRes.Account.BaseAccount.Sequence == "0" {
				balancePreRes := internal.GetBalances(address, internal.PreHeight)
				balancePostRes := internal.GetBalances(address, internal.PostHeight)
				// Make sure that the account still exists
				accountPostRes := internal.GetAccount(address, internal.PostHeight)

				// Data to store
				ibcBalance := make(map[string][]string)
				preBalance := "-1"
				postBalance := "-1"
				affected := false

				for _, k := range balancePreRes.Balances {
					if k.Denom == "aevmos" {
						preBalance = k.Denom
					} else {
						ibcBalance[k.Denom] = []string{k.Amount, "0"}
					}
				}

				for _, k := range balancePostRes.Balances {
					if k.Denom == "aevmos" {
						postBalance = k.Denom
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

				// Store the data
				_, err = insertClaims.Exec(address, preBalance, postBalance, accountPreRes.Account.BaseAccount.Sequence, accountPostRes.Account.BaseAccount.Sequence, affected)
				if err != nil {
					fmt.Println("Error adding:", err)
					return
				}

				for k, v := range ibcBalance {
					_, err = insertBalances.Exec(address, k, v[0], v[1])
					if err != nil {
						fmt.Println("Error adding ibc:", err)
						return
					}
				}

			}

		}

		// Commit
		err = txClaims.Commit()
		if err != nil {
			fmt.Printf("Error commiting transaction: %q", err)
			return
		}
		err = txBalances.Commit()
		if err != nil {
			fmt.Printf("Error commiting transaction: %q", err)
			return
		}
		fmt.Println("All addresses added")

	} else {
		fmt.Println("Invalid option")
	}
}
