package main

import (
	"bytes"
	"crypto/ecdsa"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"
)

var (
	ErrTxNotFound    = errors.New("transaction not found")
	ErrNoValidTx     = errors.New("there is no valid transaction")
	ErrBlockNotFound = errors.New("block not found")
	ErrInvalidBlock  = errors.New("block is not valid")
)

// Blockchain keeps a sequence of Blocks
type Blockchain struct {
	blocks []*Block
}

// NewBlockchain creates a new blockchain with genesis Block
func NewBlockchain(address string) (*Blockchain, error) {
	txn, err := NewCoinbaseTX(address, GenesisCoinbaseData)
	if err != nil {
		return nil, err
	}
	txn.ID = txn.Hash()
	ts := time.Now().Unix()
	gensisBlock := NewGenesisBlock(ts, txn)
	gensisBlock.Mine()
	return &Blockchain{blocks: []*Block{gensisBlock}}, nil
}

// addBlock saves the block into the blockchain
func (bc *Blockchain) addBlock(block *Block) error {
	if !bc.ValidateBlock(block) {
		return ErrInvalidBlock
	}
	bc.blocks = append(bc.blocks, block)
	return nil
}

// GetGenesisBlock returns the Genesis Block
func (bc Blockchain) GetGenesisBlock() *Block {
	gensisBlock := bc.blocks[0]
	return gensisBlock
}

// CurrentBlock returns the last block
func (bc Blockchain) CurrentBlock() *Block {
	currBlock := bc.blocks[len(bc.blocks)-1]
	return currBlock
}

// GetBlock returns the block of a given hash
func (bc Blockchain) GetBlock(hash []byte) (*Block, error) {
	if len(hash) != 32 || len(bc.blocks) <= 0 {
		return nil, ErrInvalidBlock
	}
	for _, b := range bc.blocks {
		if bytes.Equal(b.Hash, hash) {
			return b, nil
		}
	}
	return nil, ErrBlockNotFound
}

// ValidateBlock validates the block before adding it to the blockchain
func (bc *Blockchain) ValidateBlock(block *Block) bool {
	if block == nil || len(block.Transactions) == 0 {
		return false
	}
	pow := NewProofOfWork(block)
	validPow := pow.Validate()
	return validPow
}

// MineBlock mines a new block with the provided transactions
func (bc *Blockchain) MineBlock(transactions []*Transaction) (*Block, error) {
	var validTxns []*Transaction
	for _, t := range transactions {
		if bc.VerifyTransaction(t) {
			validTxns = append(validTxns, t)
		}
	}
	if len(validTxns) > 0 {
		block := NewBlock(time.Now().Unix(), validTxns, bc.CurrentBlock().Hash)
		block.Mine()
		bc.addBlock(block)
		return block, nil
	}
	return nil, ErrNoValidTx
}

// VerifyTransaction verifies if referred inputs exist
func (bc Blockchain) VerifyTransaction(tx *Transaction) bool {
	if tx.IsCoinbase() {
		return true
	}
	for _, inPuts := range tx.Vin {
		_, err := bc.FindTransaction(inPuts.Txid)
		if err != nil {
			return false
		}
	}
	return true
}

// FindTransaction finds a transaction by its ID in the whole blockchain
func (bc Blockchain) FindTransaction(ID []byte) (*Transaction, error) {
	for _, b := range bc.blocks {
		node, err := b.FindTransaction(ID)
		if err != nil {
			continue
		} else {
			return node, nil
		}
	}
	return nil, ErrTxNotFound
}

// FindUTXOSet finds and returns all unspent transaction outputs
func (bc Blockchain) FindUTXOSet() UTXOSet {
	utxos := make(map[string]map[int]TXOutput)
	for _, b := range bc.blocks { // O(n)
		for _, txn := range b.Transactions { // O(n)
			outMap := make(map[int]TXOutput)
			if txn.IsCoinbase() {
				outMap[0] = txn.Vout[0]
				utxos[fmt.Sprintf("%x", txn.ID)] = outMap
				continue
			}
			for _, in := range txn.Vin { // O(m)
				txID := fmt.Sprintf("%x", in.Txid)
				idx := in.OutIdx
				if _, ok := utxos[txID]; ok {
					delete(utxos[txID], idx)
					if len(utxos[txID]) == 0 {
						delete(utxos, txID)
					}
				}
			}
			for idx, out := range txn.Vout { // O(m)
				outMap[idx] = out
				utxos[fmt.Sprintf("%x", txn.ID)] = outMap
			}
		}
	} // worst case O(n^2*m)
	return utxos
}

// GetInputTXsOf returns a map index by the ID,
// of all transactions used as inputs in the given transaction
func (bc *Blockchain) GetInputTXsOf(tx *Transaction) (map[string]*Transaction, error) {
	prevTxs := make(map[string]*Transaction)
	for _, inp := range tx.Vin {
		tx, err := bc.FindTransaction(inp.Txid)
		if err != nil {
			return nil, ErrTxNotFound
		}
		prevTxs[hex.EncodeToString(inp.Txid)] = tx
	}
	return prevTxs, nil
}

// SignTransaction signs inputs of a Transaction
func (bc *Blockchain) SignTransaction(tx *Transaction, privKey ecdsa.PrivateKey) error {
	prevTXs, err := bc.GetInputTXsOf(tx)
	if err != nil {
		return err
	}
	return tx.Sign(privKey, prevTXs)
}

func (bc Blockchain) String() string {
	var lines []string
	for _, block := range bc.blocks {
		lines = append(lines, fmt.Sprintf("%v", block))
	}
	return strings.Join(lines, "\n")
}
