package main

import (
	"blockchain"
	"database/sql"
	"log"
	"net/http"
	"os"
)

var db *sql.DB
var bc *blockchain.BlockChain

func main() {
	var err error
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	f, err := os.OpenFile("error.log", os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		panic(err)
	}
	log.SetOutput(f)

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
