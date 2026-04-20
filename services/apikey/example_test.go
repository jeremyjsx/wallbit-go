package apikey_test

import (
	"context"
	"fmt"
	"log"

	"github.com/jeremyjsx/wallbit-go/services/apikey"
	"github.com/jeremyjsx/wallbit-go/wallbit"
)

func ExampleService_Revoke() {
	client, err := wallbit.NewClient("YOUR_API_KEY")
	if err != nil {
		log.Fatal(err)
	}

	var svc *apikey.Service = client.APIKey
	res, err := svc.Revoke(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(res.Payload.Message)
}
