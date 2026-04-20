// Package roboadvisor exposes the Wallbit Robo Advisor and Chests endpoints:
// reading portfolio balances, allocations and performance, and depositing
// or withdrawing funds to/from a portfolio from either the checking or the
// investment account.
//
// Deposits have a 10 USD minimum, are processed asynchronously, and require
// completed investment KYC. Only one pending withdrawal per user is allowed
// at a time.
//
// See https://developer.wallbit.io/docs/api-reference/roboadvisor/balance,
// https://developer.wallbit.io/docs/api-reference/roboadvisor/deposit and
// https://developer.wallbit.io/docs/api-reference/roboadvisor/withdraw.
package roboadvisor
