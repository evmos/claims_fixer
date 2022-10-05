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

func CreateDatabase(path string) *sql.DB {
	db := OpenDatabase(path)

	sqlStmt := `
       create table claims (
        id integer not null primary key,
        address text unique,
        prebalance text,
        postbalance text,
        presequence text,
        postsequence text,
        affected bool
    );
       `

	_, err := db.Exec(sqlStmt)
	if err != nil {
		fmt.Printf("Error executing the table creation: %q", err)
		panic("Stop processing")
	}

	sqlStmtBalances := `
       create table balances (
        id integer not null primary key,
        address text,
        denom text,
        preamount text,
        postamount text
    );
       `

	_, err = db.Exec(sqlStmtBalances)
	if err != nil {
		fmt.Printf("Error executing the table creation: %q", err)
		panic("Stop processing")
	}

	return db
}

func CreateInsertBalanceQuery(db *sql.DB, tx *sql.Tx)  *sql.Stmt {
	insertBalances, err := tx.Prepare("insert into balances(address, denom, preamount, postamount) values(?,?,?,?)")
	if err != nil {
		fmt.Printf("Error preparing transaction: %q", err)
		panic("Stop proceesing")
	}
	return insertBalances
}

func CreateInsertClaimsQuery(db *sql.DB) (*sql.Tx, *sql.Stmt) {
	tx, err := db.Begin()
	if err != nil {
		fmt.Printf("Error creating transaction: %q", err)
		panic("Stop proceesing")
	}

	insertClaims, err := tx.Prepare("insert into claims(address, prebalance, postbalance, presequence, postsequence, affected) values(?,?,?,?,?,?)")
	if err != nil {
		fmt.Printf("Error preparing transaction: %q", err)
		panic("Stop proceesing")
	}
	return tx, insertClaims
}
