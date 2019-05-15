package main

import (
	"blockchain"
	"encoding/json"
	"github.com/gorilla/mux"
	"io/ioutil"
	"log"
	"net/http"
)

func initRouter() *mux.Router{
	router := mux.NewRouter()
	// System routes
	router.HandleFunc("/blockchain/checkActivity", checkActivity)
	// CRUD routes
	router.HandleFunc("/blockchain/sumBlock", sumBlock).Methods("POST")
	router.HandleFunc("/blockchain/addBlock", addExecBlock).Methods("POST")

	// For test routes
	router.HandleFunc("/blockchain/printBlockChain", printBlockchain)
	return router
}

func addExecBlock(w http.ResponseWriter, r *http.Request) {
	if !bc.Status{
		w.WriteHeader(http.StatusLocked)
		return
	}
	body, err := ioutil.ReadAll(r.Body)
	if err != nil{
		log.Printf("Error in read request body\nBody: %s\nError: %s\n", body, err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	var block blockchain.Block
	err = json.Unmarshal(body, block)
	if err != nil{
		log.Printf("Error format blocks!\nBody: %s\nError: %s", body, err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	err = bc.AddBlockWithOutSum(&block)
	if err != nil{
		log.Printf("Error in add block!\nBody: %s\nError: %s", body, err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusOK)
	return
}

func sumBlock(w http.ResponseWriter, r *http.Request) {
	if !bc.Status{
		w.WriteHeader(http.StatusLocked)
		return
	}
	body, err := ioutil.ReadAll(r.Body)
	if err != nil{
		log.Printf("Error in read request body\nBody: %s\nError: %s\n", body, err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	var fb Feedback
	err = json.Unmarshal(body, fb)
	if err != nil{
		log.Printf("Error format feedbacks!\nBody: %s\nError: %s", body, err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	hash := blockchain.NewBlock(fb.Text, fb.IdEmployee, fb.Mark, bc.Tip, fb.TimeStamp).Hash
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(hash)
	return
}

func printBlockchain(w http.ResponseWriter, _ *http.Request) {
	bc.PrintBlockChain(bc.Iterator(), &w)
}

func checkActivity(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(200)
	body, err := ioutil.ReadAll(r.Body)
	if err != nil{
		log.Printf("Error in connection node: %s\nRequest Body: %s", err, body)
	}
	log.Printf("Node %s came here!\n", body)
	return
}