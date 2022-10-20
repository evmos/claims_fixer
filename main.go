package main

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"

	"github.com/hanchon/claims-fixer/internal"
)

var threads = 16

func main() {
	if len(os.Args) == 1 {
		fmt.Println("Use init or process")
		return
	}

	if os.Args[1] == "init" {
		fmt.Println("Creating accounts database...")
		db := internal.OpenDatabase("./accounts_with_claims.db")
		internal.CreateAccountDatabase(db)
		tx, stmt := internal.CreateInsertAccountQuery(db)

		fmt.Println("Parsing the genesis file...")
		content, err := os.ReadFile("genesis.json")
		if err != nil {
			fmt.Printf("Error reading the genesis: %q", err)
			return
		}

		var genesis internal.Genesis
		err = json.Unmarshal(content, &genesis)
		if err != nil {
			fmt.Println("Error unmarshalling genesis:", err)
			panic("Stop processing")
		}

		fmt.Println("Adding accounts to database...")
		for _, v := range genesis.AppState.Claims.ClaimsRecords {
			_, err := stmt.Exec(v.Address)
			if err != nil {
				fmt.Println("Error adding address:", err)
				panic("Stop processing")
			}
		}

		fmt.Println("Commint changes to database...")
		err = tx.Commit()
		if err != nil {
			fmt.Printf("Error commiting transaction: %q", err)
			panic("Failed to commit tx")
		}
		fmt.Println("All addresses added")

	} else if os.Args[1] == "process" {
		var accountsToProcess []string

		dbToRead := internal.OpenDatabase("./accounts_with_claims.db")
		fmt.Println("Database opened")

		// For each account get its info
		rows, err := dbToRead.Query("select address from account order by id")
		if err != nil {
			fmt.Println("Error reading addresses", err)
			return
		}

		for rows.Next() {
			var address string
			err := rows.Scan(&address)
			if err != nil {
				fmt.Println("Error getting row!", err)
				return
			}
			accountsToProcess = append(accountsToProcess, address)
		}
		fmt.Println("Finished getting all the addresses")
		rows.Close()
		dbToRead.Close()

		// Create the workers
		process := internal.NewProcess()
		var wg sync.WaitGroup
		wg.Add(threads)

		value := len(accountsToProcess) / threads
		j := 0
		offset := 0
		for j < threads {
			if j == threads-1 {
				go func(offset int) {
					process.DoWork(accountsToProcess[offset:])
					wg.Done()
				}(offset)
			} else {
				go func(offset int) {
					process.DoWork(accountsToProcess[offset : offset+value])
					wg.Done()
				}(offset)
			}
			offset = offset + value
			j = j + 1
		}

		wg.Wait()
		process.TxCommit()
		fmt.Println("Everything finished correctly")
	} else {
		fmt.Println("Invalid option")
	}
}
