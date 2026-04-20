# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

Nothing yet — changes land here first and graduate to a versioned
section at release time.

## [0.1.0-beta.1] — 2026-04-20

First public release.

### Added

- `wallbit.Client`: single HTTP client that exposes the Wallbit API as
  per-resource service handles. Build with `wallbit.NewClient(apiKey, opts...)`
  or `wallbit.NewClientFromConfig(apiKey, cfg)` when configuration
  comes from a struct (env loader, DI container, etc.).
- Functional options: `WithBaseURL`, `WithHTTPClient`, `WithTimeout`,
  `WithUserAgent`, `WithRetryPolicy`, `WithMaxResponseBytes`,
  `WithHook`, `WithInsecureHTTPForTesting`.
- Service handles on `*wallbit.Client`: `Balance`, `Transactions`,
  `Trades`, `Fees`, `Wallets`, `Assets`, `Operations`, `RoboAdvisor`,
  `Cards`, `AccountDetails`, `APIKey`, `Rates`.
- `transport.Response[T]` generic wrapper pairing the decoded payload
  (`Payload`) with the HTTP envelope (`StatusCode`, `Header`,
  `RequestID`).
- `transport.Sender` interface for injecting custom HTTP transports
  and `transport.SendJSON[T]` generic helper that every service uses
  internally.
- `wallbit.Error` with typed fields (`StatusCode`, `Code`, `Message`,
  `Details`, `RequestID`, `RetryAfter`) and predicates
  `IsNotFound`, `IsAuthError`, `IsRateLimit`, `IsValidationError`,
  `IsServerError`, `IsRetryable`. `ErrorFromHTTP` is fuzz-tested for
  resilience against malformed upstream bodies.
- Retry loop with equal-jitter exponential backoff, `Retry-After`
  honoring and attempt counting exposed to hooks. Default policy:
  3 attempts, 250ms base delay, 2s cap. Idempotent methods (`GET`,
  `HEAD`, `DELETE`, `OPTIONS`, `TRACE`) are retried on transport
  errors; `429` and `5xx` are retried regardless of method.
- SDK version injected into the `User-Agent` at build time via
  `-ldflags "-X github.com/jeremyjsx/wallbit-go/wallbit.Version=..."`,
  with `runtime/debug.ReadBuildInfo` and `"dev"` fallbacks.
- `wallbit.Hook` interface with `RequestMeta` / `ResponseMeta`
  carrying `Method`, `Path`, `Attempt`, `StatusCode` and `Duration`
  for every HTTP attempt (retries included).
- `wallbit.SlogHook`: adapter that wires the hook to a
  `*log/slog.Logger`, emitting one record per attempt with
  `method`, `path`, `attempt`, `status` and `duration_ms`. Levels:
  Debug on start, Info/Warn/Error on done based on status.
- `wallbit.Ptr[T any](v T) *T` helper for optional fields in request
  bodies so call-sites read `wallbit.Ptr("value")` alongside the
  client.
- Go 1.23 pagination: `transactions.ListAll` and `assets.ListAll`
  return `iter.Seq2[T, error]` and walk every page lazily.
- Timestamp fields (`CreatedAt`, `UpdatedAt`) typed as `time.Time`
  across services that expose them (`trades`, `transactions`,
  `operations`, `roboadvisor`, `rates`). Invalid timestamps surface
  as JSON decode errors instead of silently passing through.
- Nullable API fields consistently typed as pointers (`*string`,
  `*float64`, …) so `nil` is distinguishable from "explicitly zero".
- Error `details` preserved as raw JSON so callers can re-decode the
  server-specific payload without losing information.
- Input validation before issuing the HTTP call for paths that take a
  required segment (asset symbol, card UUID, currency pair), with
  typed sentinel errors (`ErrEmptySymbol`, `ErrEmptyCurrency`, …).
- `Example*` functions in every package so `pkg.go.dev` renders
  runnable usage for each service method.

### Security

- HTTPS is required by default. Non-HTTPS base URLs are rejected with
  `ErrInsecureBaseURL` unless `WithInsecureHTTPForTesting` is set,
  preventing accidental plaintext transmission of the API key.
- Cross-host redirects are blocked by a `CheckRedirect` hook on the
  cloned `*http.Client` so a hostile or misconfigured redirect cannot
  exfiltrate the `X-API-Key` header to a foreign host.
- Response bodies are read through `io.LimitReader` with a default
  cap of 10 MiB (`DefaultMaxResponseBytes`). Over-sized responses
  return `ErrResponseTooLarge` instead of a partial payload so a
  hostile or buggy upstream cannot exhaust process memory. Override
  with `WithMaxResponseBytes`.

[Unreleased]: https://github.com/jeremyjsx/wallbit-go/compare/v0.1.0-beta.1...HEAD
[0.1.0-beta.1]: https://github.com/jeremyjsx/wallbit-go/releases/tag/v0.1.0-beta.1
