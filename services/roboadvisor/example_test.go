package roboadvisor_test

import (
	"context"
	"fmt"
	"log"

	"github.com/jeremyjsx/wallbit-go/services/roboadvisor"
	"github.com/jeremyjsx/wallbit-go/wallbit"
)

func ExampleService_GetBalance() {
	client, err := wallbit.NewClient("YOUR_API_KEY")
	if err != nil {
		log.Fatal(err)
	}

	var svc *roboadvisor.Service = client.RoboAdvisor
	res, err := svc.GetBalance(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	for _, p := range res.Payload.Data {
		fmt.Printf("portfolio %d: %.2f USD\n", p.ID, p.PortfolioValue)
	}
}

func ExampleService_Deposit() {
	client, err := wallbit.NewClient("YOUR_API_KEY")
	if err != nil {
		log.Fatal(err)
	}

	var svc *roboadvisor.Service = client.RoboAdvisor
	res, err := svc.Deposit(context.Background(), roboadvisor.DepositRequest{
		RoboAdvisorID: 1,
		Amount:        500,
		From:          roboadvisor.AccountTypeDefault,
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("deposit %s status=%s\n", res.Payload.Data.UUID, res.Payload.Data.Status)
}
