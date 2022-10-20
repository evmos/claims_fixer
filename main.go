package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
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
		// Get genesis wallets
		genesisWallets := make(map[string]bool)

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

		fmt.Println("Adding accounts to database from genesis...")
		for _, v := range genesis.AppState.Claims.ClaimsRecords {
			genesisWallets[v.Address] = true
		}

		fmt.Println("Adding accounts to database from the clawback block...")
		// Get all the wallets in the block_result for the clawback block
		content, err = os.ReadFile("evmos_mainnet_5074187_block-results.json")
		if err != nil {
			fmt.Printf("Error reading the genesis: %q", err)
			return
		}

		var m internal.BlockResult
		err = json.Unmarshal(content, &m)
		if err != nil {
			panic("fail unmarshal")
		}

		newWallets := make(map[string]bool)

		for _, v := range m.Result.EndBlockEvents {
			for _, wallet := range v.Attributes {
				if rawDecodedText, err := base64.StdEncoding.DecodeString(wallet.Value); err == nil {
					walletDecoded := string(rawDecodedText)
					// Ignore community wallet because it's the destination
					if walletDecoded == "evmos1jv65s3grqf6v6jl3dp4t6c9t9rk99cd8974jnh" {
						continue
					}
					if strings.Contains(walletDecoded, "evmos1") {
						if _, ok := newWallets[walletDecoded]; ok {
							continue
						}

						if _, ok := genesisWallets[string(rawDecodedText)]; ok == false {
							newWallets[walletDecoded] = true
							_, err := stmt.Exec(walletDecoded)
							if err != nil {
								fmt.Println("Error adding address:", err)
								panic("Stop processing")
							}
						}
					}
				}
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

	} else if os.Args[1] == "addAttestationRecords" {
		// Get genesis wallets
		genesisWallets := make(map[string]bool)

		fmt.Println("Creating accounts database...")
		db := internal.OpenDatabase("./attestation_accounts.db")
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

		for _, v := range genesis.AppState.Claims.ClaimsRecords {
			genesisWallets[v.Address] = true
		}

		fmt.Println("Adding accounts to database from the clawback block...")

		// Get all the wallets in the block_result for the clawback block
		content, err = os.ReadFile("evmos_mainnet_5074187_block-results.json")
		if err != nil {
			fmt.Printf("Error reading the genesis: %q", err)
			return
		}

		var m internal.BlockResult
		err = json.Unmarshal(content, &m)
		if err != nil {
			panic("fail unmarshal")
		}

		newWallets := make(map[string]bool)

		for _, v := range m.Result.EndBlockEvents {
			for _, wallet := range v.Attributes {
				if rawDecodedText, err := base64.StdEncoding.DecodeString(wallet.Value); err == nil {
					walletDecoded := string(rawDecodedText)
					// Ignore community wallet because it's the destination
					if walletDecoded == "evmos1jv65s3grqf6v6jl3dp4t6c9t9rk99cd8974jnh" {
						continue
					}
					if strings.Contains(walletDecoded, "evmos1") {
						if _, ok := newWallets[walletDecoded]; ok {
							continue
						}

						if _, ok := genesisWallets[string(rawDecodedText)]; ok == false {
							newWallets[walletDecoded] = true
							_, err := stmt.Exec(walletDecoded)
							if err != nil {
								fmt.Println("Error adding address:", err)
								panic("Stop processing")
							}
						}
					}
				}
			}
		}

		fmt.Println("Commint changes to database...")
		err = tx.Commit()
		if err != nil {
			fmt.Printf("Error commiting transaction: %q", err)
			panic("Failed to commit tx")
		}
		fmt.Println("All addresses added")

	} else if os.Args[1] == "processAttestationRecords" {
		var accountsToProcess []string

		dbToRead := internal.OpenDatabase("./attestation_accounts.db")
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
