// Package fees exposes the Wallbit fee configuration endpoint, which returns
// the fee setting row for the requested fee type and the authenticated user's
// current investment subscription tier. When no row matches the user's tier
// and fee type, the API returns an empty array, surfaced as
// [GetData.Empty] = true.
//
// See https://developer.wallbit.io/docs/api-reference/fees/get.
package fees
