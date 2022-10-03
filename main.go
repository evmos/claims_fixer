package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

type Balance struct {
	Amount string `json:"amount"`
	Denom  string `json:"denom"`
}
type BalanceResponse struct {
	Balance Balance `json:"balance"`
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

func main() {
	if len(os.Args) == 1 {
		fmt.Println("Use init or process")
		return
	}
	if os.Args[1] == "init" {
		db, err := sql.Open("sqlite3", "./claims.db")
		if err != nil {
			fmt.Printf("Error creating/opening database: %q", err)
			return
		}
		defer db.Close()

		sqlStmt := `
       create table if not exists claims (
        id integer not null primary key,
        address text unique,
        balance text,
        sequence text
    );
       `
		_, err = db.Exec(sqlStmt)
		if err != nil {
			fmt.Printf("Error executing the table creation: %q", err)
			return
		}
		fmt.Println("Database initialized")
		tx, err := db.Begin()
		if err != nil {
			fmt.Printf("Error creating transaction: %q", err)
			return
		}
		stmt, err := tx.Prepare("insert into claims(address) values(?)")
		if err != nil {
			fmt.Printf("Error preparing transaction: %q", err)
			return
		}
		defer stmt.Close()

		// For all the evmos1 addresses
		content, err := os.ReadFile("genesis.json")
		if err != nil {
			fmt.Printf("Error reading the genesis: %q", err)
			return
		}
		r := regexp.MustCompile(`"evmos1[a-z0-9]*"`)
		matches := r.FindAllString(string(content), -1)
		for _, m := range matches {
			fmt.Println("Adding:", m)
			_, err = stmt.Exec(strings.ReplaceAll(m, "\"", ""))
			if err != nil && err.Error() != "UNIQUE constraint failed: claims.address" {
				fmt.Println("Error adding:", m, err)
				return
			}
		}

		// Commit
		err = tx.Commit()
		if err != nil {
			fmt.Printf("Error commiting transaction: %q", err)
			return
		}
		fmt.Println("All addresses added")

	} else if os.Args[1] == "process" {
		// Read accounts from this database
		dbToRead, err := sql.Open("sqlite3", "./claims.db")
		if err != nil {
			fmt.Printf("Error creating/opening database: %q", err)
			return
		}
		defer dbToRead.Close()

		// Init new database
		db, err := sql.Open("sqlite3", "./errors.db")
		if err != nil {
			fmt.Printf("Error creating/opening database: %q", err)
			return
		}
		defer db.Close()

		sqlStmt := `
       create table if not exists claims (
        id integer not null primary key,
        address text,
        balance text,
        sequence text
    );
       `
		_, err = db.Exec(sqlStmt)
		if err != nil {
			fmt.Printf("Error executing the table creation: %q", err)
			return
		}
		fmt.Println("Database initialized")
		tx, err := db.Begin()
		if err != nil {
			fmt.Printf("Error creating transaction: %q", err)
			return
		}
		stmt, err := tx.Prepare("insert into claims(address, balance, sequence) values(?,?,?)")
		if err != nil {
			fmt.Printf("Error preparing transaction: %q", err)
			return
		}
		defer stmt.Close()

		// HttpRequests
		client := &http.Client{}
		fmt.Println("Processing addresses:")
		PRE_HEIGHT := "5074186"
		//POST_HEIGHT := 5074187

		endpoint := "http://localhost:1317/"
		balance_start := "cosmos/bank/v1beta1/balances/"
		balance_end := "/by_denom?denom=aevmos"

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

			fmt.Println("Processing address:", address, id)
			req, _ := http.NewRequest("GET", endpoint+balance_start+address+balance_end, nil)
			req.Header.Set("x-cosmos-block-height", PRE_HEIGHT)
			res, err := client.Do(req)
			if err != nil {
				fmt.Println("Error getting the balance", address, err)
				return
			}
			defer res.Body.Close()
			body, err := ioutil.ReadAll(res.Body)

			m := &BalanceResponse{}
			err = json.Unmarshal(body, &m)
			if err != nil {
				fmt.Println("Error parsing the balance response", address, m)
				fmt.Println("Account with problems:", address, m.Balance.Amount)
				_, err = stmt.Exec(address, "-1", "-1")
				if err != nil {
					fmt.Println("Error adding:", m)
					return
				}
				continue
			}

			// Balance 1000000000000000 == dust to claim
			if m.Balance.Amount != "1000000000000000" {
				// Get the balance sequence

				PRE_HEIGHT := "5074186"
				//POST_HEIGHT := 5074187

				account_start := "cosmos/auth/v1beta1/accounts/"

				req, _ := http.NewRequest("GET", endpoint+account_start+address, nil)
				req.Header.Set("x-cosmos-block-height", PRE_HEIGHT)
				res, err := client.Do(req)
				if err != nil {
					fmt.Println("Error getting the balance", address, err)
					return
				}
				defer res.Body.Close()
				body, err := ioutil.ReadAll(res.Body)

				a := &AccountResponse{}
				err = json.Unmarshal(body, &a)
				if err != nil {
					fmt.Println("Error parsing the account response", address, a)
					return
				}

				if a.Account.BaseAccount.Sequence == "0" {
					fmt.Println("Account with problems:", address, m.Balance.Amount)
					_, err = stmt.Exec(address, m.Balance.Amount, a.Account.BaseAccount.Sequence)
					if err != nil {
						fmt.Println("Error adding:", m)
						return
					}
				}

			}
		}

		// Commit
		err = tx.Commit()
		if err != nil {
			fmt.Printf("Error commiting transaction: %q", err)
			return
		}

	} else {
		fmt.Println("Invalid option")
	}
}
