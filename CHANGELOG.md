# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [6.0.0] - 2026-02-21

### Changed

- `Retry` now returns `errors.Join(context.Cause(ctx), lastErr)` when the context is cancelled, instead of returning only the context error. This ensures the last operation error is always available to the caller.
- `Retry` now returns `errors.Join(context.DeadlineExceeded, lastErr)` when `WithMaxElapsedTime` is exceeded, instead of returning only the operation error. This makes timeout behaviour consistent regardless of whether the deadline comes from the context or `WithMaxElapsedTime`. (#181)

See [`ExampleRetry_outcomes`](example_test.go) for how to inspect the different error outcomes.

## [5.0.0] - 2024-12-19

### Added

- RetryAfterError can be returned from an operation to indicate how long to wait before the next retry.

### Changed

- Retry function now accepts additional options for specifying max number of tries and max elapsed time.
- Retry function now accepts a context.Context.
- Operation function signature changed to return result (any type) and error.

### Removed

- RetryNotify* and RetryWithData functions. Only single Retry function remains.
- Optional arguments from ExponentialBackoff constructor.
- Clock and Timer interfaces.

### Fixed

- The original error is returned from Retry if there's a PermanentError. (#144)
- The Retry function respects the wrapped PermanentError. (#140)
