// Package accountdetails exposes the Wallbit account details endpoint, which
// returns bank account information used for deposits. The returned fields
// vary by country and currency: ACH details (US), SEPA details (EU), or
// local bank account details. Currently only US and EU are supported in the
// public API.
//
// See https://developer.wallbit.io/docs/api-reference/account-details/get.
package accountdetails
