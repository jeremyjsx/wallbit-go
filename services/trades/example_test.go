package trades_test

import (
	"context"
	"fmt"
	"log"

	"github.com/jeremyjsx/wallbit-go/services/trades"
	"github.com/jeremyjsx/wallbit-go/wallbit"
)

func ExampleService_Create() {
	client, err := wallbit.NewClient("YOUR_API_KEY")
	if err != nil {
		log.Fatal(err)
	}

	var svc *trades.Service = client.Trades
	res, err := svc.Create(context.Background(), trades.CreateRequest{
		Symbol:    "AAPL",
		Direction: "BUY",
		Currency:  "USD",
		OrderType: "MARKET",
		Amount:    wallbit.Ptr(100.0),
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("trade %s status=%s\n", res.Payload.Data.Symbol, res.Payload.Data.Status)
}
