package rates_test

import (
	"context"
	"fmt"
	"log"

	"github.com/jeremyjsx/wallbit-go/services/rates"
	"github.com/jeremyjsx/wallbit-go/wallbit"
)

func ExampleService_Get() {
	client, err := wallbit.NewClient("YOUR_API_KEY")
	if err != nil {
		log.Fatal(err)
	}

	var svc *rates.Service = client.Rates
	res, err := svc.Get(context.Background(), rates.GetRequest{
		SourceCurrency: "ARS",
		DestCurrency:   "USD",
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("%s: 1 %s = %.4f %s\n",
		res.Payload.Data.Pair,
		res.Payload.Data.SourceCurrency,
		res.Payload.Data.Rate,
		res.Payload.Data.DestCurrency,
	)
}
