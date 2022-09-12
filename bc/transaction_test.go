package main

import (
	"bytes"
	"encoding/gob"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEquals(t *testing.T) {
	tx0ID := testTransactions["tx0"].ID
	for _, test := range []struct {
		name   string
		tx     *Transaction
		result bool
	}{
		{
			name:   "equal",
			tx:     testTransactions["tx0"],
			result: true,
		},
		{
			name:   "not equal",
			tx:     testTransactions["tx1"],
			result: false,
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.result, test.tx.Equals(tx0ID))
		})
	}
}

func TestSerialize(t *testing.T) {
	for name, tx := range testTransactions {
		t.Run(name, func(t *testing.T) {
			serialized := tx.Serialize()

			dec := gob.NewDecoder(bytes.NewReader(serialized))
			decoded := &Transaction{}
			err := dec.Decode(decoded)
			if err != nil {
				t.Fatalf("error decoding tx: %v", err)
			}
			diff(t, tx, decoded, "wrong serialization")
		})
	}
}

func TestHash(t *testing.T) {
	for name, tx := range testTransactions {
		t.Run(name, func(t *testing.T) {
			txhash := tx.Hash()
			if !bytes.Equal(tx.ID, txhash) {
				t.Errorf("wrong tx hash:\n-want: %x\n+got: %x\n", tx.ID, txhash)
			}
		})
	}
}

func TestIsCoinbase(t *testing.T) {
	tx0 := testTransactions["tx0"]
	assert.True(t, tx0.IsCoinbase())

	tx1 := testTransactions["tx1"]
	assert.False(t, tx1.IsCoinbase())
}

func TestNewCoinbaseTXWithData(t *testing.T) {
	// Passing data to the coinbase transaction
	tx, err := NewCoinbaseTX("14vRYoWsjqC61tNmaLPPzjKnxirSxFoehh", GenesisCoinbaseData)
	if tx == nil {
		t.Fatal("NewCoinbaseTX returned nil")
	}
	assert.Nil(t, err)
	assert.Equal(t, Hex2Bytes("5468652054696d65732030332f4a616e2f32303039204368616e63656c6c6f72206f6e206272696e6b206f66207365636f6e64206261696c6f757420666f722062616e6b73"), tx.Vin[0].PubKey)
	assert.Equal(t, -1, tx.Vin[0].OutIdx)
	assert.Nil(t, tx.Vin[0].Txid)
	assert.Nil(t, tx.Vin[0].Signature)
	assert.Equal(t, BlockReward, tx.Vout[0].Value)
	assert.Equal(t, Hex2Bytes("2b02ea4c157844ec0b034fdde3379726ea228b38"), tx.Vout[0].PubKeyHash)
}

func TestNewCoinbaseTXWithDefaultData(t *testing.T) {
	// Using default data
	tx, err := NewCoinbaseTX("14vRYoWsjqC61tNmaLPPzjKnxirSxFoehh", "")
	if tx == nil {
		t.Fatal("NewCoinbaseTX returned nil")
	}
	assert.Nil(t, err)
	assert.Nil(t, tx.Vin[0].Txid)
	assert.Equal(t, -1, tx.Vin[0].OutIdx)
	assert.Equal(t, BlockReward, tx.Vout[0].Value)
	assert.Equal(t, Hex2Bytes("2b02ea4c157844ec0b034fdde3379726ea228b38"), tx.Vout[0].PubKeyHash)
}

func TestNewUTXOTransaction(t *testing.T) {
	pubKey1Bytes := Hex2Bytes("f86aa0caf08359ee4227d2901ab490172c69a801910f4140cdde2f5dc8f8bb3dc19da2c9fb0ed041db106a8fea0382de25edbc83df6893574e40fc2e1e493748")
	fromAddress := "14vRYoWsjqC61tNmaLPPzjKnxirSxFoehh"

	pubKey2Bytes := Hex2Bytes("c36d68bc641029e53a38252b436c596ef3d03a4a754743da50fb9a321020e882dd401732381783c7444112abc729b3bee04643015d80fe67e0c28a5b28a20910")
	toAddress := "1HrwWkjdwQuhaHSco9H7u7SVsmo4aeDZBX"

	// "from" address have 10 (i.e., genesis coinbase) and "to" address have 0
	bc := newMockBlockchain()
	utxos := UTXOSet{
		"9402c56f49de02d2b9c4633837d82e3881227a3ea90c4073c02815fdcf5afaa2": {0: testTransactions["tx0"].Vout[0]},
	}

	// Reject if there is not sufficient funds
	tx1, err := NewUTXOTransaction(pubKey2Bytes, fromAddress, 5, utxos)
	assert.ErrorIs(t, err, ErrNoFunds)
	assert.Nil(t, tx1)

	// Accept otherwise
	tx1, err = NewUTXOTransaction(pubKey1Bytes, toAddress, 5, utxos)
	assert.Nil(t, err)
	if tx1 == nil {
		t.Fatal("NewUTXOTransaction returned nil")
	}
	removeTXInputSignature(tx1)
	diff(t, testTransactions["tx1"], tx1, "incorrect transaction")

	// update utxo and blockchain with tx1
	addMockBlock(bc, testBlockchainData["block1"])
	utxos = UTXOSet{
		"397b990007845099b4fe50ba23490f277b3bf6f5316b4082c343b14c5504ab13": {
			0: testTransactions["tx1"].Vout[0],
			1: testTransactions["tx1"].Vout[1],
		},
	}

	tx2, err := NewUTXOTransaction(pubKey2Bytes, fromAddress, 3, utxos)
	assert.Nil(t, err)
	removeTXInputSignature(tx2)
	diff(t, testTransactions["tx2"], tx2, "incorrect transaction")

	tx3, err := NewUTXOTransaction(pubKey1Bytes, toAddress, 1, utxos)
	assert.Nil(t, err)
	removeTXInputSignature(tx3)
	diff(t, testTransactions["tx3"], tx3, "incorrect transaction")
}

func TestSign(t *testing.T) {
	privKey, _ := decodeKeyPair(testEncPrivKeyUser1, testEncPubKeyUser1)

	tx := &Transaction{
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
	}

	prevTXs := make(map[string]*Transaction)
	prevTXs["9402c56f49de02d2b9c4633837d82e3881227a3ea90c4073c02815fdcf5afaa2"] = testTransactions["tx0"]

	err := tx.Sign(*privKey, prevTXs)
	assert.Nil(t, err)
	assert.NotNil(t, tx.Vin[0].Signature)
}

func TestSignIgnoreCoinbaseTX(t *testing.T) {
	privKey, _ := decodeKeyPair(testEncPrivKeyUser1, testEncPubKeyUser1)

	tx := &Transaction{
		ID: Hex2Bytes("9402c56f49de02d2b9c4633837d82e3881227a3ea90c4073c02815fdcf5afaa2"),
		Vin: []TXInput{
			{Txid: nil, OutIdx: -1, Signature: nil, PubKey: []byte(GenesisCoinbaseData)},
		},
		Vout: []TXOutput{
			{
				Value:      BlockReward,
				PubKeyHash: Hex2Bytes("2b02ea4c157844ec0b034fdde3379726ea228b38"),
			},
		},
	}
	prevTXs := make(map[string]*Transaction)

	err := tx.Sign(*privKey, prevTXs)
	assert.Nil(t, err)
	assert.Nil(t, tx.Vin[0].Signature)
}

func TestSignInvalidInputTX(t *testing.T) {
	privKey, _ := decodeKeyPair(testEncPrivKeyUser1, testEncPubKeyUser1)

	tx := &Transaction{
		ID: Hex2Bytes("397b990007845099b4fe50ba23490f277b3bf6f5316b4082c343b14c5504ab13"),
		Vin: []TXInput{
			{
				Txid:      Hex2Bytes("non-existentID"),
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
	}

	prevTXs := make(map[string]*Transaction)
	prevTXs["9402c56f49de02d2b9c4633837d82e3881227a3ea90c4073c02815fdcf5afaa2"] = testTransactions["tx0"]

	err := tx.Sign(*privKey, prevTXs)
	assert.ErrorIs(t, err, ErrTxInputNotFound)
	assert.Nil(t, tx.Vin[0].Signature)
}

func TestVerifyIgnoreCoinbaseTX(t *testing.T) {
	tx := testTransactions["tx0"]
	prevTXs := make(map[string]*Transaction)
	assert.True(t, tx.Verify(prevTXs))
}

func TestVerify(t *testing.T) {
	tx := &Transaction{
		ID: Hex2Bytes("397b990007845099b4fe50ba23490f277b3bf6f5316b4082c343b14c5504ab13"),
		Vin: []TXInput{
			{
				Txid:      Hex2Bytes("9402c56f49de02d2b9c4633837d82e3881227a3ea90c4073c02815fdcf5afaa2"),
				OutIdx:    0,
				Signature: Hex2Bytes("17b6db89942bb02b485332c9a3b37638e02a3dfafdf4c3a4fad7fc4c7b062cc8156b75957050e049cd307853522f5ef49339b1b1230359f59571af12c612bde2"),
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
	}

	prevTXs := make(map[string]*Transaction)
	prevTXs["9402c56f49de02d2b9c4633837d82e3881227a3ea90c4073c02815fdcf5afaa2"] = testTransactions["tx0"]
	
	assert.True(t, tx.Verify(prevTXs))
}

func TestVerifyInvalidInputTX(t *testing.T) {
	tx := &Transaction{
		ID: Hex2Bytes("397b990007845099b4fe50ba23490f277b3bf6f5316b4082c343b14c5504ab13"),
		Vin: []TXInput{
			{
				Txid:      Hex2Bytes("non-existentID"),
				OutIdx:    0,
				Signature: Hex2Bytes("17b6db89942bb02b485332c9a3b37638e02a3dfafdf4c3a4fad7fc4c7b062cc8156b75957050e049cd307853522f5ef49339b1b1230359f59571af12c612bde2"),
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
	}

	prevTXs := make(map[string]*Transaction)
	prevTXs["9402c56f49de02d2b9c4633837d82e3881227a3ea90c4073c02815fdcf5afaa2"] = testTransactions["tx0"]

	assert.False(t, tx.Verify(prevTXs))
}

func TestVerifyInvalidSignature(t *testing.T) {
	tx := &Transaction{
		ID: Hex2Bytes("397b990007845099b4fe50ba23490f277b3bf6f5316b4082c343b14c5504ab13"),
		Vin: []TXInput{
			{
				Txid:      Hex2Bytes("9402c56f49de02d2b9c4633837d82e3881227a3ea90c4073c02815fdcf5afaa2"),
				OutIdx:    0,
				Signature: Hex2Bytes("invalid"),
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
	}

	prevTXs := make(map[string]*Transaction)
	prevTXs["9402c56f49de02d2b9c4633837d82e3881227a3ea90c4073c02815fdcf5afaa2"] = testTransactions["tx0"]

	assert.False(t, tx.Verify(prevTXs))
}

func TestTrimmedCopy(t *testing.T) {
	tx := &Transaction{
		ID: Hex2Bytes("397b990007845099b4fe50ba23490f277b3bf6f5316b4082c343b14c5504ab13"),
		Vin: []TXInput{
			{
				Txid:      Hex2Bytes("9402c56f49de02d2b9c4633837d82e3881227a3ea90c4073c02815fdcf5afaa2"),
				OutIdx:    0,
				Signature: Hex2Bytes("17b6db89942bb02b485332c9a3b37638e02a3dfafdf4c3a4fad7fc4c7b062cc8156b75957050e049cd307853522f5ef49339b1b1230359f59571af12c612bde2"),
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
	}

	txCopy := tx.TrimmedCopy()
	if len(txCopy.Vin) == 0 {
		t.Fatal("Transaction copy does not contain inputs")
	}
	assert.Nil(t, txCopy.Vin[0].Signature)
	assert.Nil(t, txCopy.Vin[0].PubKey)
	assert.Equal(t, tx.Vin[0].Txid, txCopy.Vin[0].Txid)
	assert.Equal(t, tx.Vin[0].OutIdx, txCopy.Vin[0].OutIdx)
	assert.Equal(t, tx.Vout, txCopy.Vout)
	assert.Equal(t, tx.ID, txCopy.ID)
}
