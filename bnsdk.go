package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/gorilla/websocket"
)

const WSSURL = "wss://api.blocknative.com/v0"

const (
	Main   = 1
	Goerli = 5
	Matic  = 137
	Mumbai = 80001
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 5 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 10 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = pongWait / 2

	// retry to establish a new connection
	retryWait = 10 * time.Second
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

func Stream(dappId string, system System, network Network) (*Client, error) {
	ctx := context.Background()
	conn, _, err := websocket.DefaultDialer.DialContext(ctx, WSSURL, nil)

	if err != nil {
		log.Printf("Failed to create websocket connection.")
		return nil, err
	}

	client := &Client{
		Conn:    conn,
		DappID:  dappId,
		System:  system,
		Network: network,
	}

	if err := client.authenticate(ctx); err != nil {
		log.Printf("Failed to authenticate your APIKEY.")
		return nil, err
	}

	return client, nil
}

func (c Client) authenticate(ctx context.Context) error {
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

func (c Client) writer(sub *Subscription) {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
	c.Conn.SetPingHandler(func(string) error {
		c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
		return nil
	})

	for {
		<-ticker.C
		if err := c.write(websocket.PingMessage, []byte{}); err != nil {
			sub.ErrChan <- err
			return
		}
	}
}

func (c Client) reader(sub *Subscription) {
	defer func() {
		c.Conn.Close()
	}()

	c.Conn.SetReadDeadline(time.Now().Add(pongWait))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})
	for {
		_, msg, err := c.Conn.ReadMessage()
		if err != nil {
			sub.ErrChan <- err
			// if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway) {
			// 	log.Printf("READ error: %v", err.Error())
			// }
			// log.Printf("READ error: %v", err.Error())
			// break
		}
		sub.EventChan <- msg
	}
}

func (c Client) retryConnection(sub *Subscription, txnHandler func([]byte)) {
	ticker := time.NewTicker(retryWait)
	newSub := sub
	for {
		<-ticker.C
		newClient, err := Stream(c.DappID, c.System, c.Network)
		if err != nil {
			log.Printf("error: %v", err.Error())
		} else {
			newClient.Subscribe(newSub, txnHandler)
			return
		}
	}
}

func (c Client) Subscribe(sub *Subscription, txnHandler func([]byte)) {
	for _, config := range sub.Configs {
		configSubscriber := ConfigurationSubscriber{config}
		configurationMessage := configSubscriber.SubscriptionMessage(c.buildBaseMessage(configSubscriber))
		if err := c.Conn.WriteJSON(configurationMessage); err != nil {
			sub.ErrChan <- err
		}
	}

	subscriptionMessage := sub.Subscriber.SubscriptionMessage(c.buildBaseMessage(sub.Subscriber))
	if err := c.Conn.WriteJSON(subscriptionMessage); err != nil {
		sub.ErrChan <- err
	}

	go c.reader(sub)
	go c.writer(sub)

	for {
		select {
		case err := <-sub.ErrChan:
			log.Printf("MAIN error: %v", err.Error())
			c.retryConnection(sub, txnHandler)
			return
		default:
			e, ok := <-sub.EventChan
			if ok {
				txnHandler(e)
			}
		}
	}
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
