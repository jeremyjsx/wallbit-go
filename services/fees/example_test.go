package fees_test

import (
	"context"
	"fmt"
	"log"

	"github.com/jeremyjsx/wallbit-go/services/fees"
	"github.com/jeremyjsx/wallbit-go/wallbit"
)

func ExampleService_Get() {
	client, err := wallbit.NewClient("YOUR_API_KEY")
	if err != nil {
		log.Fatal(err)
	}

	var svc *fees.Service = client.Fees
	res, err := svc.Get(context.Background(), fees.GetRequest{Type: "TRADE"})
	if err != nil {
		log.Fatal(err)
	}

	if res.Payload.Data.Empty {
		fmt.Println("no fee setting configured")
		return
	}
	fee := res.Payload.Data.Row
	fmt.Printf("%s: %s%% + %s USD\n", fee.FeeType, fee.PercentageFee, fee.FixedFeeUSD)
}
