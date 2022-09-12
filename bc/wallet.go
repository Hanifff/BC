package main

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"fmt"

	"golang.org/x/crypto/ripemd160"
)

const (
	version            = byte(0x00)
	addressChecksumLen = 4
)

// newKeyPair creates a new cryptographic key pair
func newKeyPair() (ecdsa.PrivateKey, []byte) {
	pk, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		fmt.Printf("Could not generate the ecdsa key pair!\n%v\n", err)
		return ecdsa.PrivateKey{}, nil
	}
	pubKey := pubKeyToByte(pk.PublicKey)
	return *pk, pubKey
}

// pubKeyToByte converts the ecdsa.PublicKey to a concatenation of its coordinates in bytes
func pubKeyToByte(pubkey ecdsa.PublicKey) []byte {
	if pubkey.X == nil || pubkey.Y == nil {
		return nil
	}
	//fmt.Printf("%v\n",elliptic.Marshal(pubkey, pubkey.X, pubkey.Y)[0])
	var xy []byte
	xy = append(xy, pubkey.X.Bytes()...)
	xy = append(xy, pubkey.Y.Bytes()...)
	return xy
	/* ignoreFirstBit := elliptic.Marshal(pubkey, pubkey.X, pubkey.Y)[1:] // an other method used by rodrigo!?
	return ignoreFirstBit */
}

// GetAddress returns address
// https://en.bitcoin.it/wiki/Technical_background_of_version_1_Bitcoin_addresses#How_to_create_Bitcoin_Address
func GetAddress(pubKeyBytes []byte) []byte {
	var versionedAddress []byte
	address := HashPubKey(pubKeyBytes)
	versionedAddress = append(versionedAddress, version)
	versionedAddress = append(versionedAddress, address...)
	csAddress := checksum(versionedAddress)
	versionedAddress = append(versionedAddress, csAddress...)
	BCaddress := Base58Encode(versionedAddress)
	return BCaddress
}

// GetStringAddress returns address as string
func GetStringAddress(pubKeyBytes []byte) string {
	//return fmt.Sprintf("%x", GetAddress(pubKeyBytes))
	return ""
}

// HashPubKey hashes public key
func HashPubKey(pubKey []byte) []byte {
	shaHash := sha256.Sum256(pubKey)
	wrapper := ripemd160.New()
	wrapper.Write(shaHash[:])
	return wrapper.Sum(nil)
}

// GetPubKeyHashFromAddress returns the hash of the public key
// discarding the version and the checksum
func GetPubKeyHashFromAddress(address string) []byte {
	BCaddress := Base58Decode([]byte(address))
	ignoreCS := BCaddress[:len(BCaddress)-4]
	ignoreVersion := ignoreCS[1:]
	return ignoreVersion
}

// ValidateAddress check if an address is valid
func ValidateAddress(address string) bool {
	BCaddress := Base58Decode([]byte(address))
	// extract the checksum to get back the versioned payload
	recomputeAdd := BCaddress[:len(BCaddress)-4]
	recomputeCS := checksum(recomputeAdd)
	return bytes.Equal(BCaddress[len(BCaddress)-4:], recomputeCS)
}

// Checksum generates a checksum for a public key
func checksum(payload []byte) []byte {
	firstHash := sha256.Sum256(payload)
	wrappHash := sha256.New()
	_, err := wrappHash.Write(firstHash[:])
	if err != nil {
		fmt.Printf("Could not calculate the checksum of payload.\n%v\n", err)
		return nil
	}
	cs := wrappHash.Sum(nil)[:4]
	return cs
}

func encodeKeyPair(privateKey *ecdsa.PrivateKey, publicKey *ecdsa.PublicKey) (string, string) {
	return encodePrivateKey(privateKey), encodePublicKey(publicKey)
}

func encodePrivateKey(privateKey *ecdsa.PrivateKey) string {
	x509Encoded, _ := x509.MarshalECPrivateKey(privateKey)
	pemEncoded := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: x509Encoded})

	return string(pemEncoded)
}

func encodePublicKey(publicKey *ecdsa.PublicKey) string {
	x509EncodedPub, _ := x509.MarshalPKIXPublicKey(publicKey)
	pemEncodedPub := pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: x509EncodedPub})

	return string(pemEncodedPub)
}

func decodeKeyPair(pemEncoded string, pemEncodedPub string) (*ecdsa.PrivateKey, *ecdsa.PublicKey) {
	return decodePrivateKey(pemEncoded), decodePublicKey(pemEncodedPub)
}

func decodePrivateKey(pemEncoded string) *ecdsa.PrivateKey {
	block, _ := pem.Decode([]byte(pemEncoded))
	privateKey, _ := x509.ParseECPrivateKey(block.Bytes)

	return privateKey
}

func decodePublicKey(pemEncodedPub string) *ecdsa.PublicKey {
	blockPub, _ := pem.Decode([]byte(pemEncodedPub))
	genericPubKey, _ := x509.ParsePKIXPublicKey(blockPub.Bytes)
	publicKey := genericPubKey.(*ecdsa.PublicKey) // cast to ecdsa

	return publicKey
}
