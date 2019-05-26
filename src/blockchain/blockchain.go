package blockchain

import (
	"bytes"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"time"
)

type BlockChain struct {
	Tip      []byte
	db       *sql.DB
	length   int
	Status   bool
	hash     []byte
	ListNode []Node
}

type Iterator struct {
	currentHash []byte
	db          *sql.DB
}

func (bc *BlockChain) AddBlock(data string, employeeId int, mark int, timestamp int64) {
	prevFeedBackHash := bc.Tip
	newFeedBack := NewBlock(data, employeeId, mark, prevFeedBackHash, timestamp)

	_, err := bc.db.Exec("INSERT INTO feedbacks(hash, prev_hash, nonce, timestamp, data, id_employee, mark) VALUES (?,?,?,?,?,?,?)", newFeedBack.Hash, newFeedBack.PrevFeedBackHash, newFeedBack.Nonce, newFeedBack.TimeStamp, newFeedBack.Data, newFeedBack.EmployeeId, newFeedBack.Mark)
	if err != nil {
		log.Printf("Error in add block to blockchain: %s", err)
	}
	bc.Tip = newFeedBack.Hash
	bc.length++
}

func (bc *BlockChain) AddBlockWithOutSum(block *Block) error {
	_, err := bc.db.Exec("INSERT INTO feedbacks(hash, prev_hash, nonce, timestamp, data, id_employee, mark) VALUES (?,?,?,?,?,?,?)", block.Hash, block.PrevFeedBackHash, block.Nonce, block.TimeStamp, block.Data, block.EmployeeId, block.Mark)
	if err != nil {
		return err
	}
	bc.Tip = block.Hash
	bc.length++
	return nil
}

func (bc *BlockChain) Iterator() *Iterator {
	return &Iterator{bc.Tip, bc.db}
}

func (bci *Iterator) Next() (*Block, error) {
	var fb Block
	err := bci.db.QueryRow("SELECT hash, prev_hash, nonce, timestamp, data, id_employee, mark FROM feedbacks WHERE hash = ?", bci.currentHash).Scan(&fb.Hash, &fb.PrevFeedBackHash, &fb.Nonce, &fb.TimeStamp, &fb.Data, &fb.EmployeeId, &fb.Mark)
	bci.currentHash = fb.PrevFeedBackHash
	return &fb, err
}

func InitBlockChain(db *sql.DB) (*BlockChain, error) {
	var tip []byte
	var length int
	var bc BlockChain

	_, err := db.Exec("CREATE TABLE IF NOT EXISTS feedbacks(hash VARBINARY(256) PRIMARY KEY, prev_hash VARBINARY(256), nonce INT, timestamp INT, data TEXT, id_employee INT, mark INT);")
	if err != nil {
		return &BlockChain{}, err
	}

	err = db.QueryRow("SELECT hash, COUNT(hash) FROM feedbacks WHERE timestamp = (SELECT max(timestamp) FROM feedbacks);").Scan(&tip, &length)
	if err != nil {
		log.Printf("Error in initialization blockchain: %s", err)
	}
	if len(tip) == 0 {
		genesis := NewBlock("Genesis", -1, 0, []byte{}, 1)
		_, err := db.Exec("INSERT INTO feedbacks(hash, prev_hash, nonce, timestamp, data, id_employee, mark) VALUES (?,?,?,?,?,?,?)", genesis.Hash, genesis.PrevFeedBackHash, genesis.Nonce, genesis.TimeStamp, genesis.Data, genesis.EmployeeId, genesis.Mark)
		if err != nil {
			return &BlockChain{}, err
		}
		tip = genesis.Hash
		length = 1
	}
	bc.Tip = tip
	bc.db = db
	bc.length = length
	bc.Status = false
	bc.hash, err = bc.GetFinalHash()
	if err != nil {
		return &BlockChain{}, err
	}
	bc.ListNode = GetNodeList()
	go bc.CheckNodesInNetWork()
	return &bc, nil
}

func (bc *BlockChain) PrintBlockChain(bci *Iterator, w *http.ResponseWriter) {
	for {
		fb, err := bci.Next()
		if len(fb.Hash) == 0 || err != nil {
			break
		}
		fb.printFeedBack(w)
	}
}

func (bc *BlockChain) GetLength() int {
	return bc.length
}

func (bc *BlockChain) GetHash() []byte {
	return bc.hash
}

func (bc *BlockChain) GetFinalHash() ([]byte, error) {
	var finalHash []byte
	bci := bc.Iterator()
	for {
		fb, err := bci.Next()
		if len(fb.Hash) == 0 {
			break
		}
		if err != nil {
			return []byte{}, err
		}

		finalHash = bytes.Join([][]byte{finalHash, fb.Hash}, []byte{})
		fmt.Printf("finalhash: %x\n", finalHash)
	}
	return finalHash, nil
}

func (bc *BlockChain) CheckNodesInNetWork() {
	for true {
		for bc.Status != true {
			var err error
			time.Sleep(10 * time.Second)
			bc.Status, err = CheckNodesLive(bc.ListNode)
			if err != nil {
				log.Printf("Error in CheckNodesLive:\n%s\n", err)
				return
			}
		}
		log.Printf("BlockChain is work!")
		time.Sleep(60 * time.Second)
	}
}

func (bc *BlockChain) CheckBlockChain(bci *Iterator) (bool, error) {
	for {
		fb, err := bci.Next()
		if len(fb.Hash) == 0 {
			break
		}
		if err != nil {
			return false, err
		}
		pow := NewProofOfWork(fb)
		if !pow.Validate() {
			return false, nil
		}
	}
	return true, nil
}
