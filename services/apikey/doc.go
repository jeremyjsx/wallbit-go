// Package apikey exposes the Wallbit API key management endpoint. The only
// supported operation is revoking the API key carried in the X-API-Key
// header; this does not require the read or trade scopes since any valid key
// may revoke itself.
//
// See https://developer.wallbit.io/docs/api-reference/api-key/revoke.
package apikey
