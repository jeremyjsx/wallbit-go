// Package cards exposes the Wallbit cards endpoints: listing the
// authenticated user's active and suspended cards, and toggling a card's
// status between ACTIVE and SUSPENDED. [Service.Block] and [Service.Unblock]
// are thin wrappers around the underlying status update.
//
// See https://developer.wallbit.io/docs/api-reference/cards/list and
// https://developer.wallbit.io/docs/api-reference/cards/update-status.
package cards
