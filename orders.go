package main

import (
	"errors"
	"log"
	"os"
	"strings"
	"time"

	coinbasepro "github.com/preichenberger/go-coinbasepro/v2"
	"github.com/shopspring/decimal"
)

// GetAccount loads all accounts and returns the currency queried
func GetAccount(c *coinbasepro.Client, currency string) (coinbasepro.Account, error) {
	var account coinbasepro.Account
	accounts, err := c.GetAccounts()
	if err != nil {
		return account, err
	}
	for _, a := range accounts {
		if a.Currency == currency {
			return a, nil
		}
	}
	return account, errors.New("No account for " + currency)
}

// limitOrder sends a limit order based on the amount of funds available in your account or
// otherwise the minimum of 0.001 BTC
func limitOrder(c *coinbasepro.Client) (coinbasepro.Order, error) {
	var order coinbasepro.Order
	productID := os.Getenv("PRODUCT_ID")
	productIDs := strings.Split(productID, "-")
	fiat := productIDs[1]
	book, err := c.GetBook(productID, 1)
	if err != nil {
		return order, err
	}

	lastPrice, err := decimal.NewFromString(book.Bids[0].Price)
	if err != nil {
		return order, err
	}
	buyAmount, err := decimal.NewFromString(os.Getenv("BUY_AMOUNT"))
	if err != nil {
		return order, err
	}
	size := "0.001"
	if buyAmount.Div(lastPrice).GreaterThan(decimal.NewFromFloat(0.001)) {
		size = buyAmount.Div(lastPrice).StringFixed(3)
	}
	sizePrecise, err := decimal.NewFromString(size)
	if err != nil {
		return order, err
	}

	account, err := GetAccount(c, fiat)
	if err != nil {
		return order, err
	}
	available, err := decimal.NewFromString(account.Available)
	if err != nil {
		return order, err
	}
	if sizePrecise.Mul(lastPrice).GreaterThan(available) {
		log.Printf("Not enough funds available. You need %s more.", sizePrecise.Mul(lastPrice).Sub(available).StringFixed(2))
		return order, nil
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
		return order, err
	}
	return order, nil
}

//LoopOrder creates a limit order and if it is not filled within the orderExpiry time
//then a new order will be set at the top bid of the orderbook
func LoopOrder(c *coinbasepro.Client) error {
	order, err := limitOrder(c)
	if err != nil {
		return err
	}
	if order.ID != "" {
		log.Printf("Order %s placed.", order.ID)
	}
	limitOrderTimer = time.NewTimer(orderExpiry)
	<-limitOrderTimer.C
	order, err = c.GetOrder(order.ID)
	if err != nil {
		return err
	}
	if !order.Settled {
		if order.ID != "" {
			err = c.CancelOrder(order.ID)
			if err != nil {
				return err
			}
			log.Printf("Order %s expired.", order.ID)
		}
		return LoopOrder(c)
	}
	if order.ID != "" {
		log.Printf("Order %s settled.", order.ID)
	}
	return nil
}

// Withdraw withdraws all your selected crypto into the $COLD_WALLET
func Withdraw(c *coinbasepro.Client) error {
	productID := os.Getenv("PRODUCT_ID")
	productIDs := strings.Split(productID, "-")
	coin := productIDs[0]
	account, err := GetAccount(c, coin)
	if err != nil {
		return err
	}
	withdrawOpts := coinbasepro.WithdrawalCrypto{
		Currency: coin,
		Amount:   account.Available,
		// Amount:        "0.001",
		CryptoAddress: os.Getenv("COLD_WALLET"),
	}
	_, err = c.CreateWithdrawalCrypto(&withdrawOpts)
	if err != nil {
		log.Println("Error with withdrawal.")
		log.Println(err.Error())
	}
	log.Printf("Withdrew %s %s.", account.Available, coin)
	return nil
}
