package main

import (
	"fmt"
	"os"
	"sync"

	"github.com/hanchon/claims-fixer/internal"
)

var dust = "1000000000000000"
var threads = 16

func main() {
	if len(os.Args) == 1 {
		fmt.Println("Use init or process")
		return
	}
	if os.Args[1] == "process" {
		var accountsToProcess []string

		dbToRead := internal.OpenDatabase("./accounts_with_claims.db")
		fmt.Println("Database opened")

		// For each account get its info
		rows, err := dbToRead.Query("select address from claims order by id")
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
		fmt.Println(value)
		j := 0
		offset := 0
		for j < threads {
			if j == threads-1 {
				go func(offset int) {
					fmt.Println(offset)
					process.DoWork(accountsToProcess[offset:])
					wg.Done()
				}(offset)
			} else {
				go func(offset int) {
					fmt.Println(offset)
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
