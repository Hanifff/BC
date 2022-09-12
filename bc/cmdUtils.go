package main

import "crypto/ecdsa"

const QUERY = `1: Create a blockchain
2: Transfer 10 coins from a (miner) to b
3: Mine a block
4: Print chain
5: Print txs and all orher related info for a block
6: Print a spesific txs
7: Transfer coins from c
8: Transfer 5 coins from b to c
9: Get balance\n10: print utxo set` + "\n"

type Balance struct {
	Address string
	Funds   int
}

func getBalance(address string, u UTXOSet) *Balance {
	utxos := u.FindUTXO(GetPubKeyHashFromAddress(address))
	balance := 0
	for _, utxo := range utxos {
		balance += utxo.Value
	}
	return &Balance{Address: address, Funds: balance}
}

type Indetity struct {
	pk      ecdsa.PrivateKey
	pubkey  []byte
	address string
}

func CreateIdentities() *Indetity {
	pk, pubKey := newKeyPair()
	addr := GetAddress(pubKey)
	return &Indetity{pk: pk, pubkey: pubKey, address: string(addr)}
}
