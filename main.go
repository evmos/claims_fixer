package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
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
	} else if os.Args[1] == "create" {
		fmt.Println("Creating go file for the upgrade...")

		dbToRead := internal.OpenDatabase("./results.db")
		fmt.Println("Database opened")

		// Get all the rows without the ibc account that was fixed via governance
		rows, err := dbToRead.Query("select address, amount from claims where address != \"evmos1a53udazy8ayufvy0s434pfwjcedzqv345dnt3x\" order by address")
		if err != nil {
			fmt.Println("Error reading addresses", err)
			return
		}

		accountsString := ""
		totalRows := 0

		for rows.Next() {
			var address string
			var amount string
			err := rows.Scan(&address, &amount)
			if err != nil {
				fmt.Println("Error getting row!", err)
				return
			}
			accountsString = accountsString + "\t{\"" + address + "\", \"" + amount + "\"},\n"
			totalRows++
		}

		fmt.Println("Finished getting all the addresses")
		rows.Close()
		dbToRead.Close()

		header := "package v9\n\nvar Accounts = [" + strconv.FormatInt(int64(totalRows), 10) + "][2]string{\n"
		data := header + accountsString + "}\n"
		// NOTE: we are not creating a .go file, because it will break this project build
		err = os.WriteFile("accounts.txt", []byte(data), 0644)
		if err != nil {
			fmt.Println("Error on write file:", err)
			panic("Stop execution")
		}
		fmt.Println("File saved as accounts.txt, rename it to .go when moving it to the evmos repo!")

	} else {
		fmt.Println("Invalid option")
	}
}
