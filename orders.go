package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	coinbasepro "github.com/preichenberger/go-coinbasepro/v2"
	"github.com/shopspring/decimal"
)

// GetAccount loads all accounts and returns the currency queried
func GetAccount(c *coinbasepro.Client, currency string) coinbasepro.Account {
	accounts, err := c.GetAccounts()
	if err != nil {
		panic(err)
	}
	for _, a := range accounts {
		if a.Currency == currency {
			return a
		}
	}
	panic("No account for " + currency)
}

// limitOrder sends a limit order based on the amount of funds available in your account or
// otherwise the minimum of 0.001 BTC
func limitOrder(c *coinbasepro.Client) coinbasepro.Order {
	var order coinbasepro.Order
	productID := os.Getenv("PRODUCT_ID")
	productIDs := strings.Split(productID, "-")
	fiat := productIDs[1]
	book, err := c.GetBook(productID, 1)
	if err != nil {
		panic(err)
	}

	lastPrice, err := decimal.NewFromString(book.Bids[0].Price)
	if err != nil {
		panic(err)
	}
	buyAmount, err := decimal.NewFromString(os.Getenv("BUY_AMOUNT"))
	if err != nil {
		panic(err)
	}
	size := "0.001"
	if buyAmount.Div(lastPrice).GreaterThan(decimal.NewFromFloat(0.001)) {
		size = buyAmount.Div(lastPrice).StringFixed(3)
	}
	sizePrecise, err := decimal.NewFromString(size)
	if err != nil {
		panic(err)
	}

	account := GetAccount(c, fiat)
	available, err := decimal.NewFromString(account.Available)
	if err != nil {
		panic(err)
	}
	if sizePrecise.Mul(lastPrice).GreaterThan(available) {
		panic("Not enough funds available")
	}

	orderOpts := coinbasepro.Order{
		Type:  "limit",
		Price: lastPrice.Add(decimal.NewFromFloat(1.00)).String(),
		// Price:     "10",
		Size:      size,
		Side:      "buy",
		ProductID: productID,
	}

	order, err = c.CreateOrder(&orderOpts)
	if err != nil {
		panic(err)
	}
	return order
}

//LoopOrder creates a limit order and if it is not filled within the orderExpiry time
//then a new order will be set at the top bid of the orderbook
func LoopOrder(c *coinbasepro.Client) bool {
	order := limitOrder(c)
	println("order placed")
	limitOrderTimer = time.NewTimer(orderExpiry)
	<-limitOrderTimer.C
	order, err := c.GetOrder(order.ID)
	if err != nil {
		panic(err)
	}
	if !order.Settled {
		err = c.CancelOrder(order.ID)
		if err != nil {
			panic(err)
		}
		println("order expired")
		return LoopOrder(c)
	}
	println("order settled")
	return true
}

// Withdraw withdraws all your selected crypto into the $COLD_WALLET
func Withdraw(c *coinbasepro.Client) {
	productID := os.Getenv("PRODUCT_ID")
	productIDs := strings.Split(productID, "-")
	coin := productIDs[0]
	account := GetAccount(c, coin)
	withdrawOpts := coinbasepro.WithdrawalCrypto{
		Currency: coin,
		Amount:   account.Available,
		// Amount:        "0.001",
		CryptoAddress: os.Getenv("COLD_WALLET"),
	}
	_, err := c.CreateWithdrawalCrypto(&withdrawOpts)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Withdrew %s %s", account.Available, coin)
}
