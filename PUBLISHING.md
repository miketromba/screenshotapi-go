# Go SDK Publishing Checklist

This SDK is prepared for a standalone public repository:

```text
module github.com/miketromba/screenshotapi-go
```

## Current Readiness

- Module path resolves to the public standalone repository.
- Package name is `screenshotapi`.
- License is MIT in `LICENSE`.
- README covers install, authentication, first screenshot, advanced options, error handling, free tier/pricing, docs, and support.
- Go examples are in `example_test.go` so pkg.go.dev can render verified examples.
- Tests use `httptest` and temporary files only; no live ScreenshotAPI calls are made.
- The SDK has no third-party dependencies.

## Pre-Release Checklist

Run these from the repository root:

```bash
gofmt -w *.go
go test ./...
go vet ./...
```

Then verify:

- `README.md` import path matches `go.mod`.
- All ScreenshotAPI links in `README.md` include `utm_source=github`, `utm_medium=go_sdk`, `utm_campaign=sdk_distribution`, and `ref=go-sdk`.
- `LICENSE` is present and still MIT.
- `example_test.go` examples pass as part of `go test`.
- No live API key or generated screenshot output is committed.

## Version Tags

Use semantic versioning with unprefixed tags from the standalone public repo:

```bash
git tag -a v1.0.0 -m "Release Go SDK v1.0.0"
git push origin v1.0.0
```

Consumers install the module version with:

```bash
go get github.com/miketromba/screenshotapi-go@v1.0.0
```

For a future `v2+` release, update the module path to include the major version suffix:

```text
module github.com/miketromba/screenshotapi-go/v2
```

## pkg.go.dev Refresh

After pushing a release tag, request it through the module proxy and then open pkg.go.dev:

```bash
GOPROXY=proxy.golang.org go list -m github.com/miketromba/screenshotapi-go@v1.0.0
```

Package page:

```text
https://pkg.go.dev/github.com/miketromba/screenshotapi-go
```
