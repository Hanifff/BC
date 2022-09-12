package main

import (
	"bytes"
	"crypto/sha256"
	"math"
	"math/big"
)

var maxNonce = math.MaxInt64

// TARGETBITS define the mining difficulty
const TARGETBITS = 8

// ProofOfWork represents a block mined with a target difficulty
type ProofOfWork struct {
	block  *Block
	target *big.Int
}

// NewProofOfWork builds a ProofOfWork
func NewProofOfWork(block *Block) *ProofOfWork {
	target := big.NewInt(0).Exp(big.NewInt(2), big.NewInt(248), nil)
	return &ProofOfWork{block: block, target: target}
}

// setupHeader prepare the header of the block
func (pow *ProofOfWork) setupHeader() []byte {
	var header []byte
	timeStamp := IntToHex(pow.block.Timestamp)
	tb := IntToHex(TARGETBITS)
	header = append(header, pow.block.PrevBlockHash...)
	header = append(header, pow.block.HashTransactions()...)
	header = append(header, timeStamp...)
	header = append(header, tb...)
	return header
}

// addNonce adds a nonce to the header
func addNonce(nonce int, header []byte) []byte {
	n := IntToHex(int64(nonce))
	header = append(header, n...)
	return header
}

// Run performs the proof-of-work
func (pow *ProofOfWork) Run() (int, []byte) {
	header := pow.setupHeader()
	nonce := 0
	for nonce <= maxNonce {
		headerNonced := addNonce(nonce, header)
		hashHeader := sha256.Sum256(headerNonced)
		bigIntHash := toBigInt(hashHeader[:])
		compare := bigIntHash.Cmp(pow.target)
		if compare == -1 {
			return nonce, hashHeader[:]
		}
		nonce += 1
	}
	return 0, nil
}

// Validate validates block's Proof-Of-Work
// This function just validates if the block header hash
// is less than the target AND equals to the mined block hash.
func (pow *ProofOfWork) Validate() bool {
	oldNonce := pow.block.Nonce
	unMinedBh := pow.block.Hash
	bigIntHash := toBigInt(unMinedBh)
	hashEqual := bigIntHash.Cmp(pow.target)
	pow.block.Mine()
	newNonce := pow.block.Nonce
	afterMine := pow.block.Hash
	minedEqual := bytes.Equal(pow.block.Hash, afterMine)
	return hashEqual == -1 && minedEqual && oldNonce == newNonce
}

func toBigInt(hashHeader []byte) *big.Int {
	hIntoBigInt := big.NewInt(0)
	for _, h := range hashHeader {
		hIntoBigInt.Lsh(hIntoBigInt, 8)
		hIntoBigInt.Add(hIntoBigInt, big.NewInt(int64(h)))
	}
	return hIntoBigInt
}
