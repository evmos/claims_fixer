package internal

import (
	"database/sql"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
)

func OpenDatabase(path string) *sql.DB {
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		fmt.Printf("Error creating/opening database: %q", err)
		panic("Stop processing")
	}
	return db
}

func CreateAccountDatabase(db *sql.DB) {
	sqlStmt := `
       create table account (
        id integer not null primary key,
        address text unique
    );`
	_, err := db.Exec(sqlStmt)
	if err != nil {
		fmt.Printf("Error executing the table creation: %q", err)
		panic("Stop processing")
	}
}

func CreateInsertAccountQuery(db *sql.DB) (*sql.Tx, *sql.Stmt) {
	tx, err := db.Begin()
	if err != nil {
		fmt.Printf("Error creating transaction: %q", err)
		panic("Stop proceesing")
	}
	insertAccount, err := tx.Prepare("insert into account(address) values(?)")
	if err != nil {
		fmt.Printf("Error preparing transaction: %q", err)
		panic("Stop proceesing")
	}
	return tx, insertAccount
}

func CreateDatabase(path string) *sql.DB {
	db := OpenDatabase(path)

	sqlStmt := `
       create table claims (
        id integer not null primary key,
        address text unique,
        amount text
    );
       `

	_, err := db.Exec(sqlStmt)
	if err != nil {
		fmt.Printf("Error executing the table creation: %q", err)
		panic("Stop processing")
	}

	return db
}

func CreateInsertBalanceQuery(db *sql.DB) (*sql.Tx, *sql.Stmt) {
	tx, err := db.Begin()
	if err != nil {
		fmt.Printf("Error creating transaction: %q", err)
		panic("Stop proceesing")
	}

	insertBalances, err := tx.Prepare("insert into claims(address, amount) values(?,?)")
	if err != nil {
		fmt.Printf("Error preparing transaction: %q", err)
		panic("Stop proceesing")
	}
	return tx, insertBalances
}
