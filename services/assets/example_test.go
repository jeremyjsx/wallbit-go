package assets_test

import (
	"context"
	"fmt"
	"log"

	"github.com/jeremyjsx/wallbit-go/services/assets"
	"github.com/jeremyjsx/wallbit-go/wallbit"
)

func ExampleService_Get() {
	client, err := wallbit.NewClient("YOUR_API_KEY")
	if err != nil {
		log.Fatal(err)
	}

	var svc *assets.Service = client.Assets
	res, err := svc.Get(context.Background(), "AAPL")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("%s — %s @ %.2f USD\n", res.Payload.Data.Symbol, res.Payload.Data.Name, res.Payload.Data.Price)
}

func ExampleService_List() {
	client, err := wallbit.NewClient("YOUR_API_KEY")
	if err != nil {
		log.Fatal(err)
	}

	var svc *assets.Service = client.Assets
	res, err := svc.List(context.Background(), &assets.ListRequest{
		Search: "tech",
		Page:   wallbit.Ptr(1),
		Limit:  wallbit.Ptr(20),
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("found %d assets across %d pages\n", res.Payload.Count, res.Payload.Pages)
}
