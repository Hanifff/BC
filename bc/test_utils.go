package main

import (
	"bytes"
	"fmt"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

// Fixed block timestamp
const TestBlockTime int64 = 1563897484

// Format error message. Based on:
// https://cs.opensource.google/go/go/+/refs/tags/go1.17:src/testing/testing.go;l=537
func errorf(file string, line int, s string) string {
	if line == 0 {
		line = 1
	}
	buf := new(strings.Builder)
	// Every line is indented at least 4 spaces.
	buf.WriteString("    ")
	fmt.Fprintf(buf, "%s:%d: ", filepath.Base(file), line)
	lines := strings.Split(s, "\n")
	if l := len(lines); l > 1 && lines[l-1] == "" {
		lines = lines[:l-1]
	}
	for i, line := range lines {
		if i > 0 {
			// Second and subsequent lines are indented an additional 4 spaces.
			buf.WriteString("\n        ")
		}
		buf.WriteString(line)
	}
	buf.WriteByte('\n')
	return buf.String()
}

func diff(t *testing.T, want interface{}, got interface{}, message string) {
	if diff := cmp.Diff(want, got); diff != "" {
		_, file, line, _ := runtime.Caller(1)
		fmt.Print(errorf(file, line, fmt.Sprintf("%s: (-want +got)\n%s", message, diff)))
		t.Fail()
	}
}

func removeTXInputSignature(tx *Transaction) {
	var inputs []TXInput

	for _, vin := range tx.Vin {
		inputs = append(inputs, TXInput{vin.Txid, vin.OutIdx, nil, vin.PubKey})
	}
	*tx = Transaction{tx.ID, inputs, tx.Vout}
}

// Transactions example flow:
// tx0: genesis coinbase tx - Rodrigo received 10 coins
// tx1: Rodrigo sent 5 coins to Leander and get 5 coins as remainder
// tx2: Using tx1 output, Leander sent 3 of his coins to Rodrigo
// and get 2 coins as remainder
// tx3: Using tx1 output, Rodrigo sent 1 coin to Leander and
// get 4 coins as remainder
// tx4: Using tx2 output, Rodrigo sent 2 coins to Leander and
// get 1 coin as remainder
// tx5: Using tx3 and tx4 outputs, Leander sent 3 "coins" to Rodrigo
//
// NOTE: The mocked txs below ignores the tx signature!
var testTransactions = map[string]*Transaction{
	"tx0": {
		ID: Hex2Bytes("9402c56f49de02d2b9c4633837d82e3881227a3ea90c4073c02815fdcf5afaa2"),
		Vin: []TXInput{
			{
				Txid:      nil,
				OutIdx:    -1,
				Signature: nil,
				PubKey:    []byte(GenesisCoinbaseData),
			},
		},
		Vout: []TXOutput{
			{
				Value:      BlockReward,
				PubKeyHash: Hex2Bytes("2b02ea4c157844ec0b034fdde3379726ea228b38"),
			},
		},
	},
	"tx1": {
		ID: Hex2Bytes("397b990007845099b4fe50ba23490f277b3bf6f5316b4082c343b14c5504ab13"),
		Vin: []TXInput{
			{
				Txid:      Hex2Bytes("9402c56f49de02d2b9c4633837d82e3881227a3ea90c4073c02815fdcf5afaa2"),
				OutIdx:    0,
				Signature: nil,
				PubKey:    Hex2Bytes("f86aa0caf08359ee4227d2901ab490172c69a801910f4140cdde2f5dc8f8bb3dc19da2c9fb0ed041db106a8fea0382de25edbc83df6893574e40fc2e1e493748"),
			},
		},
		Vout: []TXOutput{
			{
				Value:      5,
				PubKeyHash: Hex2Bytes("b8f3e65b3cabc93fb9459b7e8182fa5ec4e58f04"),
			},
			{
				Value:      5,
				PubKeyHash: Hex2Bytes("2b02ea4c157844ec0b034fdde3379726ea228b38"),
			},
		},
	},
	"tx2": {
		ID: Hex2Bytes("dcd76d254f7a41888e6bda9958c4ceadf510e1bd5fd251f617c91b704fbf9492"),
		Vin: []TXInput{
			{
				Txid:      Hex2Bytes("397b990007845099b4fe50ba23490f277b3bf6f5316b4082c343b14c5504ab13"),
				OutIdx:    0,
				Signature: nil,
				PubKey:    Hex2Bytes("c36d68bc641029e53a38252b436c596ef3d03a4a754743da50fb9a321020e882dd401732381783c7444112abc729b3bee04643015d80fe67e0c28a5b28a20910"),
			},
		},
		Vout: []TXOutput{
			{
				Value:      3,
				PubKeyHash: Hex2Bytes("2b02ea4c157844ec0b034fdde3379726ea228b38"),
			},
			{
				Value:      2,
				PubKeyHash: Hex2Bytes("b8f3e65b3cabc93fb9459b7e8182fa5ec4e58f04"),
			},
		},
	},
	"tx3": {
		ID: Hex2Bytes("e9e5fc159f24b2b33310f77aef4e425e77ed71be87dbf9a0c7764b5417bd3e4b"),
		Vin: []TXInput{
			{
				Txid:      Hex2Bytes("397b990007845099b4fe50ba23490f277b3bf6f5316b4082c343b14c5504ab13"),
				OutIdx:    1,
				Signature: nil,
				PubKey:    Hex2Bytes("f86aa0caf08359ee4227d2901ab490172c69a801910f4140cdde2f5dc8f8bb3dc19da2c9fb0ed041db106a8fea0382de25edbc83df6893574e40fc2e1e493748"),
			},
		},
		Vout: []TXOutput{
			{
				Value:      1,
				PubKeyHash: Hex2Bytes("b8f3e65b3cabc93fb9459b7e8182fa5ec4e58f04"),
			},
			{
				Value:      4,
				PubKeyHash: Hex2Bytes("2b02ea4c157844ec0b034fdde3379726ea228b38"),
			},
		},
	},
	"tx4": {
		ID: Hex2Bytes("91d6fe8fe351e50fa6e16bb391ff74f5dc650646ce6ad02442e647742566b31b"),
		Vin: []TXInput{
			{
				Txid:      Hex2Bytes("dcd76d254f7a41888e6bda9958c4ceadf510e1bd5fd251f617c91b704fbf9492"),
				OutIdx:    0,
				Signature: nil,
				PubKey:    Hex2Bytes("f86aa0caf08359ee4227d2901ab490172c69a801910f4140cdde2f5dc8f8bb3dc19da2c9fb0ed041db106a8fea0382de25edbc83df6893574e40fc2e1e493748"),
			},
		},
		Vout: []TXOutput{
			{
				Value:      2,
				PubKeyHash: Hex2Bytes("b8f3e65b3cabc93fb9459b7e8182fa5ec4e58f04"),
			},
			{
				Value:      1,
				PubKeyHash: Hex2Bytes("2b02ea4c157844ec0b034fdde3379726ea228b38"),
			},
		},
	},
	"tx5": {
		ID: Hex2Bytes("b63d956b234d27c3494d9935ac9764634db0232f32ef7f576979d8ba5ec93fbc"),
		Vin: []TXInput{
			{
				Txid:      Hex2Bytes("e9e5fc159f24b2b33310f77aef4e425e77ed71be87dbf9a0c7764b5417bd3e4b"),
				OutIdx:    0,
				Signature: nil,
				PubKey:    Hex2Bytes("c36d68bc641029e53a38252b436c596ef3d03a4a754743da50fb9a321020e882dd401732381783c7444112abc729b3bee04643015d80fe67e0c28a5b28a20910"),
			},
			{
				Txid:      Hex2Bytes("91d6fe8fe351e50fa6e16bb391ff74f5dc650646ce6ad02442e647742566b31b"),
				OutIdx:    0,
				Signature: nil,
				PubKey:    Hex2Bytes("c36d68bc641029e53a38252b436c596ef3d03a4a754743da50fb9a321020e882dd401732381783c7444112abc729b3bee04643015d80fe67e0c28a5b28a20910"),
			},
		},
		Vout: []TXOutput{
			{
				Value:      3,
				PubKeyHash: Hex2Bytes("2b02ea4c157844ec0b034fdde3379726ea228b38"),
			},
		},
	},
}

func getTestInputsTX(tx string) []TXInput {
	return testTransactions[tx].Vin
}

func newMockCoinbaseTX(to, data, txID string) *Transaction {
	tx := &Transaction{
		ID: Hex2Bytes(txID),
		Vin: []TXInput{
			{
				Txid:      nil,
				OutIdx:    -1,
				Signature: nil,
				PubKey:    []byte(data),
			},
		},
		Vout: []TXOutput{
			{
				Value:      BlockReward,
				PubKeyHash: Hex2Bytes(to),
			},
		},
	}
	return tx
}

// Miner address: 12znKfjybYauJASaggYEKCWyN9MLKYfA5i
var minerCoinbaseTx = map[string]*Transaction{
	"tx1": newMockCoinbaseTX("15e5ab1b9f1e79b58c95a1a0b3caa63c61617971", "1", "0ca136effc2424a42d2bcf6b498e7c0c226ada6eff5499a7fa600c0ae6bad9c0"),
	"tx2": newMockCoinbaseTX("15e5ab1b9f1e79b58c95a1a0b3caa63c61617971", "2", "64e97834110d5525f68fbf719743cd22feffb4e91ffb50639f5e232228e3f1e5"),
	"tx3": newMockCoinbaseTX("15e5ab1b9f1e79b58c95a1a0b3caa63c61617971", "3", "68f0b05abdfa09bbfb732e37248ccb2a737db189d03d22224c3aa13afe593994"),
	"tx4": newMockCoinbaseTX("15e5ab1b9f1e79b58c95a1a0b3caa63c61617971", "4", "c8b152a0040e1f98b261b41444d6eeca09c3bfcc7d1ec69a792748f70efa1efb"),
}

var testBlockchainData = map[string]*Block{
	"block0": { // genesis block
		Timestamp: TestBlockTime,
		Transactions: []*Transaction{
			testTransactions["tx0"],
		},
		PrevBlockHash: nil,
		Hash:          Hex2Bytes("00d4eeaee903dce5468d4c6975376dfbc4c45ea1bc6c5bbbfd8e13b26aaf6e3b"),
		Nonce:         59,
	},
	"block1": {
		Timestamp: TestBlockTime,
		Transactions: []*Transaction{
			minerCoinbaseTx["tx1"],
			testTransactions["tx1"],
		},
		PrevBlockHash: Hex2Bytes("00d4eeaee903dce5468d4c6975376dfbc4c45ea1bc6c5bbbfd8e13b26aaf6e3b"),
		Hash:          Hex2Bytes("0017379eb3ca189e5deaecc58523533883452ffe7dbba43cd71c3319d392a931"),
		Nonce:         35,
	},
	"block2": {
		Timestamp: TestBlockTime,
		Transactions: []*Transaction{
			minerCoinbaseTx["tx2"],
			testTransactions["tx3"],
			testTransactions["tx2"],
		},
		PrevBlockHash: Hex2Bytes("0017379eb3ca189e5deaecc58523533883452ffe7dbba43cd71c3319d392a931"),
		Hash:          Hex2Bytes("00f53a20971ded89b7e8192fbac5255211bc538e1ef705255dff352a78855b46"),
		Nonce:         626,
	},
	"block3": {
		Timestamp: TestBlockTime,
		Transactions: []*Transaction{
			minerCoinbaseTx["tx3"],
			testTransactions["tx4"],
		},
		PrevBlockHash: Hex2Bytes("00f53a20971ded89b7e8192fbac5255211bc538e1ef705255dff352a78855b46"),
		Hash:          Hex2Bytes("00efff94a452db304494aecb66dde7dad4d4ddea986d6eb7a7c71b70d6512a7d"),
		Nonce:         279,
	},
	"block4": {
		Timestamp: TestBlockTime,
		Transactions: []*Transaction{
			minerCoinbaseTx["tx4"],
			testTransactions["tx5"],
		},
		PrevBlockHash: Hex2Bytes("00efff94a452db304494aecb66dde7dad4d4ddea986d6eb7a7c71b70d6512a7d"),
		Hash:          Hex2Bytes("00cb768ee28075b65bec7bb949571a1a940863389bfbbbcce9a3e82b78a94feb"),
		Nonce:         154,
	},
}

var testUTXOs = map[string]struct {
	utxos         UTXOSet
	expectedUTXOs UTXOSet
}{
	"block0": { // (0 input -> 1 output, generating "coins")
		utxos: UTXOSet{},
		expectedUTXOs: UTXOSet{
			"9402c56f49de02d2b9c4633837d82e3881227a3ea90c4073c02815fdcf5afaa2": {0: testTransactions["tx0"].Vout[0]},
			// tx0: Address 14vRYoWsjqC61tNmaLPPzjKnxirSxFoehh create coinbase transaction and received 10 "coins"
		},
	},
	"block1": { // (1 input -> 2 outputs, splitting one input)
		utxos: UTXOSet{
			"9402c56f49de02d2b9c4633837d82e3881227a3ea90c4073c02815fdcf5afaa2": {0: testTransactions["tx0"].Vout[0]},
		},
		expectedUTXOs: UTXOSet{
			"397b990007845099b4fe50ba23490f277b3bf6f5316b4082c343b14c5504ab13": {
				0: testTransactions["tx1"].Vout[0],
				1: testTransactions["tx1"].Vout[1],
			},
			// tx1: 14vRYoWsjqC61tNmaLPPzjKnxirSxFoehh sent 5 "coins" to 1HrwWkjdwQuhaHSco9H7u7SVsmo4aeDZBX and get 5 as remainder
			"0ca136effc2424a42d2bcf6b498e7c0c226ada6eff5499a7fa600c0ae6bad9c0": {
				0: {
					Value:      BlockReward,
					PubKeyHash: Hex2Bytes("15e5ab1b9f1e79b58c95a1a0b3caa63c61617971"),
				},
			},
		},
	},
	"block2": { // (1 input -> 2 output, with multiple txs)
		utxos: UTXOSet{
			"397b990007845099b4fe50ba23490f277b3bf6f5316b4082c343b14c5504ab13": {
				0: testTransactions["tx1"].Vout[0],
				1: testTransactions["tx1"].Vout[1],
			},
		},
		expectedUTXOs: UTXOSet{
			"dcd76d254f7a41888e6bda9958c4ceadf510e1bd5fd251f617c91b704fbf9492": {
				0: testTransactions["tx2"].Vout[0],
				1: testTransactions["tx2"].Vout[1],
			},
			// tx2: 14vRYoWsjqC61tNmaLPPzjKnxirSxFoehh sent 1 "coin" to 1HrwWkjdwQuhaHSco9H7u7SVsmo4aeDZBX and get 4 as remainder
			"e9e5fc159f24b2b33310f77aef4e425e77ed71be87dbf9a0c7764b5417bd3e4b": {
				0: testTransactions["tx3"].Vout[0],
				1: testTransactions["tx3"].Vout[1],
			},
			// tx3: 1HrwWkjdwQuhaHSco9H7u7SVsmo4aeDZBX sent 3 "coins" to 14vRYoWsjqC61tNmaLPPzjKnxirSxFoehh and get 2 as remainder
			"64e97834110d5525f68fbf719743cd22feffb4e91ffb50639f5e232228e3f1e5": {
				0: {
					Value:      BlockReward,
					PubKeyHash: Hex2Bytes("15e5ab1b9f1e79b58c95a1a0b3caa63c61617971"),
				},
			},
		},
	},
	// NOTE: Some outputs were ignored from the initial utxos to reduce combination of possible input sources
	"block3": { // (1 input -> 2 outputs)
		utxos: UTXOSet{
			// tx3 was intentionally ignored
			"dcd76d254f7a41888e6bda9958c4ceadf510e1bd5fd251f617c91b704fbf9492": {
				0: testTransactions["tx2"].Vout[0],
				1: testTransactions["tx2"].Vout[1],
			},
		},
		expectedUTXOs: UTXOSet{
			"dcd76d254f7a41888e6bda9958c4ceadf510e1bd5fd251f617c91b704fbf9492": {1: testTransactions["tx2"].Vout[1]},
			"91d6fe8fe351e50fa6e16bb391ff74f5dc650646ce6ad02442e647742566b31b": {
				0: testTransactions["tx4"].Vout[0],
				1: testTransactions["tx4"].Vout[1],
			},
			// tx4: 14vRYoWsjqC61tNmaLPPzjKnxirSxFoehh sent 2 "coins" to 1HrwWkjdwQuhaHSco9H7u7SVsmo4aeDZBX and get 1 as remainder
			"68f0b05abdfa09bbfb732e37248ccb2a737db189d03d22224c3aa13afe593994": {
				0: {
					Value:      BlockReward,
					PubKeyHash: Hex2Bytes("15e5ab1b9f1e79b58c95a1a0b3caa63c61617971"),
				},
			},
		},
	},
	"block4": { // (2 inputs -> 1 output)
		utxos: UTXOSet{
			"e9e5fc159f24b2b33310f77aef4e425e77ed71be87dbf9a0c7764b5417bd3e4b": {
				0: testTransactions["tx3"].Vout[0],
				1: testTransactions["tx3"].Vout[1],
			},
			"91d6fe8fe351e50fa6e16bb391ff74f5dc650646ce6ad02442e647742566b31b": {
				0: testTransactions["tx4"].Vout[0],
				1: testTransactions["tx4"].Vout[1],
			},
		},
		expectedUTXOs: UTXOSet{
			"e9e5fc159f24b2b33310f77aef4e425e77ed71be87dbf9a0c7764b5417bd3e4b": {1: testTransactions["tx3"].Vout[1]},
			"91d6fe8fe351e50fa6e16bb391ff74f5dc650646ce6ad02442e647742566b31b": {1: testTransactions["tx4"].Vout[1]},
			"b63d956b234d27c3494d9935ac9764634db0232f32ef7f576979d8ba5ec93fbc": {0: testTransactions["tx5"].Vout[0]},
			// tx5: 1HrwWkjdwQuhaHSco9H7u7SVsmo4aeDZBX sent 3 "coins" to 14vRYoWsjqC61tNmaLPPzjKnxirSxFoehh
			"c8b152a0040e1f98b261b41444d6eeca09c3bfcc7d1ec69a792748f70efa1efb": {
				0: {
					Value:      BlockReward,
					PubKeyHash: Hex2Bytes("15e5ab1b9f1e79b58c95a1a0b3caa63c61617971"),
				},
			},
		},
	},
}

func getTestExpectedUTXOSet(block string) UTXOSet {
	return testUTXOs[block].expectedUTXOs
}

func getTestSpendableOutputs(utxos UTXOSet, pubKeyHash []byte) map[string][]int {
	unspentOutputs := make(map[string][]int)

	for txID, outputs := range utxos {
		for outIdx, out := range outputs {
			if bytes.Equal(out.PubKeyHash, pubKeyHash) {
				unspentOutputs[txID] = append(unspentOutputs[txID], outIdx)
			}
		}
	}
	return unspentOutputs
}
