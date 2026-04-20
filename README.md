# wallbit-go

[![Go Reference](https://pkg.go.dev/badge/github.com/jeremyjsx/wallbit-go.svg)](https://pkg.go.dev/github.com/jeremyjsx/wallbit-go)
[![Go Report Card](https://goreportcard.com/badge/github.com/jeremyjsx/wallbit-go)](https://goreportcard.com/report/github.com/jeremyjsx/wallbit-go)
[![License: Apache-2.0](https://img.shields.io/badge/License-Apache_2.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)

A Go SDK for the [Wallbit public API](https://developer.wallbit.io). Type-safe,
context-aware, retry-friendly, with first-class error inspection.

## Disclaimer

This is an **unofficial, community-maintained** SDK. It is **not affiliated with, endorsed by, or sponsored by Wallbit**. "Wallbit" is a trademark of its respective owner and is used here solely to describe the API this library targets (nominative use). For official documentation, see <https://developer.wallbit.io>.

## Install

```bash
go get github.com/jeremyjsx/wallbit-go
```

Requires Go 1.23 or newer.

## Quickstart

```go
package main

import (
    "context"
    "fmt"
    "log"
    "os"

    "github.com/jeremyjsx/wallbit-go/wallbit"
)

func main() {
    client, err := wallbit.NewClient(os.Getenv("WALLBIT_API_KEY"))
    if err != nil {
        log.Fatal(err)
    }

    balance, err := client.Balance.GetChecking(context.Background())
    if err != nil {
        log.Fatal(err)
    }

    for _, b := range balance.Data {
        fmt.Printf("%s: %.2f\n", b.Currency, b.Balance)
    }
}
```

## Endpoint coverage

All endpoints documented in the Wallbit OpenAPI spec are covered.

| Service          | Method                                | API endpoint                                        |
| ---------------- | ------------------------------------- | --------------------------------------------------- |
| `Balance`        | `GetChecking`, `GetStocks`            | `GET /balance/{checking,stocks}`                    |
| `Transactions`   | `List`                                | `GET /transactions`                                 |
| `Trades`         | `Create`                              | `POST /trades`                                      |
| `Fees`           | `Get`                                 | `POST /fees`                                        |
| `AccountDetails` | `Get`                                 | `GET /account-details`                              |
| `Wallets`        | `Get`                                 | `GET /wallets`                                      |
| `Assets`         | `List`, `Get`                         | `GET /assets`, `GET /assets/{symbol}`               |
| `Operations`     | `Internal`, `DepositInvestment`, `WithdrawInvestment` | `POST /operations/internal`             |
| `RoboAdvisor`    | `GetBalance`, `Deposit`, `Withdraw`   | `GET /roboadvisor/balance`, `POST /roboadvisor/{deposit,withdraw}` |
| `Cards`          | `List`, `Block`, `Unblock`            | `GET /cards`, `PATCH /cards/{uuid}/status`          |
| `APIKey`         | `Revoke`                              | `DELETE /api-key`                                   |

## Error handling

API errors return a typed `*wallbit.Error` carrying the status code, error code, message, validation details and `RetryAfter` hint. Use `errors.As` to inspect:

```go
import (
    "errors"
    "time"

    "github.com/jeremyjsx/wallbit-go/wallbit"
)

_, err := client.Trades.Create(ctx, req)
var apiErr *wallbit.Error
if errors.As(err, &apiErr) {
    switch {
    case wallbit.IsValidationError(err):
        // 400 / 422 — inspect apiErr.Errors / apiErr.Details
    case wallbit.IsRateLimit(err):
        time.Sleep(apiErr.RetryAfter())
    case wallbit.IsAuthError(err):
        // 401 / 403 — refresh credentials or surface to the user
    }
}
```

Helper predicates: `IsNotFound`, `IsAuthError`, `IsRateLimit`,
`IsValidationError`, `IsServerError`, `IsRetryable`.

## Retries

The client retries idempotent requests (`GET`, `HEAD`, `DELETE`, `OPTIONS`, `TRACE`) on transport errors and on retryable API responses (`429` and `5xx`).
`POST`, `PATCH` and `PUT` are never retried automatically.

```go
client, _ := wallbit.NewClient(apiKey,
    wallbit.WithRetryPolicy(wallbit.RetryPolicy{
        MaxAttempts: 4,
        BaseDelay:   500 * time.Millisecond,
        MaxDelay:    5 * time.Second,
    }),
)
```

The client honors `Retry-After` headers and the API's `retry_after` field.
Backoff is exponential and capped by `MaxDelay`.

## Hooks (observability)

Plug in metrics or logging by implementing the `Hook` interface:

```go
type metricsHook struct{}

func (metricsHook) OnRequestStart(m *wallbit.RequestMeta) { /* ... */ }
func (metricsHook) OnRequestDone(m *wallbit.ResponseMeta) { /* ... */ }

client, _ := wallbit.NewClient(apiKey, wallbit.WithHook(metricsHook{}))
```

Hooks are called on every attempt (including retries) and must be safe for concurrent use.

## Custom HTTP client

```go
client, _ := wallbit.NewClient(apiKey,
    wallbit.WithHTTPClient(&http.Client{
        Timeout:   45 * time.Second,
        Transport: myTracedTransport,
    }),
)
```

The SDK clones the provided `http.Client` and installs a `CheckRedirect` that blocks cross-host redirects, so your `X-API-Key` is never leaked to a foreign host even if the API ever returns a malicious redirect.

## Security

- HTTPS is enforced for the base URL. Override with `wallbit.WithInsecureHTTPForTesting()` for local servers / `httptest`.
- Cross-host redirects are blocked by default (see above).
- Never commit your API key. Read it from an environment variable or secret manager.

## License

Licensed under the [Apache License, Version 2.0](./LICENSE). See [NOTICE](./NOTICE)
for attribution and trademark information.
