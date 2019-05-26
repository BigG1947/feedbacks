package main

import (
	"blockchain"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
)

func initRouter() *mux.Router {
	router := mux.NewRouter()
	// System routes
	router.HandleFunc("/blockchain/checkActivity", checkActivity)
	// CRUD routes
	router.HandleFunc("/blockchain/sumBlock", sumBlock).Methods("POST")
	router.HandleFunc("/blockchain/getValidBlockChain", getValidBlockChain)
	router.HandleFunc("/blockchain/addBlockWithoutSum", addBlockWithoutSum).Methods("POST")
	router.HandleFunc("/blockchain/addBlock", addFeedback).Methods("POST")

	// For test routes
	router.HandleFunc("/blockchain/printBlockChain", printBlockchain)
	return router
}

func getValidBlockChain(w http.ResponseWriter, r *http.Request) {
	bci := bc.Iterator()
	var blocks []blockchain.Block
	for {
		fb, _ := bci.Next()
		if len(fb.Hash) == 0 {
			break
		}
		blocks = append(blocks, *fb)
	}
	jsonBlockChain, err := json.Marshal(blocks)
	fmt.Printf("%s", jsonBlockChain)
	if err != nil {
		log.Printf("Error in sendValidBlockChain: %s\n", err)
		w.WriteHeader(500)
		return
	}
	w.WriteHeader(200)
	w.Write(jsonBlockChain)
	return
}

func addBlockWithoutSum(w http.ResponseWriter, r *http.Request) {
	if !bc.Status {
		w.WriteHeader(http.StatusLocked)
		return
	}
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Printf("Error in read request body\nBody: %s\nError: %s\n", body, err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	var block blockchain.Block
	fmt.Printf("Body: %s\n", body)
	err = json.Unmarshal(body, &block)
	fmt.Printf("Block: %v\n", block)
	if err != nil {
		log.Printf("Error format blocks!\nBody: %s\nError: %s\n", body, err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	err = bc.AddBlockWithOutSum(&block)
	if err != nil {
		log.Printf("Error in add block!\nBody: %s\nError: %s\n", body, err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	log.Printf("Block insert into blockchain!\n")
	w.WriteHeader(http.StatusOK)
	return
}

func sumBlock(w http.ResponseWriter, r *http.Request) {
	if !bc.Status {
		w.WriteHeader(http.StatusLocked)
		return
	}
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Printf("Error in read request body\nBody: %s\nError: %s\n", body, err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	var fb Feedback
	err = json.Unmarshal(body, &fb)
	if err != nil {
		log.Printf("Error format feedbacks!\nBody: %s\nError: %s", body, err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	block := blockchain.NewBlock(fb.Text, fb.IdEmployee, fb.Mark, bc.Tip, fb.TimeStamp)
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(block.Hash)
	log.Printf("Block hash sum successed!\n")
	return
}

func printBlockchain(w http.ResponseWriter, _ *http.Request) {
	bc.Print(bc.Iterator(), &w)
}

func checkActivity(w http.ResponseWriter, r *http.Request) {
	responseStruct := struct {
		Length int      `json:"length"`
		Hash   [32]byte `json:"hash"`
	}{
		bc.GetLength(),
		bc.GetHash(),
	}
	response, err := json.Marshal(responseStruct)
	if err != nil {
		log.Printf("Error in marshaling response: %s\n", err)
		return
	}
	w.WriteHeader(200)
	_, err = w.Write(response)
	if err != nil {
		log.Printf("Error in writing response: %s\n", err)
		return
	}
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Printf("Error in connection node: %s\nRequest Body: %s", err, body)
	}
	log.Printf("Node %s came here!\n", body)
	return
}

func addFeedback(w http.ResponseWriter, r *http.Request) {
	idEmployee, err := strconv.Atoi(r.FormValue("id_employee"))
	if err != nil {
		fmt.Fprintf(w, "Ошибка в указании пользователя")
		return
	}
	mark, err := strconv.Atoi(r.FormValue("mark"))
	if err != nil {
		fmt.Fprintf(w, "Ошибка в указании оценки")
		return
	}
	feedback := r.FormValue("text")
	err = bc.AddBlock(feedback, idEmployee, mark, 0)
	if err != nil {
		log.Printf("Error in add block to BlockChain: %s\n", err)
		fmt.Fprint(w, "Система отзывов временно недоступна, приносим вам свои извинения")
		return
	}
	fmt.Fprintf(w, "Ваш отзыв принят в обработку\n")
}
