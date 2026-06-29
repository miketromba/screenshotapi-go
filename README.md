# ScreenshotAPI Go SDK

Official Go SDK for [ScreenshotAPI][home]. Capture website screenshots, PDFs, and HTML renders from Go with the standard library only.

[![Go Reference](https://pkg.go.dev/badge/github.com/miketromba/screenshotapi-go.svg)](https://pkg.go.dev/github.com/miketromba/screenshotapi-go)

## Installation

```bash
go get github.com/miketromba/screenshotapi-go
```

```go
import screenshotapi "github.com/miketromba/screenshotapi-go"
```

Requires Go 1.21 or later.

## Authentication

Create an API key from the [ScreenshotAPI dashboard][api-keys], then keep it in an environment variable:

```bash
export SCREENSHOTAPI_KEY="sk_live_your_key_here"
```

The SDK sends the key in the `x-api-key` header. Never commit live keys to source control.

## First Screenshot

```go
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	screenshotapi "github.com/miketromba/screenshotapi-go"
)

func main() {
	apiKey := os.Getenv("SCREENSHOTAPI_KEY")
	if apiKey == "" {
		log.Fatal("SCREENSHOTAPI_KEY is required")
	}

	client := screenshotapi.NewClient(apiKey)
	metadata, err := client.Save(context.Background(), screenshotapi.ScreenshotOptions{
		URL: "https://example.com",
	}, "screenshot.png")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Saved screenshot.png, credits remaining: %d\n", metadata.CreditsRemaining)
}
```

## Advanced Options

Use `Screenshot` when you want the image bytes in memory:

```go
result, err := client.Screenshot(ctx, screenshotapi.ScreenshotOptions{
	URL:                 "https://example.com",
	Width:               1440,
	Height:              900,
	FullPage:            true,
	Type:                screenshotapi.WebP,
	Quality:             85,
	ColorScheme:         screenshotapi.Dark,
	WaitUntil:           screenshotapi.NetworkIdle2,
	WaitForSelector:     "#app-ready",
	Delay:               250,
	BlockAds:            true,
	RemoveCookieBanners: true,
	CSSInject:           ".newsletter-modal { display: none !important; }",
	JSInject:            "document.body.dataset.capture = 'true'",
	StealthMode:         true,
	DevicePixelRatio:    2,
	Timezone:            "America/New_York",
	Locale:              "en-US",
	CacheTTL:            300,
	PreloadFonts:        true,
	RemoveElements:      []string{".popup", "#promo-banner"},
	RemovePopups:        true,
	Geolocation: &screenshotapi.Geolocation{
		Latitude:  37.7749,
		Longitude: -122.4194,
		Accuracy:  25,
	},
})
if err != nil {
	return err
}

fmt.Println(result.ContentType)
fmt.Println(result.Metadata.ScreenshotID)
```

Render raw HTML with a POST request by setting `HTML`:

```go
result, err := client.Screenshot(ctx, screenshotapi.ScreenshotOptions{
	HTML: "<main><h1>Hello from Go</h1></main>",
	Type: screenshotapi.PDF,
})
```

Device mockups are available with `MockupDevice: screenshotapi.BrowserMockup`, `IPhoneMockup`, or `MacBookMockup`. Do not combine mockups with `FullPage` or `Type: screenshotapi.PDF`.

See the [Go SDK docs][docs] and [Screenshot API reference][api-reference] for the full parameter reference.

## Error Handling

The SDK returns typed errors for common API failures:

```go
import "errors"

result, err := client.Screenshot(ctx, screenshotapi.ScreenshotOptions{
	URL: "https://example.com",
})
if err != nil {
	var authErr *screenshotapi.AuthenticationError
	var creditsErr *screenshotapi.InsufficientCreditsError
	var keyErr *screenshotapi.InvalidAPIKeyError
	var failedErr *screenshotapi.ScreenshotFailedError

	switch {
	case errors.As(err, &authErr):
		// 401: API key missing or malformed.
	case errors.As(err, &creditsErr):
		fmt.Printf("Out of credits, balance: %d\n", creditsErr.Balance)
	case errors.As(err, &keyErr):
		// 403: API key revoked or invalid.
	case errors.As(err, &failedErr):
		// 500: Screenshot capture failed server-side.
	default:
		// Network errors, validation errors, or other API responses.
	}
}

_ = result
```

When contacting support, include `result.Metadata.ScreenshotID` if a request reached the API.

## Client Configuration

```go
client := screenshotapi.NewClient(apiKey,
	screenshotapi.WithTimeout(30*time.Second),
	screenshotapi.WithBaseURL("https://screenshotapi.to"),
	screenshotapi.WithHTTPClient(&http.Client{Timeout: 30 * time.Second}),
)
```

`WithBaseURL` is mainly for proxies, private gateways, and tests. The default is `https://screenshotapi.to`.

## Pricing And Free Tier

New accounts include **200 free screenshots per month** with no credit card required. [Create a free account][signup] or review [pricing and credit packs][pricing] when you are ready to scale.

## Support

- Documentation: [Go SDK docs][docs]
- API reference: [Screenshot endpoint][api-reference]
- Support: [support@screenshotapi.to][support]

## Development

```bash
go test ./...
go vet ./...
```

Tests use `httptest` and do not call the live ScreenshotAPI service.

## License

MIT

[home]: https://screenshotapi.to?utm_source=github&utm_medium=go_sdk&utm_campaign=sdk_distribution&ref=go-sdk
[signup]: https://screenshotapi.to/sign-up?utm_source=github&utm_medium=go_sdk&utm_campaign=sdk_distribution&ref=go-sdk
[pricing]: https://screenshotapi.to/pricing?utm_source=github&utm_medium=go_sdk&utm_campaign=sdk_distribution&ref=go-sdk
[docs]: https://screenshotapi.to/docs/sdks/go?utm_source=github&utm_medium=go_sdk&utm_campaign=sdk_distribution&ref=go-sdk
[api-reference]: https://screenshotapi.to/docs/api/screenshot?utm_source=github&utm_medium=go_sdk&utm_campaign=sdk_distribution&ref=go-sdk
[api-keys]: https://screenshotapi.to/dashboard/api-keys?utm_source=github&utm_medium=go_sdk&utm_campaign=sdk_distribution&ref=go-sdk
[support]: mailto:support@screenshotapi.to?subject=ScreenshotAPI%20Go%20SDK%20support
