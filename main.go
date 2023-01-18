package main

import (
	"log"
	"os"

	"github.com/ethereum/go-ethereum/common"
)

const SYSTEM = "ethereum"
const NETWORK = 1
const MONITORADDRESS = "0x68b3465833fb72a70ecdf485e0e4c7bd8665fc45" //uniswap autorouter

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
	config := NewConfig("global", filters)

	sub := NewSubscription(common.HexToAddress(MONITORADDRESS), config)
	client, err := Stream(APIKEY, SYSTEM, NETWORK)
	if err != nil {
		log.Printf("error: %v", err.Error())
		return
	}
	client.Subscribe(sub, txnHandler)
}
