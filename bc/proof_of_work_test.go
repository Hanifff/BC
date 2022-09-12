package main

import (
	"bytes"
	"fmt"
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
)

func newMockHeader(prevBlockHash []byte, merkleRoot []byte) []byte {
	return bytes.Join(
		[][]byte{
			prevBlockHash,
			merkleRoot,
			IntToHex(TestBlockTime),
			IntToHex(TARGETBITS),
		},
		[]byte{},
	)
}

// TARGETBITS == 8 => target difficulty of 2^248
// Hexadecimal: 100000000000000000000000000000000000000000000000000000000000000
// Big Int: 452312848583266388373324160190187140051835877600158453279131187530910662656
var testTargetDifficulty, _ = new(big.Int).SetString("452312848583266388373324160190187140051835877600158453279131187530910662656", 10)

func TestNewProofOfWork(t *testing.T) {
	b := &Block{
		Timestamp:    TestBlockTime,
		Transactions: []*Transaction{testTransactions["tx0"]},
	}

	pow := NewProofOfWork(b)

	assert.Equal(t, testTargetDifficulty, pow.target)
	diff(t, b, pow.block, "wrong block mined")
}

func TestSetupHeader(t *testing.T) {
	pow := &ProofOfWork{
		block: &Block{
			Timestamp:    TestBlockTime,
			Transactions: []*Transaction{testTransactions["tx0"]},
		},
		target: testTargetDifficulty,
	}
	header := pow.setupHeader()

	expectedHeader := newMockHeader(nil, Hex2Bytes("fdfa9ad1db072757d55c11ba05aecae0bbd99e29b8dc2a869a68ebeb1ca09147"))
	assert.Equalf(t, expectedHeader, header, "The current block header: %x isn't equal to the expected %x\n", header, expectedHeader)
}

func TestAddNonce(t *testing.T) {
	header := newMockHeader(nil, Hex2Bytes("fdfa9ad1db072757d55c11ba05aecae0bbd99e29b8dc2a869a68ebeb1ca09147"))
	expectedHeader := Hex2Bytes("fdfa9ad1db072757d55c11ba05aecae0bbd99e29b8dc2a869a68ebeb1ca09147000000005d372e8c00000000000000080000000000000009")

	diff(t, expectedHeader, addNonce(9, header), "addNonce failed")
}

func TestRun(t *testing.T) {
	for k, block := range testBlockchainData {
		t.Run(k, func(t *testing.T) {
			b := &Block{
				Timestamp:     TestBlockTime,
				Transactions:  block.Transactions,
				PrevBlockHash: block.PrevBlockHash,
			}
			pow := &ProofOfWork{b, testTargetDifficulty}
			nonce, hash := pow.Run()
			diff(t, testBlockchainData[k].Nonce, nonce, fmt.Sprintf("wrong nonce for %s", k))
			diff(t, testBlockchainData[k].Hash, hash, fmt.Sprintf("wrong hash for %s", k))
		})
	}
}

func TestValidatePoW(t *testing.T) {
	for k, block := range testBlockchainData {
		t.Run(k, func(t *testing.T) {
			pow := &ProofOfWork{block, testTargetDifficulty}
			assert.True(t, pow.Validate())
		})
	}
}
