package wallbit_test

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/jeremyjsx/wallbit-go/wallbit"
)

// Example shows the typical lifecycle of a Wallbit client: build it once
// at startup with the desired options, then reuse it across goroutines.
// The client is safe for concurrent use.
func Example() {
	client, err := wallbit.NewClient(
		"YOUR_API_KEY",
		wallbit.WithTimeout(15*time.Second),
		wallbit.WithRetryPolicy(wallbit.RetryPolicy{
			MaxAttempts: 3,
			BaseDelay:   250 * time.Millisecond,
			MaxDelay:    2 * time.Second,
		}),
	)
	if err != nil {
		log.Fatal(err)
	}

	res, err := client.Balance.GetChecking(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	for _, b := range res.Payload.Data {
		fmt.Printf("%s: %.2f\n", b.Currency, b.Balance)
	}
}

// ExampleNewClientFromConfig shows building a client from a single
// [Config] block instead of multiple [Option] calls. Use this when the
// configuration comes from a struct already populated elsewhere (env
// loader, config file, dependency injection container).
func ExampleNewClientFromConfig() {
	cfg := &wallbit.Config{
		UserAgent: "my-app/1.0",
		RetryPolicy: wallbit.RetryPolicy{
			MaxAttempts: 5,
			BaseDelay:   500 * time.Millisecond,
			MaxDelay:    5 * time.Second,
		},
	}

	client, err := wallbit.NewClientFromConfig("YOUR_API_KEY", cfg)
	if err != nil {
		log.Fatal(err)
	}

	_ = client
}
