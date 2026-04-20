// Package transport defines the [Sender] interface that decouples the
// Wallbit [github.com/jeremyjsx/wallbit-go/wallbit] client from the
// per-resource service packages.
//
// Most consumers do not need to import this package directly. Import it when
// implementing a custom [Sender] for tests, mocking, or wrapping the real
// client with middlewares (logging, metrics, distributed tracing, custom
// retries). All user-facing knobs (retry policy, hooks, request/response
// metadata) live on the wallbit package itself.
package transport
