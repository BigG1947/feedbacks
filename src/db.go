package main

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
)

func initDB() (*sql.DB, error){
	db, err := sql.Open("sqlite3", "blockchain.db")
	return db, err
}
