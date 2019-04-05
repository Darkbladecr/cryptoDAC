package main

import (
	"fmt"
	"log"
	"os"

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
	log.Printf("Starting stream on %s", wssURI)

	var wsDialer ws.Dialer
	wsConn, _, err := wsDialer.Dial(wssURI, nil)
	if err != nil {
		log.Println(err.Error())
	}
	defer wsConn.Close()
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
	signedMessage, err := message.Sign(
		os.Getenv("COINBASE_PRO_SECRET"),
		os.Getenv("COINBASE_PRO_KEY"),
		os.Getenv("COINBASE_PRO_PASSPHRASE"))
	if err != nil {
		log.Println(err.Error())
	}
	if err := wsConn.WriteJSON(signedMessage); err != nil {
		log.Println(err.Error())
	}

	for {
		if exit {
			return
		}
		message := coinbasepro.Message{}
		if err := wsConn.ReadJSON(&message); err != nil {
			log.Println(err.Error())
			return
		}
		if message.Type == "heartbeat" {
			time := message.Time.Time()
			heartbeat <- fmt.Sprintf("Ping %s @ %s",
				message.ProductID,
				time.Format("2006-01-02 15:04:05"))
		} else if message.Type != "subscriptions" {
			messages <- message
		}
	}
}
