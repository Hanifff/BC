package main

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFindSpendableOutputsFromOneOutput(t *testing.T) {
	utxos := getTestExpectedUTXOSet("block0")
	expectedOut := utxos["9402c56f49de02d2b9c4633837d82e3881227a3ea90c4073c02815fdcf5afaa2"]
	expectedValue := expectedOut[0].Value
	pubKeyHash := expectedOut[0].PubKeyHash
	expectedUnspentOutputs := getTestSpendableOutputs(utxos, pubKeyHash)

	// Find the 14vRYoWsjqC61tNmaLPPzjKnxirSxFoehh unspent TXOutputs
	accumulatedAmount, unspentOutputs := utxos.FindSpendableOutputs(pubKeyHash, 5)

	assert.Equal(t, expectedValue, accumulatedAmount)
	assert.Equal(t, expectedUnspentOutputs, unspentOutputs)
}

func TestFindSpendableOutputsFromMultipleOutputs(t *testing.T) {
	utxos := getTestExpectedUTXOSet("block2")
	out1 := utxos["e9e5fc159f24b2b33310f77aef4e425e77ed71be87dbf9a0c7764b5417bd3e4b"]
	out2 := utxos["dcd76d254f7a41888e6bda9958c4ceadf510e1bd5fd251f617c91b704fbf9492"]
	expectedValue := out1[1].Value + out2[0].Value

	expectedUnspentOutputs := getTestSpendableOutputs(utxos, out1[1].PubKeyHash)

	// Find the 14vRYoWsjqC61tNmaLPPzjKnxirSxFoehh unspent TXOutputs
	accumulatedAmount, unspentOutputs := utxos.FindSpendableOutputs(out1[1].PubKeyHash, 5)

	assert.Equal(t, expectedValue, accumulatedAmount)
	assert.Equal(t, expectedUnspentOutputs, unspentOutputs)
}

func TestFindUTXO(t *testing.T) {
	// 14vRYoWsjqC61tNmaLPPzjKnxirSxFoehh create a coinbase transaction, receiving 10 "coins"
	utxos := getTestExpectedUTXOSet("block0")

	rodrigoPubKeyHash := Hex2Bytes("2b02ea4c157844ec0b034fdde3379726ea228b38")
	leanderPubKeyHash := Hex2Bytes("b8f3e65b3cabc93fb9459b7e8182fa5ec4e58f04")

	utxoRodrigo := utxos.FindUTXO(rodrigoPubKeyHash)
	assert.Equal(t, []TXOutput{{BlockReward, rodrigoPubKeyHash}}, utxoRodrigo)

	utxoLeander := utxos.FindUTXO(leanderPubKeyHash)
	assert.Equal(t, []TXOutput(nil), utxoLeander)

	// 14vRYoWsjqC61tNmaLPPzjKnxirSxFoehh sent 5 "coins" to 1HrwWkjdwQuhaHSco9H7u7SVsmo4aeDZBX
	// update utxo
	utxos = getTestExpectedUTXOSet("block1")
	utxoRodrigo = utxos.FindUTXO(rodrigoPubKeyHash)
	assert.Equal(t, []TXOutput{{5, rodrigoPubKeyHash}}, utxoRodrigo)

	utxoLeander = utxos.FindUTXO(leanderPubKeyHash)
	assert.Equal(t, []TXOutput{{5, leanderPubKeyHash}}, utxoLeander)

	// 14vRYoWsjqC61tNmaLPPzjKnxirSxFoehh sent 1 "coin" to 1HrwWkjdwQuhaHSco9H7u7SVsmo4aeDZBX and
	// 1HrwWkjdwQuhaHSco9H7u7SVsmo4aeDZBX sent 3 "coins" to 14vRYoWsjqC61tNmaLPPzjKnxirSxFoehh
	// update utxo
	utxos = getTestExpectedUTXOSet("block2")

	utxoRodrigo = utxos.FindUTXO(rodrigoPubKeyHash)
	assert.ElementsMatch(t, []TXOutput{
		{4, rodrigoPubKeyHash},
		{3, rodrigoPubKeyHash},
	}, utxoRodrigo)
	assert.Equal(t, 2, len(utxoRodrigo))

	utxoLeander = utxos.FindUTXO(leanderPubKeyHash)
	assert.ElementsMatch(t, []TXOutput{
		{2, leanderPubKeyHash},
		{1, leanderPubKeyHash},
	}, utxoLeander)
	assert.Equal(t, 2, len(utxoLeander))
}

func TestCountUTXOs(t *testing.T) {
	utxos := getTestExpectedUTXOSet("block0")
	assert.Equal(t, 1, utxos.CountUTXOs())

	utxos = getTestExpectedUTXOSet("block1")
	assert.Equal(t, 3, utxos.CountUTXOs())

	utxos = getTestExpectedUTXOSet("block2")
	assert.Equal(t, 5, utxos.CountUTXOs())
}

func TestUpdate(t *testing.T) {
	for k, m := range testUTXOs {
		t.Run(k, func(t *testing.T) {
			utxos := m.utxos
			failMsg := fmt.Sprintf("UTXO update failed for %s", k)
			switch k {
			case "block0":
				utxos.Update(testBlockchainData["block0"].Transactions)
				diff(t, m.expectedUTXOs, utxos, failMsg)
			case "block1":
				utxos.Update(testBlockchainData["block1"].Transactions)
				diff(t, m.expectedUTXOs, utxos, failMsg)
			case "block2":
				utxos.Update(testBlockchainData["block2"].Transactions)
				diff(t, m.expectedUTXOs, utxos, failMsg)
			case "block3":
				utxos.Update(testBlockchainData["block3"].Transactions)
				diff(t, m.expectedUTXOs, utxos, failMsg)
			case "block4":
				utxos.Update(testBlockchainData["block4"].Transactions)
				diff(t, m.expectedUTXOs, utxos, failMsg)
			}
		})
	}
}
