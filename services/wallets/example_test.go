package wallets_test

import (
	"context"
	"fmt"
	"log"

	"github.com/jeremyjsx/wallbit-go/services/wallets"
	"github.com/jeremyjsx/wallbit-go/wallbit"
)

func ExampleService_Get() {
	client, err := wallbit.NewClient("YOUR_API_KEY")
	if err != nil {
		log.Fatal(err)
	}

	var svc *wallets.Service = client.Wallets
	res, err := svc.Get(context.Background(), &wallets.GetRequest{
		Currency: "USDT",
	})
	if err != nil {
		log.Fatal(err)
	}

	for _, w := range res.Payload.Data {
		fmt.Printf("[%s/%s] %s\n", w.CurrencyCode, w.Network, w.Address)
	}
}
