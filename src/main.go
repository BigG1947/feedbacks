package main

import (
	"blockchain"
	"database/sql"
	"log"
	"net/http"
)

var db *sql.DB
var bc *blockchain.BlockChain

func main() {
	var err error
	db, err = initDB()
	if err != nil {
		log.Printf("Error in initialization to db:\n%s\n", err)
		return
	}

	bc, err = blockchain.InitBlockChain(db)
	if err != nil {
		log.Printf("Error in initialization blockchain:\n%s\n", err)
	}

	log.Print("Server is starting...")
	log.Print(http.ListenAndServe(":8082", initRouter()))
}
