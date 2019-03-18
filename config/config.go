package config

//coinbase API v2
var (
	BaseURL     string
	Version     string
	VersionDate string
)

func init() {
	BaseURL = "https://api.coinbase.com/v2/"
	Version = "/v2/"
	VersionDate = "2019-03-17"
}
