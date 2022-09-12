package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

var (
	bc    *Blockchain
	txns  []*Transaction
	utxos UTXOSet
)

func main() {
	utxos = make(UTXOSet)
	a := CreateIdentities()
	b := CreateIdentities()
	c := CreateIdentities()
	nonce := 0
	cbReward, err := NewCoinbaseTX(a.address, "COIN Reward"+strconv.Itoa(nonce))
	if err != nil {
		panic(err)
	}
	txns = []*Transaction{cbReward}

	for {
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("Please choose the corresponding request number\n")
		fmt.Print(QUERY)
		text, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Plase make sure a blockchain is created and try again!")
		}
		text = strings.Trim(text, "\n")
		switch text {
		case "1":
			if bc != nil {
				fmt.Println("There is already one blockchain created!")
				continue
			}
			bc, err = NewBlockchain(string(a.address))
			if err != nil {
				fmt.Println("Could not generate the chain!")
			}
			fmt.Println("New block created, and miner got his reward!")
			fmt.Println()
			utxos.Update(bc.GetGenesisBlock().Transactions)
		case "2":
			fmt.Println()
			fmt.Println("We trasnfer the miner's reward from a to b.")
			txn, err := NewUTXOTransaction(a.pubkey, b.address, BlockReward, utxos)
			if err != nil {
				fmt.Println(err)
				continue
			}
			bc.SignTransaction(txn, a.pk)
			txns = append(txns, txn)
			fmt.Println("Transfered!")
		case "3":
			_, err = bc.MineBlock(txns)
			if err != nil {
				fmt.Println("Plase make sure a blockchain is created and try again!")
				continue
			}
			utxos.Update(txns)
			cbReward, err := NewCoinbaseTX(a.address, "COIN Reward"+strconv.Itoa(nonce+1))
			if err != nil {
				panic(err)
			}
			txns = []*Transaction{cbReward}
		case "4":
			for i, b := range bc.blocks {
				fmt.Printf("%d: %s\n", i+1, b.String())
			}
		case "5":
			fmt.Println()
			fmt.Print("Please choose the corresponding block number given its hash\n")
			for i, b := range bc.blocks {
				fmt.Printf("#%d: Hash: %x\n", i+1, b.Hash)
			}
			fmt.Println()
			userIdx, err := reader.ReadString('\n')
			if err != nil {
				fmt.Println("Plase make sure a blockchain is created and try again!")
				continue
			}
			userIdx = strings.Trim(userIdx, "\n")
			idxToInt, err := strconv.Atoi(userIdx)
			if err != nil {
				fmt.Println("Could not process your request, please try again!")
				continue
			}
			b := bc.blocks[idxToInt-1]
			fmt.Printf("All information corresponding to the block number: %s\n", userIdx)
			fmt.Println(b.String())
			fmt.Println()
		case "6":
			fmt.Println("Please choose from the txs from the list below providing the required information!")
			for i, b := range bc.blocks {
				for j, t := range b.Transactions {
					fmt.Printf("#Block: %d, #Txn: %d, #ID: %x\n", i+1, j+1, t.ID)
				}
			}
			fmt.Println()
			fmt.Println("Your chosen #Block:")
			bNr, err := reader.ReadString('\n')
			if err != nil {
				fmt.Println("Plase make sure a blockchain is created and try again!")
				continue
			}
			fmt.Println("Your chosen #Txn:")
			txnNr, err := reader.ReadString('\n')
			if err != nil {
				fmt.Println("Plase make sure a blockchain is created and try again!")
				continue
			}
			bNr = strings.Trim(bNr, "\n")
			txnNr = strings.Trim(txnNr, "\n")
			bNrToInt, err := strconv.Atoi(bNr)
			if err != nil {
				fmt.Println("Could not process your request, please try again!")
				continue
			}
			txnNrToInt, err := strconv.Atoi(txnNr)
			if err != nil {
				fmt.Println("Could not process your request, please try again!")
				continue
			}
			t := bc.blocks[bNrToInt-1].Transactions[txnNrToInt-1]
			txn, err := bc.FindTransaction(t.ID)
			if err != nil {
				fmt.Println("Could not process your request, please try again!")
				continue
			}
			fmt.Printf("Transaction with the ID %x, has the following data: %s\n", txn.ID, txn.String())
			fmt.Println()
		case "7":
			fmt.Println("We attempt to transfer some coins from c to either b or a.")
			txn, err := NewUTXOTransaction(c.pubkey, b.address, 2, utxos)
			if err != nil {
				fmt.Println(err)
				continue
			}
			txns = append(txns, txn)
			utxos.Update(txns)
			fmt.Println("Transfered!")
		case "8":
			fmt.Println("We attempt to transfer 5 coins from b to c")
			txn, err := NewUTXOTransaction(b.pubkey, c.address, 5, utxos)
			if err != nil {
				fmt.Println(err)
				continue
			}
			err = bc.SignTransaction(txn, b.pk)
			if err != nil {
				fmt.Println(err)
				continue
			}
			txns = append(txns, txn)
			fmt.Println("Transfered!")
		case "9":
			a := getBalance(a.address, utxos)
			b := getBalance(b.address, utxos)
			c := getBalance(c.address, utxos)
			fmt.Printf("Balance of address: %s, is: %d\n", a.Address, a.Funds)
			fmt.Printf("Balance of address: %s, is: %d\n", b.Address, b.Funds)
			fmt.Printf("Balance of address: %s, is: %d\n", c.Address, c.Funds)
		case "10":
			fmt.Println(utxos.String())
		default:
			continue
		}
	}
}
