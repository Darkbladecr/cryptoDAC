package main

import (
	"encoding/json"
	"fmt"
	"time"
)

// Pprint will pretty print any json/interface
func Pprint(v interface{}) {
	if msg, err := json.MarshalIndent(v, "", "  "); err != nil {
		panic(err)
	} else {
		println(string(msg))
	}
}

// GetOrderInterval will get Duration object for the next order
func GetOrderInterval() time.Duration {
	now := time.Now()
	monthly := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	next := monthly.AddDate(0, 1, 0)
	fmt.Printf("Next order: %s", next.String())
	return next.Sub(now)
}
