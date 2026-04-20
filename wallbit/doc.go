// Package wallbit is a Go SDK for the Wallbit public API
// (https://developer.wallbit.io). It provides a single [Client] composed of
// per-resource service handles (Balance, Transactions, Trades, Fees, Wallets,
// Assets, Operations, RoboAdvisor, Cards, AccountDetails, APIKey, Rates)
// backed by a configurable HTTP transport.
//
// # Authentication
//
// All requests are authenticated using an API key sent in the X-API-Key
// header. Obtain one from the Wallbit dashboard under Agents → Create Agent.
//
//	client, err := wallbit.NewClient(os.Getenv("WALLBIT_API_KEY"))
//
// # Responses
//
// Every service method returns a [*transport.Response] generic wrapper
// pairing the decoded payload (Payload) with the HTTP envelope
// (StatusCode, Header, RequestID). RequestID is the X-Request-ID header
// from the server and is useful when reporting issues to Wallbit support.
//
//	res, err := client.Balance.GetChecking(ctx)
//	if err != nil {
//	    return err
//	}
//	fmt.Println(res.StatusCode, res.RequestID, res.Payload.Data)
//
// # Configuration
//
// Customize the client with functional options: [WithBaseURL],
// [WithHTTPClient], [WithTimeout], [WithUserAgent], [WithRetryPolicy],
// [WithMaxResponseBytes] and [WithHook]. Options may be passed in any
// order.
//
// # Defaults
//
// [NewClient] applies these defaults; all are overridable via the
// matching option:
//
//   - BaseURL: https://api.wallbit.io (HTTPS-only; see Security below).
//   - HTTP timeout: 30s. Covers tail latency during API incidents
//     without leaking goroutines when the server never replies.
//     Override with [WithTimeout] or supply a pre-configured
//     [*net/http.Client] via [WithHTTPClient].
//   - Retry policy: [DefaultRetryPolicy] — 3 attempts, 250ms base
//     delay, 2s cap, exponential with equal-jitter. Worst-case added
//     wait for a failing call is ~750ms of backoff on top of the
//     server's own response times. See the Retries section below for
//     which requests are eligible.
//   - Max response body: [DefaultMaxResponseBytes] (10 MiB). Guards
//     against runaway or hostile responses; overflow returns
//     [ErrResponseTooLarge] instead of a truncated payload. Raise it
//     with [WithMaxResponseBytes] if you consume deliberately large
//     list endpoints.
//   - User-Agent: wallbit-go-sdk/<version>. Version resolves at build
//     time via -ldflags, then via [runtime/debug.ReadBuildInfo], with
//     "dev" as the final fallback. Override with [WithUserAgent].
//
// # Errors
//
// API errors are returned as [*Error], carrying the HTTP status, API error
// code, message, validation details and optional Retry-After hint. Use
// [errors.As] together with the predicates [IsNotFound], [IsAuthError],
// [IsRateLimit], [IsValidationError], [IsServerError] and [IsRetryable] to
// branch on error categories.
//
// # Retries
//
// The client retries idempotent requests (GET, HEAD, DELETE, OPTIONS, TRACE)
// on transport failures and on retryable API responses (HTTP 429 and 5xx).
// POST, PATCH and PUT are never retried automatically. Backoff is exponential
// and honors Retry-After.
//
// # Observability
//
// Register a [Hook] via [WithHook] to observe every HTTP attempt (including
// retries). For standard structured logging, use [SlogHook] which adapts a
// [*log/slog.Logger] to the [Hook] interface and emits one record per
// attempt with method, path, attempt, status and duration_ms. Filter volume
// by configuring the logger's level; request.start is emitted at Debug,
// request.done at Info/Warn/Error depending on status.
//
//	client, _ := wallbit.NewClient(key,
//	    wallbit.WithHook(wallbit.SlogHook(slog.Default())),
//	)
//
// # Security
//
// HTTPS is required by default; non-HTTPS base URLs are rejected unless
// [WithInsecureHTTPForTesting] is set. The provided HTTP client is cloned
// and a CheckRedirect hook is installed to block cross-host redirects so
// that the API key is never leaked to a foreign host.
package wallbit
