package bnsdk

import (
	"github.com/ethereum/go-ethereum/common"
)

// // Subscription represents a stream of events. Implementations
// // carry a channel with which to store events returned by the subscription backend
type Subscriber interface {
	Message() (CategoryCode, EventCode)
	SubscriptionMessage(BaseMessage) *SubscriptionMessage
}

type SubscriptionMessage struct {
	BaseMessage
	Account     *Account     `json:"account,omitempty"`
	Transaction *Transaction `json:"hash,omitempty"`
	Config      *Config      `json:"config,omitempty"`
	//distributionSubscriber
	//simulationSubscriber
}

type ConfigurationMessage struct {
	BaseMessage
	Config `json:"config,omitempty"`
}

type SubscriptionOption interface {
	Hex() string
}

// Subscribe to an Address
type AddressSubscriber struct {
	Account `json:"account,omitempty"`
}

// Subscribe to a configuration/filters
type ConfigurationSubscriber struct {
	Config `json:"account,omitempty"`
}

// Subscribe to a Transaction Hash
type TransactionSubscriber struct {
	Transaction `json:"hash"`
}

type Subscription struct {
	Subscriber Subscriber // Address or Transaction
	Configs    []Config
	EventChan  chan []byte
	ErrChan    chan error
}

// NewSubscription creates a carrier for tracking events
func NewSubscription(s SubscriptionOption, configs ...Config) *Subscription {
	var subscriber Subscriber

	switch s.(type) {
	case common.Address:
		subscriber = AddressSubscriber{
			Account: Account{Address: s.Hex()},
		}
	case common.Hash:
		subscriber = TransactionSubscriber{
			Transaction: Transaction{Hash: s.Hex()},
		}
	default:
		//TODO
	}

	return &Subscription{
		Subscriber: subscriber,
		Configs:    configs,
		EventChan:  make(chan []byte),
		ErrChan:    make(chan error, 1),
	}
}

// NewConfig returns a new config instance
func NewConfig(scope string, filters []map[string]string, abis ...interface{}) Config {
	watchAddress := true
	if scope == "global" {
		watchAddress = false
	}

	cfg := Config{
		Scope:        scope,
		Filters:      filters,
		WatchAddress: watchAddress,
	}

	if abis != nil {
		cfg.ABI = abis
	}

	return cfg
}

func (c ConfigurationSubscriber) Message() (CategoryCode, EventCode) {
	return "configs", "put"
}

func (c ConfigurationSubscriber) SubscriptionMessage(b BaseMessage) *SubscriptionMessage {
	return &SubscriptionMessage{
		BaseMessage: b,
		Config:      &c.Config,
	}
}

func (a AddressSubscriber) Message() (CategoryCode, EventCode) {
	return "accountAddress", "watch"
}

func (a AddressSubscriber) SubscriptionMessage(b BaseMessage) *SubscriptionMessage {
	return &SubscriptionMessage{
		BaseMessage: b,
		Account:     &a.Account,
	}
}

func (t TransactionSubscriber) Message() (CategoryCode, EventCode) {
	return "accountAddress", "watch"
}

func (t TransactionSubscriber) SubscriptionMessage(b BaseMessage) *SubscriptionMessage {
	return &SubscriptionMessage{
		BaseMessage: b,
		Transaction: &t.Transaction,
	}
}
