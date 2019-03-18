package main

import (
	"encoding/json"
	"os"
	"time"

	ws "github.com/gorilla/websocket"
	coinbasepro "github.com/preichenberger/go-coinbasepro/v2"
)

var wssURI string

func init() {
	wssURI = "wss://ws-feed.pro.coinbase.com"
	if sandbox, ok := os.LookupEnv("COINBASE_PRO_SANDBOX"); ok && sandbox == "1" {
		wssURI = "wss://ws-feed-public.sandbox.pro.coinbase.com"
	}
}

//Websocket connects to coinbase's stream to watch order flow
func Websocket(c *coinbasepro.Client) {
	defer Websocket(c) // restart stream if breaks
	var wsDialer ws.Dialer
	wsConn, _, err := wsDialer.Dial(wssURI, nil)
	if err != nil {
		println(err.Error())
	}
	message := coinbasepro.Message{
		Type:       "subscribe",
		ProductIds: []string{os.Getenv("PRODUCT_ID")},
		Channels: []coinbasepro.MessageChannel{
			coinbasepro.MessageChannel{
				Name: "heartbeat",
			},
			coinbasepro.MessageChannel{
				Name: "user",
			},
		},
	}
	signedMessage, err := message.Sign(os.Getenv("COINBASE_PRO_SECRET"), os.Getenv("COINBASE_PRO_KEY"), os.Getenv("COINBASE_PRO_PASSPHRASE"))
	if err != nil {
		println(err.Error())
	}
	if err := wsConn.WriteJSON(signedMessage); err != nil {
		println(err.Error())
	}

	messages <- "starting stream"

	for true {
		message := coinbasepro.Message{}
		if err := wsConn.ReadJSON(&message); err != nil {
			println(err.Error())
			break
		}
		if message.Type != "heartbeat" {
			if msg, err := json.MarshalIndent(message, "", "  "); err != nil {
				panic(err)
			} else {
				messages <- string(msg)
			}
			if message.Reason == "filled" {
				limitOrderTimer.Reset(orderExpiry)
				if message.RemainingSize == "0.00000000" {
					limitOrderTimer.Reset(2 * time.Second)
				}
			}
		}
	}
}
