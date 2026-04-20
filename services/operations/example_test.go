package operations_test

import (
	"context"
	"fmt"
	"log"

	"github.com/jeremyjsx/wallbit-go/services/operations"
	"github.com/jeremyjsx/wallbit-go/wallbit"
)

func ExampleService_Internal() {
	client, err := wallbit.NewClient("YOUR_API_KEY")
	if err != nil {
		log.Fatal(err)
	}

	var svc *operations.Service = client.Operations
	res, err := svc.Internal(context.Background(), operations.InternalRequest{
		Currency: "USD",
		From:     operations.AccountDefault,
		To:       operations.AccountInvestment,
		Amount:   100,
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("transfer %s status=%s\n", res.Payload.Data.UUID, res.Payload.Data.Status)
}
