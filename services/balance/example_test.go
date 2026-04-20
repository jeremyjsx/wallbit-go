package balance_test

import (
	"context"
	"fmt"
	"log"

	"github.com/jeremyjsx/wallbit-go/services/balance"
	"github.com/jeremyjsx/wallbit-go/wallbit"
)

func ExampleService_GetChecking() {
	client, err := wallbit.NewClient("YOUR_API_KEY")
	if err != nil {
		log.Fatal(err)
	}

	var svc *balance.Service = client.Balance
	res, err := svc.GetChecking(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	for _, b := range res.Payload.Data {
		fmt.Printf("%s: %.2f\n", b.Currency, b.Balance)
	}
}

func ExampleService_GetStocks() {
	client, err := wallbit.NewClient("YOUR_API_KEY")
	if err != nil {
		log.Fatal(err)
	}

	var svc *balance.Service = client.Balance
	res, err := svc.GetStocks(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	for _, p := range res.Payload.Data {
		fmt.Printf("%s: %.4f shares\n", p.Symbol, p.Shares)
	}
}
