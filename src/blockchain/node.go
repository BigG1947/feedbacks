package blockchain

import (
	"fmt"
	"log"
	"net/http"
	"strings"
)

type Node struct {
	ip     string
	port   int
	length int
	hash   []byte
	status bool
}

func GetNodeList() []Node {
	return []Node{
		{
			ip:     "localhost",
			port:   8080,
			length: 0,
			hash:   []byte{},
			status: false,
		},
		{
			ip:     "127.0.0.1",
			port:   8083,
			hash:   []byte{},
			length: 0,
			status: false,
		},
	}
}

func CheckNodesLive(list []Node) (bool, error) {
	client := &http.Client{}
	var countActiveNode int
	for i, _ := range list {
		stringURL := fmt.Sprintf("http://%s:%d/blockchain/checkActivity", list[i].ip, list[i].port)
		addrNode := fmt.Sprintf("%s:%d", "127.0.0.1", 8082)
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
			countActiveNode++
			log.Printf("Node: %s:%d have status OK!\n", list[i].ip, list[i].port)
		}
	}

	if countActiveNode%2 != 0 || countActiveNode < 2 {
		return false, nil
	}
	return true, nil
}
