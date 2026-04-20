// Package operations exposes the Wallbit internal operations endpoint, which
// moves funds between the checking (DEFAULT) account and the investment
// (INVESTMENT) account. [Service.DepositInvestment] and
// [Service.WithdrawInvestment] are convenience wrappers around the generic
// [Service.Internal] for the most common direction.
//
// Investment KYC must be completed for these operations to succeed.
//
// See https://developer.wallbit.io/docs/api-reference/operations/internal.
package operations
