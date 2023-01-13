package main

import (
	bnsdk "blocknative-sdk/pkg"
	"context"
	"log"
	"os"

	"github.com/ethereum/go-ethereum/common"
	"github.com/gorilla/websocket"
)

const MONITORADDRESS = "0x68b3465833fb72a70ecdf485e0e4c7bd8665fc45" //uniswap autorouter

func main() {
	ctx := context.Background()
	conn, _, err := websocket.DefaultDialer.DialContext(
		ctx,
		"wss://api.blocknative.com/v0",
		nil,
	)

	if err != nil {
		log.Printf("Failed to create websocket connection.")
		return
	}

	APIKEY, ok := os.LookupEnv("APIKEY")
	if !ok {
		log.Printf("No environment variable APIRKEY.")
		return

	}

	client := &bnsdk.Client{
		Conn:    conn,
		DappID:  APIKEY,
		System:  "ethereum",
		Network: 1,
	}

	if err := client.Authenticate(ctx); err != nil {
		log.Printf("Failed to authenticate your APIKEY.")
		return
	}

	var filters []map[string]string
	filter := map[string]string{
		"status": "pending",
	}
	filters = append(filters, filter)
	config := bnsdk.NewConfig("global", filters)

	sub := bnsdk.NewSubscription(common.HexToAddress(MONITORADDRESS), config)
	client.Subscribe(ctx, sub.Subscriber, sub.Configs...)

	go client.Reader(sub)
	go client.Writer()

	for {
		e, ok := <-sub.EventChan
		if !ok {
			break
		}
		log.Printf("%v", string(e))
	}

}
