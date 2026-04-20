// Package transport defines the [Sender] interface that decouples the
// Wallbit [github.com/jeremyjsx/wallbit-go/wallbit] client from the
// per-resource service packages, together with the generic [Response]
// wrapper that every service method returns.
//
// Most consumers import this package for one of two reasons:
//
//   - To reference the return type of a service method explicitly, since
//     every method returns a [*Response] of its own payload type (for
//     example [*transport.Response[balance.CheckingBalanceResponse]]).
//     In most call sites type inference makes this import unnecessary.
//   - To implement a custom [Sender] for tests, mocking, or wrapping the
//     real client with middlewares (logging, metrics, distributed
//     tracing, custom retries).
//
// All user-facing knobs (retry policy, hooks, base URL, timeouts) live on
// the wallbit package itself.
package transport
