package main

import (
	"log"
	"os"
	"time"

	"blocknative-go-sdk/bnsdk"

	"github.com/ethereum/go-ethereum/common"
)

const SYSTEM = "ethereum"
const NETWORK = 1
const MONITORADDRESS = "0xEf1c6E67703c7BD7107eed8303Fbe6EC2554BF6B" //uniswap universal router
const RETRYPERIOD = 10 * time.Second

func txnHandler(e []byte) {
	log.Printf("%v", string(e))
}

func errHandler(e error) {
	log.Printf("BLAIR: %v", e)
}

func closeHandler(APIKEY string, sub *bnsdk.Subscription) {
	ticker := time.NewTicker(RETRYPERIOD)
	for {
		<-ticker.C
		log.Printf("Attempting to start a new connection...")
		err := startSubscription(APIKEY, sub)
		if err == nil {
			return
		}
	}
}

func startSubscription(APIKEY string, sub *bnsdk.Subscription) error {
	client, err := bnsdk.Stream(APIKEY, SYSTEM, NETWORK)
	if err != nil {
		log.Printf("error: %v", err.Error())
		return err
	}
	return client.Subscribe(sub)
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

	sub := bnsdk.NewSubscription(
		common.HexToAddress(MONITORADDRESS),
		txnHandler,
		errHandler,
		closeHandler,
		config,
	)
	startSubscription(APIKEY, sub)
}
