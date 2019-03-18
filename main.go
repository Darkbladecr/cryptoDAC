package main

import (
	"time"

	"github.com/joho/godotenv"
	coinbasepro "github.com/preichenberger/go-coinbasepro/v2"
)

func init() {
	err := godotenv.Load()
	if err != nil {
		println("Error loading .env file")
	}
}

var (
	messages        chan string
	limitOrderTimer *time.Timer
	orderExpiry     time.Duration
	orderInterval   time.Duration
)

func main() {
	messages = make(chan string)
	orderExpiry = 5 * time.Minute
	orderInterval = GetOrderInterval()
	client := coinbasepro.NewClient()
	defer client.CancelAllOrders()
	go Websocket(client)

	ticker := time.NewTimer(time.Second)
	go func() {
		println("ticker loaded")
		for {
			<-ticker.C
			ticker.Reset(GetOrderInterval())
			println("ticker done")
			LoopOrder(client)
			Withdraw(client)
		}
	}()

	for {
		println(<-messages)
	}
}
