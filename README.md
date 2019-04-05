# cryptoDAC

Small server worker to take money from your Coinbase Pro account and place orders on the first day of the month. You have a number of environmental variables that you can tweak for the pricing and crypto pair.

```Shell
COINBASE_PRO_PASSPHRASE="passphrase"
COINBASE_PRO_KEY="key"
COINBASE_PRO_SECRET="secret"
BUY_AMOUNT=100
PRODUCT_ID="BTC-USD"
COLD_WALLET="wallet address"
```

You can test the package with the coinbase sandbox by activating the environmental variable:

```Shell
COINBASE_PRO_SANDBOX=1
```

For a quick test of this package I would recommend trying a [free dyno from Heroku.](https://www.heroku.com/go)

Otherwise you can clone this package and run it with:

```Shell
go run .
```
