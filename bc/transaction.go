package main

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/gob"
	"errors"
	"fmt"
	"math/big"
	"strings"
)

var (
	ErrNoFunds         = errors.New("not enough funds")
	ErrTxInputNotFound = errors.New("transaction input not found")
)

// Transaction represents a Bitcoin transaction
type Transaction struct {
	ID   []byte
	Vin  []TXInput
	Vout []TXOutput
}

// NewCoinbaseTX creates a new coinbase transaction
func NewCoinbaseTX(to, data string) (*Transaction, error) {
	if data == "" {
		data = fmt.Sprintf("Reward to %s", to)
	}
	txin := TXInput{OutIdx: -1, PubKey: []byte(data)}
	txout := TXOutput{Value: BlockReward, PubKeyHash: GetPubKeyHashFromAddress(to)}
	txn := &Transaction{Vin: []TXInput{txin}, Vout: []TXOutput{txout}}
	txn.ID = txn.Hash()
	return txn, nil
}

// NewUTXOTransaction creates a new UTXO transaction
// NOTE: The returned tx is NOT signed!
func NewUTXOTransaction(pubKey []byte, to string, amount int, utxos UTXOSet) (*Transaction, error) {
	outputs := []TXOutput{}
	hpubkey := HashPubKey(pubKey)
	curBalance, inputs := utxoTxInputs(utxos, pubKey)
	if curBalance >= amount {
		txout := TXOutput{Value: amount, PubKeyHash: GetPubKeyHashFromAddress(to)}
		outputs = append(outputs, txout)
		unspent := curBalance - amount
		if unspent > 0 {
			outMyself := TXOutput{Value: unspent, PubKeyHash: hpubkey}
			outputs = append(outputs, outMyself)
		}
		txn := &Transaction{Vin: inputs, Vout: outputs}
		id := txn.Hash()
		txn.ID = id
		return txn, nil
	}
	return nil, ErrNoFunds
}

// IsCoinbase checks whether the transaction is coinbase
func (tx Transaction) IsCoinbase() bool {
	return tx.Vin[0].OutIdx == -1
}

// Equals checks if the given transaction ID matches the ID of tx
func (tx Transaction) Equals(ID []byte) bool {
	return bytes.Equal(tx.ID, ID)
}

// Serialize returns a serialized Transaction
func (tx Transaction) Serialize() []byte {
	var buff bytes.Buffer
	enc := gob.NewEncoder(&buff)
	err := enc.Encode(tx)
	if err != nil {
		panic("Could not encode the transaction!")
	}
	return buff.Bytes()
}

// Hash returns the hash of the Transaction
func (tx *Transaction) Hash() []byte {
	txCopy := *tx
	txCopy.ID = nil
	serlizedTxn := txCopy.Serialize()
	hbytes := sha256.Sum256(serlizedTxn)
	return hbytes[:]
}

// TrimmedCopy creates a trimmed copy of Transaction to be used in signing
func (tx Transaction) TrimmedCopy() Transaction {
	txinputs := []TXInput{}
	for _, inp := range tx.Vin {
		txin := TXInput{Txid: inp.Txid,
			OutIdx:    inp.OutIdx,
			Signature: nil,
			PubKey:    nil}
		txinputs = append(txinputs, txin)
	}
	tx.Vin = txinputs
	return tx
}

// Sign signs each input of a Transaction
func (tx *Transaction) Sign(privKey ecdsa.PrivateKey, prevTXs map[string]*Transaction) error {
	var signature []byte
	var txinputs []TXInput
	if tx.IsCoinbase() {
		return nil
	}
	if !validInputs(tx.Vin, prevTXs) {
		return ErrTxInputNotFound
	}
	// data to be signed
	trimCopy := tx.TrimmedCopy()
	data := dataToSign(&trimCopy, prevTXs)
	r, s, err := ecdsa.Sign(rand.Reader, &privKey, data.Serialize())
	if err != nil {
		return errors.New("could not sign the transaction input")
	}
	signature = append(signature, r.Bytes()...)
	signature = append(signature, s.Bytes()...)

	for _, inp := range trimCopy.Vin {
		if inp.Signature == nil {
			txin := TXInput{Txid: inp.Txid,
				OutIdx:    inp.OutIdx,
				Signature: signature,
				PubKey:    pubKeyToByte(privKey.PublicKey)}
			txinputs = append(txinputs, txin)
		}
	}
	tx.Vin = txinputs
	return nil
}

// Verify verifies signatures of Transaction inputs
func (tx Transaction) Verify(prevTXs map[string]*Transaction) bool {
	if tx.IsCoinbase() {
		return true
	}
	if !validInputs(tx.Vin, prevTXs) {
		return false
	}
	// reconstruct the signing data
	trimCopy := tx.TrimmedCopy()
	data := dataToSign(&trimCopy, prevTXs)
	elCurve := elliptic.P256()
	for _, inp := range tx.Vin {
		if len(inp.Signature) == 0 {
			return false
		}
		r, s := extractSign(inp.Signature)
		X, Y := extractPubkey(inp.PubKey)
		//X,Y := elliptic.Unmarshal(elCurve, inp.PubKey)
		verfiedPubKey := ecdsa.PublicKey{Curve: elCurve, X: X, Y: Y}
		verifySign := ecdsa.Verify(&verfiedPubKey, data.Serialize(), r, s)
		if !verifySign {
			return false
		}
	}
	return true
}

// String returns a human-readable representation of a transaction
func (tx Transaction) String() string {
	var lines []string

	lines = append(lines, fmt.Sprintf("--- Transaction %x :", tx.ID))

	for i, input := range tx.Vin {
		lines = append(lines, fmt.Sprintf("     Input %d:", i))
		lines = append(lines, fmt.Sprintf("       TXID:      %x", input.Txid))
		lines = append(lines, fmt.Sprintf("       OutIdx:    %d", input.OutIdx))
		lines = append(lines, fmt.Sprintf("       Signature: %x", input.Signature))
		lines = append(lines, fmt.Sprintf("       PubKey: %x", input.PubKey))
	}

	for i, output := range tx.Vout {
		lines = append(lines, fmt.Sprintf("     Output %d:", i))
		lines = append(lines, fmt.Sprintf("       Value:  %d", output.Value))
		lines = append(lines, fmt.Sprintf("       PubKeyHash: %x", output.PubKeyHash))
	}

	return strings.Join(lines, "\n")
}

func utxoTxInputs(utxos UTXOSet, pubKey []byte) (int, []TXInput) {
	curBalance := 0
	inputs := []TXInput{}
	hpubkey := HashPubKey(pubKey)
	for id, utxoOut := range utxos {
		tx := Hex2Bytes(id)
		for idx, utxo := range utxoOut {
			if utxo.IsLockedWithKey(hpubkey) {
				curBalance += utxo.Value
				txin := TXInput{Txid: tx, OutIdx: idx, PubKey: pubKey}
				inputs = append(inputs, txin)
			}
		}
	}
	return curBalance, inputs
}

func validInputs(inputs []TXInput, prevTXs map[string]*Transaction) bool {
	for _, inp := range inputs {
		if _, ok := prevTXs[fmt.Sprintf("%x", inp.Txid)]; !ok {
			return false
		}
	}
	return true
}

func dataToSign(trimCopy *Transaction, prevTXs map[string]*Transaction) *Transaction {
	var txinputsKeys []TXInput
	for _, inp := range trimCopy.Vin {
		if inp.Signature == nil {
			getPrevOutput := prevTXs[fmt.Sprintf("%x", inp.Txid)].Vout[inp.OutIdx]
			txin := inp
			txin.PubKey = getPrevOutput.PubKeyHash
			txinputsKeys = append(txinputsKeys, txin)
		}
	}
	trimCopy.Vin = txinputsKeys // update the trimcopy with new input.publickeys
	return trimCopy
}

func extractSign(sign []byte) (*big.Int, *big.Int) {
	rs := sign // reconstruct the signature
	mid := len(rs) / 2
	r := toBigInt(rs[:mid])
	s := toBigInt(rs[mid:])
	return r, s
}

func extractPubkey(XY []byte) (*big.Int, *big.Int) {
	mid := len(XY) / 2
	X := toBigInt(XY[:mid])
	Y := toBigInt(XY[mid:])
	if X == nil {
		panic("could not revert the pubkey")
	}
	return X, Y
}
