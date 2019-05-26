package blockchain

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"strings"
)

type Node struct {
	ip     string
	port   int
	length int
	hash   [32]byte
	status bool
}

func GetNodeList() []Node {
	return []Node{
		{
			ip:     "localhost",
			port:   8080,
			length: 0,
			hash:   [32]byte{},
			status: false,
		},
		{
			ip:     "127.0.0.1",
			port:   8083,
			length: 0,
			hash:   [32]byte{},
			status: false,
		},
	}
}

func (bc *BlockChain) CheckNodesLive(list []Node) (bool, error) {
	client := &http.Client{}
	var countActiveNode int
	var noMatchNode int

	for i := range list {
		stringURL := fmt.Sprintf("http://%s:%d/blockchain/checkActivity", list[i].ip, list[i].port)
		addrNode := fmt.Sprintf("%s:%d", "localhost", 8082)
		req, err := http.NewRequest("GET", stringURL, strings.NewReader(addrNode))
		if err != nil {
			return false, err
		}
		resp, err := client.Do(req)
		if err != nil {
			list[i].status = false
			log.Printf("Node: %s:%d don`t active!\n", list[i].ip, list[i].port)
			continue
		}
		if resp.StatusCode == http.StatusOK {
			list[i].status = true
			body, _ := ioutil.ReadAll(resp.Body)
			responseStruct := struct {
				Length int      `json:"length"`
				Hash   [32]byte `json:"hash"`
			}{}
			err = json.Unmarshal(body, &responseStruct)
			if err != nil {
				log.Printf("Node: %s:%d\nError: %s\n", list[i].ip, list[i].port, err)
				return false, nil
			}
			list[i].length = responseStruct.Length
			list[i].hash = responseStruct.Hash
			countActiveNode++
			if bc.length != list[i].length || !bytes.Equal(bc.hash[:], list[i].hash[:]) {
				noMatchNode++
				log.Printf("Node: %s:%d No match len or hash\n", list[i].ip, list[i].port)
			}
			log.Printf("Node: %s:%d have status OK!\n", list[i].ip, list[i].port)
		}
	}
	if noMatchNode > int(math.Round(float64(len(list))/2.0)) {
		bc.matched = false
		return false, nil
	} else {
		bc.matched = true
	}

	if countActiveNode < 2 || countActiveNode%2 != 0 {
		log.Printf("Not enough nodes in Network, must be 3 and more!\n")
		return false, nil
	}
	return true, nil
}
