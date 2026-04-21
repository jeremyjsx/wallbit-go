// Package rates exposes the Wallbit exchange-rate endpoint. [Service.Get]
// returns the current conversion rate between a source and destination
// currency pair; identity pairs (e.g. USD→USD) resolve to rate 1.0 with a
// nil UpdatedAt.
//
// See https://developer.wallbit.io/docs/api-reference/rates/get.
package rates
