package blockchain

import (
	"bytes"
	"crypto/sha256"
	"math"
	"math/big"
	"strconv"
)

const targetBits = 0
const maxNonce = math.MaxInt64

type ProofOfWork struct {
	feedback *Block
	target   *big.Int
}

func NewProofOfWork(fb *Block) *ProofOfWork {
	target := big.NewInt(1)
	target.Lsh(target, uint(256-targetBits))

	pow := &ProofOfWork{fb, target}
	return pow
}

func (pow *ProofOfWork) prepareData(nonce int) []byte {
	timestampHex := []byte(strconv.FormatInt(pow.feedback.TimeStamp, 16))
	targetBitsHex := []byte(strconv.FormatInt(targetBits, 16))
	nonceHex := []byte(strconv.FormatInt(int64(nonce), 16))
	employeeIdHex := []byte(strconv.FormatInt(int64(pow.feedback.EmployeeId), 16))
	data := bytes.Join([][]byte{pow.feedback.PrevFeedBackHash, pow.feedback.Data, employeeIdHex, timestampHex, targetBitsHex, nonceHex}, []byte{})
	return data
}

func (pow *ProofOfWork) Run() (int, []byte) {
	var hashInt big.Int
	var hash [32]byte
	nonce := 0
	for nonce < maxNonce {
		data := pow.prepareData(nonce)
		hash = sha256.Sum256(data)
		hashInt.SetBytes(hash[:])

		if hashInt.Cmp(pow.target) == -1 {
			break
		} else {
			nonce++
		}
	}
	return nonce, hash[:]
}

func (pow *ProofOfWork) Validate() bool {
	var hashInt big.Int

	data := pow.prepareData(pow.feedback.Nonce)
	hash := sha256.Sum256(data)
	hashInt.SetBytes(hash[:])

	isValid := hashInt.Cmp(pow.target) == -1

	return isValid
}
