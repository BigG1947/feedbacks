package blockchain

import (
	"bytes"
	"crypto/sha256"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"
)

type BlockChain struct {
	Tip      []byte
	Status   bool
	db       *sql.DB
	length   int
	listNode []Node
	hash     [32]byte
	broken   bool
	matched  bool
	mutex    *sync.Mutex
}

type Iterator struct {
	currentHash []byte
	db          *sql.DB
}

func (bc *BlockChain) AddBlock(data string, employeeId int, mark int, timestamp int64) error {
	var err error
	if !bc.Status {
		return errors.New("BlockChain status if fall ")
	}
	ok, err := bc.CheckNodesLive(bc.listNode)
	if err != nil {
		return err
	}
	if !ok {
		bc.Status = false
		return errors.New("BlockChain status if fall ")
	}
	prevFeedBackHash := bc.Tip
	newFeedBack := NewBlock(data, employeeId, mark, prevFeedBackHash, timestamp)
	ok, err = bc.compareBlockWithOtherNodes(employeeId, mark, data, newFeedBack.TimeStamp, newFeedBack.Hash)
	if err != nil {
		return err
	}
	if !ok {
		return errors.New("Block Hash not match to Hash in other nodes ")
	}
	ok, err = bc.sendBlockToOtherNodes(newFeedBack)
	if err != nil {
		return err
	}
	if !ok {
		return errors.New("Node returns a BadRequest header ")
	}
	_, err = bc.db.Exec("INSERT INTO feedbacks(hash, prev_hash, nonce, timestamp, data, id_employee, mark) VALUES (?,?,?,?,?,?,?)", newFeedBack.Hash, newFeedBack.PrevFeedBackHash, newFeedBack.Nonce, newFeedBack.TimeStamp, newFeedBack.Data, newFeedBack.EmployeeId, newFeedBack.Mark)
	if err != nil {
		return err
	}
	bc.Tip = newFeedBack.Hash
	bc.length++
	bc.hash = sha256.Sum256(bytes.Join([][]byte{bc.hash[:], newFeedBack.Hash}, []byte{}))
	log.Printf("Block insert into BlockChain!\nBC Lenght: %d\nBC Hash: %x\n", bc.length, bc.hash)
	return nil
}

func (bc *BlockChain) AddBlockWithOutSum(block *Block) error {
	_, err := bc.db.Exec("INSERT INTO feedbacks(hash, prev_hash, nonce, timestamp, data, id_employee, mark) VALUES (?,?,?,?,?,?,?)", block.Hash, block.PrevFeedBackHash, block.Nonce, block.TimeStamp, block.Data, block.EmployeeId, block.Mark)
	if err != nil {
		return err
	}
	bc.Tip = block.Hash
	bc.length++
	bc.hash = sha256.Sum256(bytes.Join([][]byte{bc.hash[:], block.Hash}, []byte{}))
	log.Printf("Block insert into BlockChain!\nBC Lenght: %d\nBC Hash: %x\n", bc.length, bc.hash)
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

	_, err := db.Exec("CREATE TABLE IF NOT EXISTS feedbacks(hash VARBINARY(256) PRIMARY KEY, prev_hash VARBINARY(256), nonce INT, timestamp INT, data TEXT, id_employee INT, mark INT);")
	if err != nil {
		return &BlockChain{}, err
	}

	err = db.QueryRow("SELECT hash FROM feedbacks WHERE timestamp = (SELECT max(timestamp) FROM feedbacks);").Scan(&tip)
	if err != nil && err != sql.ErrNoRows {
		log.Printf("Error in initialization blockchain: %s", err)
	}
	err = db.QueryRow("SELECT COUNT(hash) FROM feedbacks;").Scan(&length)
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
	bc := BlockChain{tip, false, db, length, GetNodeList(), [32]byte{}, false, false, &sync.Mutex{}}
	bc.hash, err = bc.GetFinalHash()
	if err != nil {
		return &BlockChain{}, err
	}
	go bc.CheckNodesInNetWork()
	time.Sleep(1 * time.Second)
	go bc.Check()
	return &bc, nil
}

func (bc *BlockChain) FillTestFeedBack() {
	time.Sleep(5 * time.Second)
	bc.AddBlock("Дуже чудовый выкладач! Зрозуміло виклдає і викликає зацікавленість до предмету", 12, 100, 1570136400)
	bc.AddBlock("Викликає повагу! Ділиться чудовим досвідом про професію та сприяє професійному росту", 13, 99, 1555707600)
	bc.AddBlock("Хорошо, добра та чутлива. Завжди допоможе кожному студенту і нікого не залишить з проблемами", 6, 99, 1555102800)
	bc.AddBlock("Добре розуміє студентів і завжди допомогає їм", 6, 100, 1559377243)
	bc.AddBlock("Веселий та добрий викладач. Завжди з великим задоволенням хожу на її пари", 9, 100, 1549355256)
	bc.AddBlock("Завжди інтересно і зрозуміло викладає матеріал. Допомагає зрозуміти навіть найскладніші моменти", 9, 99, 1556355256)
	bc.AddBlock("Завдяки їй з'явилося бажаня вкладатися у навчання!", 12, 100, 1553185256)

	bc.AddBlock("Дуже чудовый выкладач! Зрозуміло виклдає і викликає зацікавленість до предмету", 6, 100, 1570136400)
	bc.AddBlock("Викликає повагу! Ділиться чудовим досвідом про професію та сприяє професійному росту", 9, 99, 1555707600)
	bc.AddBlock("Хорошо, добра та чутлива. Завжди допоможе кожному студенту і нікого не залишить з проблемами", 12, 99, 1555102800)
	bc.AddBlock("Добре розуміє студентів і завжди допомогає їм", 13, 100, 1559377243)
	bc.AddBlock("Веселий та добрий викладач. Завжди з великим задоволенням хожу на її пари", 16, 100, 1549355256)
	bc.AddBlock("Завжди інтересно і зрозуміло викладає матеріал. Допомагає зрозуміти навіть найскладніші моменти", 12, 99, 1556355256)
	bc.AddBlock("Завдяки їй з'явилося бажаня вкладатися у навчання!", 14, 100, 1553185256)

	bc.AddBlock("Дуже чудовый выкладач! Зрозуміло виклдає і викликає зацікавленість до предмету", 9, 100, 1570136400)
	bc.AddBlock("Викликає повагу! Ділиться чудовим досвідом про професію та сприяє професійному росту", 12, 99, 1555707600)
	bc.AddBlock("Хорошо, добра та чутлива. Завжди допоможе кожному студенту і нікого не залишить з проблемами", 9, 99, 1555102800)
	bc.AddBlock("Добре розуміє студентів і завжди допомогає їм", 6, 100, 1559377243)
	bc.AddBlock("Веселий та добрий викладач. Завжди з великим задоволенням хожу на її пари", 12, 100, 1549355256)
	bc.AddBlock("Завжди інтересно і зрозуміло викладає матеріал. Допомагає зрозуміти навіть найскладніші моменти", 16, 99, 1556355256)
	bc.AddBlock("Завдяки їй з'явилося бажаня вкладатися у навчання!", 16, 100, 1553185256)
}

func (bc *BlockChain) Print(bci *Iterator, w *http.ResponseWriter) {
	for {
		fb, err := bci.Next()
		if fb == nil || len(fb.Hash) == 0 || err != nil {
			break
		}
		fb.printFeedBack(w)
	}
}

func (bc *BlockChain) GetLength() int {
	return bc.length
}

func (bc *BlockChain) GetNodeList() []Node {
	return bc.listNode
}

func (bc *BlockChain) GetFinalHash() ([32]byte, error) {
	var finalHash []byte
	bci := bc.Iterator()
	for {
		fb, err := bci.Next()
		if fb == nil || len(fb.Hash) == 0 {
			break
		}
		if err != nil {
			return [32]byte{}, err
		}

		finalHash = bytes.Join([][]byte{finalHash, fb.Hash}, []byte{})
	}
	return sha256.Sum256(finalHash), nil
}

func (bc *BlockChain) compareBlockWithOtherNodes(idEmployee int, mark int, text string, timestamp int64, hash []byte) (bool, error) {
	client := http.Client{}
	var applyNode int

	var data = struct {
		IdEmployee int    `json:"id_employee"`
		Mark       int    `json:"mark"`
		Text       string `json:"text"`
		TimeStamp  int64  `json:"time"`
	}{
		IdEmployee: idEmployee,
		Mark:       mark,
		Text:       text,
		TimeStamp:  timestamp,
	}
	jsonData, err := json.Marshal(data)
	if err != nil {
		return false, err
	}
	for _, node := range bc.listNode {
		if !node.status {
			continue
		}
		stringURL := fmt.Sprintf("http://%s:%d/blockchain/sumBlock", node.ip, node.port)
		req, err := http.NewRequest("POST", stringURL, bytes.NewBuffer(jsonData))
		if err != nil {
			return false, err
		}
		resp, err := client.Do(req)
		if err != nil {
			return false, err
		}
		if resp.StatusCode == http.StatusBadRequest {
			return false, nil
		}
		body, _ := ioutil.ReadAll(resp.Body)
		if bytes.Equal(body, hash) {
			applyNode++
			continue
		}
	}
	if applyNode == len(bc.listNode) {
		return true, nil
	}
	return false, nil
}

func (bc *BlockChain) sendBlockToOtherNodes(block *Block) (bool, error) {
	client := http.Client{}
	jsonBlock, err := json.Marshal(block)
	if err != nil {
		return false, err
	}
	for _, node := range bc.listNode {
		if !node.status {
			continue
		}
		stringURL := fmt.Sprintf("http://%s:%d/blockchain/addBlockWithoutSum", node.ip, node.port)
		req, err := http.NewRequest("POST", stringURL, bytes.NewBuffer(jsonBlock))
		if err != nil {
			return false, err
		}
		resp, err := client.Do(req)
		if err != nil {
			return false, err
		}
		if resp.StatusCode == http.StatusBadRequest {
			return false, nil
		}
		if resp.StatusCode == http.StatusOK {
			continue
		}
	}
	return true, nil
}

func (bc *BlockChain) CheckNodesInNetWork() {
	defer bc.mutex.Unlock()
	for true {
		bc.mutex.Lock()
		for true {
			var err error
			bc.Status, err = bc.CheckNodesLive(bc.listNode)
			if err != nil {
				log.Printf("Error in CheckNodesLive:\n%s\n", err)
				return
			}
			if bc.matched == false {
				break
			}
			if bc.Status == true && bc.matched == true {
				log.Printf("BlockChain is work!")
				break
			}
			time.Sleep(10 * time.Second)
		}
		bc.mutex.Unlock()
		time.Sleep(30 * time.Second)
	}
}

func (bc *BlockChain) Check() {
	defer bc.mutex.Unlock()
	for true {
		bc.mutex.Lock()
		bci := bc.Iterator()
		bc.broken = false
		for true {
			fb, err := bci.Next()
			if fb == nil || len(fb.Hash) == 0 {
				break
			}
			if err != nil {
				log.Printf("Error in BlockChain.Check method: %s\n", err)
			}
			pow := NewProofOfWork(fb)
			if !pow.Validate() {
				log.Printf("Broken Block:\nHash: %x\nTimestamp: %d\n", fb.Hash, fb.TimeStamp)
				bc.broken = true
				break
			}
		}
		if !bc.broken && bc.matched {
			log.Printf("BlockChain is Ok!\n")
		} else if bc.broken {
			log.Printf("BlockChain is Broken\n")
			bc.repair()
		} else if bc.matched == false {
			log.Printf("BlockChain is not math to other blockChain nodes\n")
			bc.repair()
		}
		bc.mutex.Unlock()
		time.Sleep(30 * time.Second)
	}
}

func (bc *BlockChain) repair() {
	log.Printf("Start repair blockchain...\n")
	mapLengthNode := make(map[int][]Node)
	for _, node := range bc.listNode {
		mapLengthNode[node.length] = append(mapLengthNode[node.length], node)
	}
	var maxLength int
	for i := range mapLengthNode {
		if i > maxLength {
			maxLength = i
		}
	}
	blocks, err := getValidBlockChain(&mapLengthNode[maxLength][0])
	if err != nil {
		log.Printf("Repair blockchain is failed!\nErr: %s", err)
		return
	}
	tx, err := bc.db.Begin()
	if err != nil {
		log.Printf("BlockChain.repair. Error in begin transaction: %s\n", err)
		return
	}
	defer tx.Rollback()
	_, err = bc.db.Exec("DELETE FROM feedbacks;")
	if err != nil {
		log.Printf("Error in delete broken blockchain!\nErr: %s\n", err)
		return
	}
	stmt, err := tx.Prepare("INSERT INTO feedbacks(hash, prev_hash, nonce, timestamp, data, id_employee, mark) VALUES (?,?,?,?,?,?,?)")
	if err != nil {
		log.Printf("BlockChain.repair. Error in prepared query: %s\n", err)
		return
	}
	defer stmt.Close()
	for i := range blocks {
		_, err = stmt.Exec(blocks[i].Hash, blocks[i].PrevFeedBackHash, blocks[i].Nonce, blocks[i].TimeStamp, blocks[i].Data, blocks[i].EmployeeId, blocks[i].Mark)
		if err != nil {
			log.Printf("BlockChain.repair. Error in execute query: %s\n", err)
			return
		}
	}
	err = tx.Commit()
	if err != nil {
		log.Printf("BlockChain.repair. Error in commit transaction: %s\n", err)
		return
	}
	bc.length = mapLengthNode[maxLength][0].length
	bc.hash = mapLengthNode[maxLength][0].hash
	bc.Tip = blocks[0].Hash[:]
	log.Printf("Repair is success!")
}

func (bc *BlockChain) GetHash() [32]byte {
	return bc.hash
}

func getValidBlockChain(node *Node) ([]Block, error) {
	var client http.Client
	backUpBlockChain := struct {
		Blocks []Block `json:"blocks"`
	}{}

	stringURL := fmt.Sprintf("http://%s:%d/blockchain/getValidBlockChain", node.ip, node.port)
	selfAddress := fmt.Sprintf("%s:%d", "localhost", 8082)
	req, err := http.NewRequest("GET", stringURL, strings.NewReader(selfAddress))
	if err != nil {
		return []Block{}, err
	}
	resp, err := client.Do(req)
	if err != nil {
		err = fmt.Errorf("getValidBlockChain: %s:%d node don`t active!\n", node.ip, node.port)
		return []Block{}, err
	}
	if resp.StatusCode != http.StatusOK {
		return []Block{}, fmt.Errorf("Http status code: %d ", resp.StatusCode)
	} else {
		body, _ := ioutil.ReadAll(resp.Body)
		err = json.Unmarshal(body, &backUpBlockChain.Blocks)
		if err != nil {
			err = fmt.Errorf("getValidBlockChain: %s:%d node\nError Unmarshaling: %s\n", node.ip, node.port, err)
			return []Block{}, err
		}
	}
	return backUpBlockChain.Blocks, nil
}
