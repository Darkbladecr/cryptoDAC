package main

import (
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	coinbasepro "github.com/preichenberger/go-coinbasepro/v2"
)

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Println("Error loading .env file")
	}
	goEnv = os.Getenv("GO_ENV")
	exit = false
}

var (
	goEnv           string
	exit            bool
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
	defer func() {
		msg, err := client.CancelAllOrders()
		if err != nil {
			log.Println(err.Error())
		}
		log.Println(msg)
	}()
	go Websocket(client)

	ticker := time.NewTimer(orderInterval)
	go func() {
		for {
			<-ticker.C
			orderInterval = GetOrderInterval()
			ticker.Reset(GetOrderInterval())
			err := LoopOrder(client)
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
		if goEnv != "PRODUCTION" {
			log.Println(<-messages)
		}
	}
}
