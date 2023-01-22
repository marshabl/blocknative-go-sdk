package main

import (
	"log"
	"os"

	"github.com/ethereum/go-ethereum/common"
	"github.com/marshabl/blocknative-go-sdk/bnsdk"
)

const SYSTEM = "ethereum"
const NETWORK = 1
const MONITORADDRESS = "0xEf1c6E67703c7BD7107eed8303Fbe6EC2554BF6B" //uniswap universal router

func txnHandler(e []byte) {
	log.Printf("%v", string(e))
}

func main() {
	APIKEY, ok := os.LookupEnv("APIKEY")
	if !ok {
		log.Printf("No environment variable APIRKEY.")
		return
	}

	var filters []map[string]string
	filter := map[string]string{
		"status": "pending",
	}
	filters = append(filters, filter)
	config := bnsdk.NewConfig("global", filters)

	sub := bnsdk.NewSubscription(common.HexToAddress(MONITORADDRESS), config)
	client, err := bnsdk.Stream(APIKEY, SYSTEM, NETWORK)
	if err != nil {
		log.Printf("error: %v", err.Error())
		return
	}
	client.Subscribe(sub, txnHandler)
}
