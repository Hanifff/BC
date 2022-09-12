package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewBlock(t *testing.T) {
	genesisBlock := testBlockchainData["block0"]
	b := NewBlock(TestBlockTime, []*Transaction{testTransactions["tx1"]}, genesisBlock.Hash)
	assert.NotNil(t, b)
	assert.Nil(t, b.Hash, "The block hash of non-mined block should be the zero value")
	assert.Equal(t, genesisBlock.Hash, b.PrevBlockHash, "Previous block of the current should be the genesis block")
}

func TestGenesisBlock(t *testing.T) {
	// Genesis block
	gb := NewGenesisBlock(TestBlockTime, testTransactions["tx0"])
	assert.NotNil(t, gb)
	assert.Nil(t, gb.PrevBlockHash, "Genesis block should not have PrevBlockHash")

	assert.Equal(t, testTransactions["tx0"].ID, gb.Transactions[0].ID, "Genesis block should contains the genesis transaction")
}

func TestBlockHashTransactions(t *testing.T) {
	// Merkle root of block1
	merkleRootTxsHash := Hex2Bytes("153b097cb029051834d285f69901237812f13d6dc166d684778096422f331677")
	b := &Block{
		Transactions: []*Transaction{testTransactions["tx1"]},
	}
	root := b.HashTransactions()

	assert.Equalf(t, merkleRootTxsHash, root, "The block hash %x isn't equal to %x", root, merkleRootTxsHash)
}

func TestMine(t *testing.T) {
	genesisBlock := testBlockchainData["block0"]

	b := &Block{
		Timestamp: TestBlockTime,
		Transactions: []*Transaction{
			minerCoinbaseTx["tx1"],
			testTransactions["tx1"],
		},
		PrevBlockHash: genesisBlock.Hash,
	}
	b.Mine()

	assert.Equalf(t, testBlockchainData["block1"].Hash, b.Hash, "The block hash %x isn't equal to %x", b.Hash, testBlockchainData["block1"].Hash)
	assert.Equalf(t, testBlockchainData["block1"].Nonce, b.Nonce, "The block nonce %d isn't equal to %d", b.Nonce, testBlockchainData["block1"].Nonce)
}

func TestFindTransaction(t *testing.T) {
	block := testBlockchainData["block2"]
	expectedTX := testTransactions["tx2"]

	tx, err := block.FindTransaction(expectedTX.ID)
	assert.Nil(t, err)
	if tx == nil {
		t.Fatal("FindTransaction returned nil but expected a tx")
	}
	assert.Equalf(t, expectedTX, tx, "The found tx: %x is not equal to the expected tx: %x", tx.ID, expectedTX.ID)
}
