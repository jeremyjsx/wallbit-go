package transactions_test

import (
	"context"
	"fmt"
	"log"

	"github.com/jeremyjsx/wallbit-go/services/transactions"
	"github.com/jeremyjsx/wallbit-go/wallbit"
)

func ExampleService_List() {
	client, err := wallbit.NewClient("YOUR_API_KEY")
	if err != nil {
		log.Fatal(err)
	}

	var svc *transactions.Service = client.Transactions
	res, err := svc.List(context.Background(), &transactions.ListRequest{
		Status: "COMPLETED",
		Limit:  wallbit.Ptr(50),
	})
	if err != nil {
		log.Fatal(err)
	}

	for _, tx := range res.Payload.Data.Data {
		fmt.Printf("%s %s %.2f %s\n", tx.UUID, tx.Type, tx.SourceAmount, tx.SourceCurrency.Code)
	}
}
