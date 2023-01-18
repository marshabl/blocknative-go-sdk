package bnsdk

import (
	"time"
)

type CategoryCode string
type EventCode string
type System string
type Network int

// Blockchain is a type fulfilling the blockchain params
type Blockchain struct {
	System  System `json:"system"`
	Network string `json:"network"`
}

// BaseMessage is the base message required for all interactions with the websockets api
type BaseMessage struct {
	CategoryCode CategoryCode `json:"categoryCode"`
	EventCode    EventCode    `json:"eventCode"`
	Timestamp    time.Time    `json:"timeStamp"`
	DappID       string       `json:"dappId"` // api key
	Version      string       `json:"version"`
	Blockchain   `json:"blockchain"`
}

// AuthRepsonse is the message we receive when opening a connection to the API
type AuthRepsonse struct {
	ConnectionID  string `json:"connectionId"`
	ServerVersion string `json:"serverVersion"`
	ShowUX        bool   `json:"showUX"`
	Status        string `json:"status"`
	Reason        string `json:"reason,omitempty"`
	Version       int    `json:"version"`
}

// Account bundles a single account address
type Account struct {
	Address string `json:"address,omitempty"`
}

// Transaction bundles a single tx
type Transaction struct {
	Hash string `json:"hash"`
}

// Config provides a specific config instance
type Config struct {
	//  valid Ethereum address or 'global'
	Scope string `json:"scope,omitempty"`
	// A slice of valid filters (jsql: https://github.com/deitch/searchjs)
	Filters []map[string]string `json:"filters,omitempty"`
	// JSON abis
	ABI interface{} `json:"abi,omitempty"`
	// defines whether the service should automatically watch the address as defined in
	WatchAddress bool `json:"watchAddress,omitempty"`
}
