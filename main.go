package main

import (
	"encoding/json"
	"log"
	"time"

	"github.com/joho/godotenv"
	coinbasepro "github.com/preichenberger/go-coinbasepro/v2"
)

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Println("Error loading .env file")
	}
	exit = false
}

var (
	exit            bool
	messages        chan coinbasepro.Message
	heartbeat       chan string
	limitOrderTimer *time.Timer
)

func main() {
	heartbeat = make(chan string)
	messages = make(chan coinbasepro.Message)
	orderExpiry := 5 * time.Minute

	client := coinbasepro.NewClient()
	defer func() {
		msg, err := client.CancelAllOrders()
		if err != nil {
			log.Println(err.Error())
		}
		log.Println(msg)
	}()
	go Websocket(client)

	orderInterval := GetOrderInterval()
	ticker := time.NewTimer(orderInterval)
	go func() {
		for {
			<-ticker.C
			orderInterval = GetOrderInterval()
			ticker.Reset(GetOrderInterval())
			err := LoopOrder(client, orderExpiry)
			if err != nil {
				log.Println(err.Error())
				exit = true
				return
			}
			err = Withdraw(client)
			if err != nil {
				log.Println(err.Error())
				exit = true
				return
			}
		}
	}()

	for {
		if exit {
			log.Println("Closing server.")
			return
		}
		select {
		case msg := <-messages:
			if msg.Reason == "filled" {
				limitOrderTimer.Reset(orderExpiry)
				if msg.RemainingSize == "0.00000000" {
					limitOrderTimer.Reset(2 * time.Second)
				}
			}
			pretty, err := json.MarshalIndent(msg, "", "  ")
			if err != nil {
				log.Println(err.Error())
			}
			log.Println(string(pretty))
		case msg := <-heartbeat:
			now := time.Now()
			if now.Hour() == 0 && now.Minute() == 0 {
				log.Println(msg)
			}
		default:
		}
	}
}
