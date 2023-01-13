package bnsdk

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/gorilla/websocket"
)

const (
	Main   = 1
	Goerli = 5
	Matic  = 137
	Mumbai = 80001
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 10 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = pongWait / 2

	// Maximum message size allowed from peer.
	maxMessageSize = 512
)

const Version = "1.1"

type Conn interface {
	ReadJSON(interface{}) error
	WriteJSON(interface{}) error
	Close() error
	ReadMessage() (int, []byte, error)
	SetPingHandler(func(string) error)
	SetPongHandler(func(string) error)
	SetWriteDeadline(time.Time) error
	SetReadDeadline(time.Time) error
	WriteMessage(int, []byte) error
}

type Client struct {
	Conn    Conn
	DappID  string
	System  System
	Network Network
}

func (n Network) String() string {
	switch n {
	case Main:
		return "main"
	case Goerli:
		return "goerli"
	case Matic:
		return "matic-main"
	case Mumbai:
		return "mumbai"
	default:
		return fmt.Sprintf("Network(%d)", n)
	}
}

func (c Client) Authenticate(ctx context.Context) error {
	authmsg := &BaseMessage{
		Version:      Version,
		CategoryCode: "initialize",
		EventCode:    "checkDappId",
		Timestamp:    time.Now(),
		DappID:       c.DappID,
		Blockchain: Blockchain{
			System:  c.System,
			Network: c.Network.String(),
		},
	}

	return c.Conn.WriteJSON(authmsg)
}

// write writes a message with the given message type and payload.
func (c Client) write(mt int, payload []byte) error {
	c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
	return c.Conn.WriteMessage(mt, payload)
}

// write writes a message with the given message type and payload.
func (c Client) read() ([]byte, error) {
	_, msg, err := c.Conn.ReadMessage()
	if err != nil {
		log.Printf("error: %v", err)
		return msg, err
	}
	return msg, nil
}

func (c Client) Writer() {
	ticker := time.NewTicker(pingPeriod)
	for {
		select {
		case <-ticker.C:
			if err := c.write(websocket.PingMessage, []byte{}); err != nil {
				log.Printf("error: %v", err)
				break
			}
		}
	}
}

func (c Client) Reader(sub *Subscription) {
	defer func() {
		c.Conn.Close()
		// TODO
	}()

	c.Conn.SetReadDeadline(time.Now().Add(pongWait))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})
	for {
		select {
		case <-sub.ErrChan:
			return
		default:
			msg, err := c.read()
			if err != nil {
				log.Printf("error: %v", err)
				return
			}
			sub.EventChan <- msg
		}
	}
}

func (c Client) Subscribe(ctx context.Context, s Subscriber, configs ...Config) error {
	for _, config := range configs {
		configSubscriber := ConfigurationSubscriber{config}
		configurationMessage := configSubscriber.SubscriptionMessage(c.buildBaseMessage(configSubscriber))
		c.Conn.WriteJSON(configurationMessage)
	}

	subscriptionMessage := s.SubscriptionMessage(c.buildBaseMessage(s))
	return c.Conn.WriteJSON(subscriptionMessage)
}

func (c Client) buildBaseMessage(s Subscriber) BaseMessage {
	categoryCode, eventCode := s.Message()

	authmsg := BaseMessage{
		Version:      Version,
		CategoryCode: categoryCode,
		EventCode:    eventCode,
		Timestamp:    time.Now(),
		DappID:       c.DappID,
		Blockchain: Blockchain{
			System:  c.System,
			Network: c.Network.String(),
		},
	}

	return authmsg
}
