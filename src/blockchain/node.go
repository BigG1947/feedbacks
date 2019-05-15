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
	status bool
}

func GetNodeList() *[]Node {
	return &[]Node{
		{
			ip:     "localhost",
			port:   8080,
			length: 0,
			status: false,
		},
		{
			ip:     "127.0.0.1",
			port:   8083,
			length: 0,
			status: false,
		},
	}
}

func CheckNodesLive(list *[]Node) (bool, error) {
	client := &http.Client{}
	var countActiveNode int
	for _, node := range *list {
		stringURL := fmt.Sprintf("http://%s:%d/blockchain/checkActivity", node.ip, node.port)
		addrNode := fmt.Sprintf("%s:%d", node.ip, node.port)
		req, err := http.NewRequest("GET", stringURL, strings.NewReader(addrNode))
		if err != nil {
			return false, err
		}
		resp, err := client.Do(req)
		if err != nil {
			node.status = false
			log.Printf("Node: %s:%d don`t active!\n", node.ip, node.port)
			continue
		}
		if resp.StatusCode == http.StatusOK {
			node.status = true
			countActiveNode++
			log.Printf("Node: %s:%d have status OK!\n", node.ip, node.port)
		}
	}

	if countActiveNode%2 != 0 || countActiveNode < 2 {
		return false, nil
	}
	return true, nil
}
