package main

import (
	"fmt"
	"log"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/smartcontract/blockchain"
	"github.com/smartcontract/cmd/configuration"
)

const END_POINT string = "https://ropsten.infura.io/"

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	config := configuration.NewConfiguration()
	client, err := rpc.Dial(END_POINT)
	if err != nil {
		fmt.Println(err)
	}
	ethclient := ethclient.NewClient(client)
	bc := blockchain.NewBlockchain(client, ethclient, config.Signer, config.Contract, config.AbiPath)
	bc.GetTreasure()
	owner := "0x2cc72d8857ac57ba058eddd36b2f14adc2a058bd"
	var amount int64 = 100

	amountBig := big.NewInt(amount)

	tx, err := bc.AddTreasure(owner, amountBig)
	if err != nil {
		log.Println(err)
	}
	log.Println("raw tx: ", tx, tx.Hash().String())
}
