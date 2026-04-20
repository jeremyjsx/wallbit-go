package accountdetails_test

import (
	"context"
	"fmt"
	"log"

	"github.com/jeremyjsx/wallbit-go/services/accountdetails"
	"github.com/jeremyjsx/wallbit-go/wallbit"
)

func ExampleService_Get() {
	client, err := wallbit.NewClient("YOUR_API_KEY")
	if err != nil {
		log.Fatal(err)
	}

	var svc *accountdetails.Service = client.AccountDetails
	res, err := svc.Get(context.Background(), &accountdetails.GetRequest{
		Country:  accountdetails.CountryUS,
		Currency: accountdetails.CurrencyUSD,
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Bank: %s, Holder: %s\n", res.Payload.Data.BankName, res.Payload.Data.HolderName)
}
