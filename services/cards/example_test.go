package cards_test

import (
	"context"
	"fmt"
	"log"

	"github.com/jeremyjsx/wallbit-go/services/cards"
	"github.com/jeremyjsx/wallbit-go/wallbit"
)

func ExampleService_List() {
	client, err := wallbit.NewClient("YOUR_API_KEY")
	if err != nil {
		log.Fatal(err)
	}

	var svc *cards.Service = client.Cards
	res, err := svc.List(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	for _, c := range res.Payload.Data {
		fmt.Printf("%s **** %s (%s)\n", c.CardNetwork, c.CardLast4, c.Status)
	}
}

func ExampleService_Block() {
	client, err := wallbit.NewClient("YOUR_API_KEY")
	if err != nil {
		log.Fatal(err)
	}

	var svc *cards.Service = client.Cards
	res, err := svc.Block(context.Background(), "card_uuid_here")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("card %s is now %s\n", res.Payload.Data.UUID, res.Payload.Data.Status)
}
