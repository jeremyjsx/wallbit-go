// Package wallbit is a Go SDK for the Wallbit public API
// (https://developer.wallbit.io). It provides a single [Client] composed of
// per-resource service handles (Balance, Transactions, Trades, Fees, Wallets,
// Assets, Operations, RoboAdvisor, Cards, AccountDetails, APIKey) backed by
// a configurable HTTP transport.
//
// # Authentication
//
// All requests are authenticated using an API key sent in the X-API-Key
// header. Obtain one from the Wallbit dashboard under Settings → API Keys.
//
//	client, err := wallbit.NewClient(os.Getenv("WALLBIT_API_KEY"))
//
// # Configuration
//
// Customize the client with functional options: [WithBaseURL],
// [WithHTTPClient], [WithTimeout], [WithUserAgent], [WithRetryPolicy] and
// [WithHook]. Options may be passed in any order.
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
// # Security
//
// HTTPS is required by default; non-HTTPS base URLs are rejected unless
// [WithInsecureHTTPForTesting] is set. The provided HTTP client is cloned
// and a CheckRedirect hook is installed to block cross-host redirects so
// that the API key is never leaked to a foreign host.
package wallbit
