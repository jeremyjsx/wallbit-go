// Package trades exposes the Wallbit trades endpoint, which executes BUY or
// SELL operations on market assets. Both MARKET and LIMIT orders are
// supported; orders may be sized by USD amount or by share count, but never
// both at the same time.
//
// See https://developer.wallbit.io/docs/api-reference/trades/create.
package trades
