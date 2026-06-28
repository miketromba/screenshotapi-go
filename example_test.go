package screenshotapi_test

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"

	screenshotapi "github.com/miketromba/screenshotapi-go"
)

func ExampleClient_Save() {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/png")
		w.Header().Set("x-credits-remaining", "199")
		w.Header().Set("x-screenshot-id", "ss_example")
		w.Header().Set("x-duration-ms", "450")
		_, _ = w.Write([]byte("png"))
	}))
	defer server.Close()

	file, err := os.CreateTemp("", "screenshotapi-*.png")
	if err != nil {
		log.Fatal(err)
	}
	path := file.Name()
	_ = file.Close()
	defer os.Remove(path)

	client := screenshotapi.NewClient(
		"sk_test_abc123",
		screenshotapi.WithBaseURL(server.URL),
	)
	metadata, err := client.Save(context.Background(), screenshotapi.ScreenshotOptions{
		URL: "https://example.com",
	}, path)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Credits remaining: %d\n", metadata.CreditsRemaining)

	// Output:
	// Credits remaining: 199
}

func ExampleClient_Screenshot_advancedOptions() {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/webp")
		w.Header().Set("x-credits-remaining", "198")
		w.Header().Set("x-screenshot-id", "ss_advanced")
		w.Header().Set("x-duration-ms", "875")
		_, _ = w.Write([]byte("webp"))
	}))
	defer server.Close()

	client := screenshotapi.NewClient(
		"sk_test_abc123",
		screenshotapi.WithBaseURL(server.URL),
	)
	result, err := client.Screenshot(context.Background(), screenshotapi.ScreenshotOptions{
		URL:                 "https://example.com",
		Width:               1440,
		Height:              900,
		FullPage:            true,
		Type:                screenshotapi.WebP,
		Quality:             85,
		ColorScheme:         screenshotapi.Dark,
		WaitUntil:           screenshotapi.NetworkIdle2,
		BlockAds:            true,
		RemoveCookieBanners: true,
		DevicePixelRatio:    2,
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("%s %s\n", result.ContentType, result.Metadata.ScreenshotID)

	// Output:
	// image/webp ss_advanced
}

func ExampleClient_Screenshot_errorHandling() {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusPaymentRequired)
		_, _ = w.Write([]byte(`{"error":"Insufficient credits","balance":0}`))
	}))
	defer server.Close()

	client := screenshotapi.NewClient(
		"sk_test_abc123",
		screenshotapi.WithBaseURL(server.URL),
	)
	_, err := client.Screenshot(context.Background(), screenshotapi.ScreenshotOptions{
		URL: "https://example.com",
	})

	var creditsErr *screenshotapi.InsufficientCreditsError
	if errors.As(err, &creditsErr) {
		fmt.Printf("out of credits: %d\n", creditsErr.Balance)
	}

	// Output:
	// out of credits: 0
}
