package main

import (
	"fmt"
	"strings"
)

// UTXOSet represents a set of UTXO as an in-memory cache
// The key of the most external map is the transaction ID
// (encoded as string) that contains these outputs
// {map of transaction ID -> {map of TXOutput Index -> TXOutput}}
type UTXOSet map[string]map[int]TXOutput

// FindSpendableOutputs finds and returns unspent outputs in the UTXO Set
// to reference in inputs
func (u UTXOSet) FindSpendableOutputs(pubKeyHash []byte, amount int) (int, map[string][]int) {
	spendable := make(map[string][]int)
	funds := 0
	for id, utxo := range u {
		for idx, out := range utxo {
			if out.IsLockedWithKey(pubKeyHash) {
				spendable[id] = append(spendable[id], idx)
				funds += out.Value
			}
		}
	}
	if funds >= amount {
		return funds, spendable
	}
	return 0, make(map[string][]int)
}

// FindUTXO finds all UTXO in the UTXO Set for a given unlockingData key (e.g., address)
// This function ignores the index of each output and returns
// a list of all outputs in the UTXO Set that can be unlocked by the user
func (u UTXOSet) FindUTXO(pubKeyHash []byte) []TXOutput {
	var UTXO []TXOutput
	for _, utxo := range u {
		for _, out := range utxo {
			if out.IsLockedWithKey(pubKeyHash) {
				UTXO = append(UTXO, out)
			}
		}
	}
	return UTXO
}

// CountUTXOs returns the number of transactions outputs in the UTXO set
func (u UTXOSet) CountUTXOs() int {
	nrOfTxo := 0
	for _, outputs := range u {
		nrOfTxo += len(outputs)
	}
	return nrOfTxo
}

// Update updates the UTXO Set with the new set of transactions
func (u UTXOSet) Update(transactions []*Transaction) {
	checkDS := make(map[*TXInput]int)
	for _, t := range transactions {
		txnOuts := make(map[int]TXOutput)
		for _, inputs := range t.Vin {
			if _, exist := checkDS[&inputs]; !exist { // check for double speding
				utxo := u[fmt.Sprintf("%x", inputs.Txid)]
				delete(utxo, inputs.OutIdx)
				if len(utxo) == 0 {
					delete(u, fmt.Sprintf("%x", inputs.Txid))
				}
			}
		}
		for idx, out := range t.Vout {
			txnOuts[idx] = out
			u[fmt.Sprintf("%x", t.ID)] = txnOuts
		}
	}
}

func (u UTXOSet) String() string {
	var lines []string

	lines = append(lines, fmt.Sprintf("--- UTXO SET:"))
	for txid, outputs := range u {
		lines = append(lines, fmt.Sprintf("     TxID: %s", txid))
		for i, out := range outputs {
			lines = append(lines, fmt.Sprintf("           Output %d: %v", i, out))
		}
	}

	return strings.Join(lines, "\n")
}
