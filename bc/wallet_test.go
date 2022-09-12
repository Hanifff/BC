package main

import (
	"bytes"
	_ "embed"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Keys for address: 14vRYoWsjqC61tNmaLPPzjKnxirSxFoehh
//go:embed keys/testEncPrivKeyUser1.key
var testEncPrivKeyUser1 string

//go:embed keys/testEncPubKeyUser1.pub
var testEncPubKeyUser1 string

// Keys for address: 1HrwWkjdwQuhaHSco9H7u7SVsmo4aeDZBX
//go:embed keys/testEncPrivKeyUser2.key
var testEncPrivKeyUser2 string

//go:embed keys/testEncPubKeyUser2.pub
var testEncPubKeyUser2 string

var addressTable = []struct {
	pubkey           []byte
	pubKeyHash       []byte
	version          byte
	versionedPayload []byte
	checksum         []byte
	encodedAddress   []byte
	address          string
}{
	{
		pubkey:           Hex2Bytes("4c4b60e3ac2ebd25ca272bc404333a3a71eb605e8ae890287ab1f978169567566ed57778cc3c6fb99e9338d257fdb73b66c74e1689799aa82939b3486dc5ecae"),
		pubKeyHash:       Hex2Bytes("50c6020fbbe3b589489a425b4a3a685d0d92ee84"),
		version:          byte(0x00),
		versionedPayload: Hex2Bytes("0050c6020fbbe3b589489a425b4a3a685d0d92ee84"),
		checksum:         Hex2Bytes("fbfd119a"),
		encodedAddress:   Hex2Bytes("31384e36474272547961545462755663534a6961624a64596864514d3839794a7177"),
		address:          "18N6GBrTyaTTbuVcSJiabJdYhdQM89yJqw",
	},
	{
		pubkey:           Hex2Bytes("452e8c6d393f1e0017f51caabc3aa136ac9e330921346daed61340f1ce5e2f1af1f5ae75cc6748a6cb28c1063259797b37f9bf54e6d87370a23ba06ffba214c6"),
		pubKeyHash:       Hex2Bytes("1f675dd842f2f876e0662e24c4c74262d673fab6"),
		version:          byte(0x00),
		versionedPayload: Hex2Bytes("001f675dd842f2f876e0662e24c4c74262d673fab6"),
		checksum:         Hex2Bytes("8a22d183"),
		encodedAddress:   Hex2Bytes("313373336d797a71586e6f5a5165364c416343474343566d61544875755859714345"),
		address:          "13s3myzqXnoZQe6LAcCGCCVmaTHuuXYqCE",
	},
	{
		pubkey:           Hex2Bytes("516fda9d1fdb513af7c722bdd1988d41137f9a83355944ccc9d77087d7e2242ab0d609b148ce124ceaa179358f255627db349b6bc46ddef6119f7f50f75d2bcb"),
		pubKeyHash:       Hex2Bytes("9980e583501ac027023e43544dec75e67c36a644"),
		version:          byte(0x00),
		versionedPayload: Hex2Bytes("009980e583501ac027023e43544dec75e67c36a644"),
		checksum:         Hex2Bytes("ae8ce329"),
		encodedAddress:   Hex2Bytes("31457a656f4570666d79754a586d386666324c6d4c59387678396d5a363652374570"),
		address:          "1EzeoEpfmyuJXm8ff2LmLY8vx9mZ66R7Ep",
	},
}

var invalidAddresses = []string{
	"1CFQMcVSfsMEDW14y7DRbAM3J9F6KBgF", // address is not 25 bytes long
	"14hWt7Snsg8fNB8rW5ZRkUVoMR1zuv4QV",
	"huehuehuehuehuehuehuehuehuhehuehuehhuehuehue",
	"444hWt7Snsg8fNB8rW5ZRkUVkkkkzuv4QV", // bad sha256 checksum
	"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
	"IIDSDdsdsidsoiajsdoijdsoijsijdijdi",
	"14hWt7Snsg8fNB8rW5ZRkUVoMR1zuv4QI", // invalid base58 characters (I,l,0,O)
	"14hWt7Snsg8fNB8rW5ZRkUV0MR1zuv4QV",
	"14hWt7Snsg8fNB8rW5ZRkUVOMRlzuv4QV",
}

func TestGetAddress(t *testing.T) {
	for i := 0; i < len(addressTable); i++ {
		t.Run(addressTable[i].address, func(t *testing.T) {
			addr := GetAddress(addressTable[i].pubkey)
			if !bytes.Equal(addr, addressTable[i].encodedAddress) {
				t.Errorf("expected address: %x, but got: %x", addressTable[i].encodedAddress, addr)
			}
		})
	}
}

func TestHashPubKey(t *testing.T) {
	for i := 0; i < len(addressTable); i++ {
		t.Run(addressTable[i].address, func(t *testing.T) {
			pubKeyHash := HashPubKey(addressTable[i].pubkey)
			if !bytes.Equal(pubKeyHash, addressTable[i].pubKeyHash) {
				t.Errorf("expected pubKeyHash: %x, but got: %x", addressTable[i].pubKeyHash, pubKeyHash)
			}
			versionedPayload := append([]byte{addressTable[i].version}, pubKeyHash...)
			checksumPayload := append(versionedPayload, addressTable[i].checksum...)
			encodedAddress := Base58Encode(checksumPayload)
			if !bytes.Equal(encodedAddress, addressTable[i].encodedAddress) {
				t.Errorf("expected the encoded address: %x using the pubkeyhash, but got: %x", addressTable[i].encodedAddress, encodedAddress)
			}
		})
	}
}

func TestGetPubKeyHashFromAddress(t *testing.T) {
	for i := 0; i < len(addressTable); i++ {
		t.Run(addressTable[i].address, func(t *testing.T) {
			pubKeyHash := GetPubKeyHashFromAddress(addressTable[i].address)
			if !bytes.Equal(pubKeyHash, addressTable[i].pubKeyHash) {
				t.Errorf("expected pubKeyHash: %x, but got: %x", addressTable[i].pubKeyHash, pubKeyHash)
			}
		})
	}
}

func TestValidateAddress(t *testing.T) {
	for i := 0; i < len(addressTable); i++ {
		t.Run(addressTable[i].address, func(t *testing.T) {
			if !ValidateAddress(addressTable[i].address) {
				t.Fatalf("expect address %x to be valid", addressTable[i].address)
			}
		})
	}
}

func TestInvalidAddresses(t *testing.T) {
	for _, addr := range invalidAddresses {
		t.Run(addr, func(t *testing.T) {
			if ValidateAddress(addr) {
				t.Fatalf("expect address %s to be invalid", addr)
			}
		})
	}
}

func TestChecksum(t *testing.T) {
	for i := 0; i < len(addressTable); i++ {
		t.Run(addressTable[i].address, func(t *testing.T) {
			checksum := checksum(addressTable[i].versionedPayload)
			if !bytes.Equal(checksum, addressTable[i].checksum) {
				t.Errorf("expected checksum: %x, but got: %x", addressTable[i].checksum, checksum)
			}
		})
	}
}

func TestNewKeyPair(t *testing.T) {
	privKey, pubKey := newKeyPair()

	if len(pubKey) == 0 {
		t.Fatal("newKeyPair returned an unexpected result")
	}

	assert.Equalf(t, append(privKey.PublicKey.X.Bytes(), privKey.PublicKey.Y.Bytes()...), pubKey, "The public key should be represented as a concatenation of it's coordinates")
}

func TestPubKeyToByte(t *testing.T) {
	pubkey := decodePublicKey(testEncPubKeyUser1)

	assert.Equalf(t, Hex2Bytes("f86aa0caf08359ee4227d2901ab490172c69a801910f4140cdde2f5dc8f8bb3dc19da2c9fb0ed041db106a8fea0382de25edbc83df6893574e40fc2e1e493748"), pubKeyToByte(*pubkey), "The public key should be represented as a concatenation of it's coordinates")
}
