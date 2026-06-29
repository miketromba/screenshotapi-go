package screenshotapi

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const testAPIKey = "sk_test_abc123"

func TestScreenshotBuildsGETRequestWithAllOptions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("method = %s, want %s", r.Method, http.MethodGet)
		}
		if r.URL.Path != "/api/v1/screenshot" {
			t.Fatalf("path = %s, want /api/v1/screenshot", r.URL.Path)
		}
		if got := r.Header.Get("x-api-key"); got != testAPIKey {
			t.Fatalf("x-api-key = %q, want %q", got, testAPIKey)
		}

		want := map[string]string{
			"url":                 "https://example.com",
			"width":               "1280",
			"height":              "720",
			"fullPage":            "true",
			"type":                "jpeg",
			"quality":             "80",
			"colorScheme":         "dark",
			"waitUntil":           "networkidle0",
			"waitForSelector":     "#main",
			"delay":               "500",
			"blockAds":            "true",
			"removeCookieBanners": "true",
			"cssInject":           "body { background: black; }",
			"jsInject":            "document.body.dataset.ready = 'true'",
			"stealthMode":         "true",
			"devicePixelRatio":    "2",
			"timezone":            "America/New_York",
			"locale":              "en-US",
			"cacheTtl":            "300",
			"preloadFonts":        "true",
			"removeElements":      ".cookie,#promo",
			"removePopups":        "true",
			"mockupDevice":        "browser",
			"geoLatitude":         "37.7749",
			"geoLongitude":        "-122.4194",
			"geoAccuracy":         "25",
		}
		query := r.URL.Query()
		for key, value := range want {
			if got := query.Get(key); got != value {
				t.Fatalf("query[%s] = %q, want %q", key, got, value)
			}
		}
		if query.Has("html") {
			t.Fatal("GET query unexpectedly included html")
		}

		w.Header().Set("Content-Type", "image/jpeg")
		w.Header().Set("x-credits-remaining", "950")
		w.Header().Set("x-screenshot-id", "ss_get")
		w.Header().Set("x-duration-ms", "1234")
		_, _ = w.Write([]byte("image-bytes"))
	}))
	defer server.Close()

	client := NewClient(testAPIKey, WithBaseURL(server.URL+"///"))
	result, err := client.Screenshot(context.Background(), ScreenshotOptions{
		URL:                 "https://example.com",
		Width:               1280,
		Height:              720,
		FullPage:            true,
		Type:                JPEG,
		Quality:             80,
		ColorScheme:         Dark,
		WaitUntil:           NetworkIdle0,
		WaitForSelector:     "#main",
		Delay:               500,
		BlockAds:            true,
		RemoveCookieBanners: true,
		CSSInject:           "body { background: black; }",
		JSInject:            "document.body.dataset.ready = 'true'",
		StealthMode:         true,
		DevicePixelRatio:    2,
		Timezone:            "America/New_York",
		Locale:              "en-US",
		CacheTTL:            300,
		PreloadFonts:        true,
		RemoveElements:      []string{".cookie", "#promo"},
		RemovePopups:        true,
		MockupDevice:        BrowserMockup,
		Geolocation: &Geolocation{
			Latitude:  37.7749,
			Longitude: -122.4194,
			Accuracy:  25,
		},
	})
	if err != nil {
		t.Fatalf("Screenshot returned error: %v", err)
	}

	if string(result.Image) != "image-bytes" {
		t.Fatalf("image = %q, want image-bytes", string(result.Image))
	}
	if result.ContentType != "image/jpeg" {
		t.Fatalf("content type = %q, want image/jpeg", result.ContentType)
	}
	if result.Metadata.CreditsRemaining != 950 {
		t.Fatalf("credits = %d, want 950", result.Metadata.CreditsRemaining)
	}
	if result.Metadata.ScreenshotID != "ss_get" {
		t.Fatalf("screenshot id = %q, want ss_get", result.Metadata.ScreenshotID)
	}
	if result.Metadata.DurationMs != 1234 {
		t.Fatalf("duration = %d, want 1234", result.Metadata.DurationMs)
	}
}

func TestScreenshotOmitsUnsetGETOptions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()
		if len(query) != 1 {
			t.Fatalf("query = %v, want only url", query)
		}
		if got := query.Get("url"); got != "https://example.com" {
			t.Fatalf("url = %q, want https://example.com", got)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient(testAPIKey, WithBaseURL(server.URL))
	result, err := client.Screenshot(context.Background(), ScreenshotOptions{
		URL: "https://example.com",
	})
	if err != nil {
		t.Fatalf("Screenshot returned error: %v", err)
	}
	if result.ContentType != "image/png" {
		t.Fatalf("content type = %q, want image/png", result.ContentType)
	}
}

func TestScreenshotBuildsPOSTRequestForHTML(t *testing.T) {
	type postBody struct {
		URL                 string       `json:"url"`
		HTML                string       `json:"html"`
		Width               int          `json:"width"`
		Type                ImageType    `json:"type"`
		BlockAds            bool         `json:"blockAds"`
		RemoveCookieBanners bool         `json:"removeCookieBanners"`
		CSSInject           string       `json:"cssInject"`
		JSInject            string       `json:"jsInject"`
		DevicePixelRatio    int          `json:"devicePixelRatio"`
		RemoveElements      []string     `json:"removeElements"`
		Geolocation         *Geolocation `json:"geoLocation"`
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("method = %s, want %s", r.Method, http.MethodPost)
		}
		if got := r.Header.Get("Content-Type"); !strings.HasPrefix(got, "application/json") {
			t.Fatalf("content type = %q, want application/json", got)
		}
		if got := r.Header.Get("x-api-key"); got != testAPIKey {
			t.Fatalf("x-api-key = %q, want %q", got, testAPIKey)
		}

		var body postBody
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("failed to decode body: %v", err)
		}
		if body.URL != "https://example.com/base" {
			t.Fatalf("body url = %q, want https://example.com/base", body.URL)
		}
		if body.HTML != "<h1>Hello</h1>" {
			t.Fatalf("body html = %q, want <h1>Hello</h1>", body.HTML)
		}
		if body.Width != 1200 {
			t.Fatalf("body width = %d, want 1200", body.Width)
		}
		if body.Type != PDF {
			t.Fatalf("body type = %q, want %q", body.Type, PDF)
		}
		if !body.BlockAds || !body.RemoveCookieBanners {
			t.Fatalf("boolean options not encoded: %+v", body)
		}
		if body.CSSInject != "body{color:red}" || body.JSInject != "document.title='x'" {
			t.Fatalf("inject options not encoded: %+v", body)
		}
		if body.DevicePixelRatio != 3 {
			t.Fatalf("devicePixelRatio = %d, want 3", body.DevicePixelRatio)
		}
		if len(body.RemoveElements) != 2 || body.RemoveElements[0] != ".modal" || body.RemoveElements[1] != "#banner" {
			t.Fatalf("removeElements = %#v, want .modal/#banner", body.RemoveElements)
		}
		if body.Geolocation == nil || body.Geolocation.Latitude != 0 || body.Geolocation.Longitude != 0 {
			t.Fatalf("geoLocation = %#v, want zero coordinates", body.Geolocation)
		}

		w.Header().Set("Content-Type", "application/pdf")
		w.Header().Set("x-credits-remaining", "199")
		w.Header().Set("x-screenshot-id", "ss_post")
		w.Header().Set("x-duration-ms", "987")
		_, _ = w.Write([]byte("%PDF"))
	}))
	defer server.Close()

	client := NewClient(testAPIKey, WithBaseURL(server.URL))
	result, err := client.Screenshot(context.Background(), ScreenshotOptions{
		URL:                 "https://example.com/base",
		HTML:                "<h1>Hello</h1>",
		Width:               1200,
		Type:                PDF,
		BlockAds:            true,
		RemoveCookieBanners: true,
		CSSInject:           "body{color:red}",
		JSInject:            "document.title='x'",
		DevicePixelRatio:    3,
		RemoveElements:      []string{".modal", "#banner"},
		Geolocation: &Geolocation{
			Latitude:  0,
			Longitude: 0,
		},
	})
	if err != nil {
		t.Fatalf("Screenshot returned error: %v", err)
	}
	if result.ContentType != "application/pdf" {
		t.Fatalf("content type = %q, want application/pdf", result.ContentType)
	}
	if result.Metadata.ScreenshotID != "ss_post" {
		t.Fatalf("screenshot id = %q, want ss_post", result.Metadata.ScreenshotID)
	}
}

func TestScreenshotRequiresURLOrHTML(t *testing.T) {
	client := NewClient(testAPIKey)
	_, err := client.Screenshot(context.Background(), ScreenshotOptions{})
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "URL or HTML is required") {
		t.Fatalf("error = %q, want URL or HTML is required", err.Error())
	}
}

func TestSaveWritesFileAndReturnsMetadata(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/webp")
		w.Header().Set("x-credits-remaining", "42")
		w.Header().Set("x-screenshot-id", "ss_save")
		w.Header().Set("x-duration-ms", "321")
		_, _ = w.Write([]byte("saved-image"))
	}))
	defer server.Close()

	client := NewClient(testAPIKey, WithBaseURL(server.URL))
	path := filepath.Join(t.TempDir(), "screenshot.webp")
	metadata, err := client.Save(context.Background(), ScreenshotOptions{
		URL:  "https://example.com",
		Type: WebP,
	}, path)
	if err != nil {
		t.Fatalf("Save returned error: %v", err)
	}
	if metadata.CreditsRemaining != 42 || metadata.ScreenshotID != "ss_save" || metadata.DurationMs != 321 {
		t.Fatalf("metadata = %+v, want credits/id/duration", metadata)
	}
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read saved file: %v", err)
	}
	if string(content) != "saved-image" {
		t.Fatalf("file content = %q, want saved-image", string(content))
	}
}

func TestScreenshotTypedErrors(t *testing.T) {
	tests := []struct {
		name   string
		status int
		body   string
		assert func(*testing.T, error)
	}{
		{
			name:   "authentication",
			status: http.StatusUnauthorized,
			body:   `{"error":"API key required"}`,
			assert: func(t *testing.T, err error) {
				var authErr *AuthenticationError
				if !errors.As(err, &authErr) {
					t.Fatalf("error = %T, want AuthenticationError", err)
				}
				if authErr.Code != "authentication_error" || authErr.StatusCode != http.StatusUnauthorized {
					t.Fatalf("auth error = %+v", authErr)
				}
			},
		},
		{
			name:   "insufficient credits",
			status: http.StatusPaymentRequired,
			body:   `{"error":"Insufficient credits","creditBalance":5}`,
			assert: func(t *testing.T, err error) {
				var creditsErr *InsufficientCreditsError
				if !errors.As(err, &creditsErr) {
					t.Fatalf("error = %T, want InsufficientCreditsError", err)
				}
				if creditsErr.Balance != 5 {
					t.Fatalf("balance = %d, want 5", creditsErr.Balance)
				}
			},
		},
		{
			name:   "invalid key",
			status: http.StatusForbidden,
			body:   `{"error":"Invalid API key"}`,
			assert: func(t *testing.T, err error) {
				var keyErr *InvalidAPIKeyError
				if !errors.As(err, &keyErr) {
					t.Fatalf("error = %T, want InvalidAPIKeyError", err)
				}
			},
		},
		{
			name:   "screenshot failed",
			status: http.StatusInternalServerError,
			body:   `{"message":"Render timed out"}`,
			assert: func(t *testing.T, err error) {
				var failedErr *ScreenshotFailedError
				if !errors.As(err, &failedErr) {
					t.Fatalf("error = %T, want ScreenshotFailedError", err)
				}
				if failedErr.Message != "Render timed out" {
					t.Fatalf("message = %q, want Render timed out", failedErr.Message)
				}
			},
		},
		{
			name:   "unknown json",
			status: http.StatusTooManyRequests,
			body:   `{"message":"Rate limited"}`,
			assert: func(t *testing.T, err error) {
				var apiErr *APIError
				if !errors.As(err, &apiErr) {
					t.Fatalf("error = %T, want APIError", err)
				}
				if apiErr.Code != "unknown_error" || apiErr.StatusCode != http.StatusTooManyRequests {
					t.Fatalf("api error = %+v", apiErr)
				}
			},
		},
		{
			name:   "non json",
			status: http.StatusBadGateway,
			body:   "Bad Gateway",
			assert: func(t *testing.T, err error) {
				var apiErr *APIError
				if !errors.As(err, &apiErr) {
					t.Fatalf("error = %T, want APIError", err)
				}
				if !strings.Contains(apiErr.Message, "Bad Gateway") {
					t.Fatalf("message = %q, want Bad Gateway", apiErr.Message)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.status)
				_, _ = w.Write([]byte(tt.body))
			}))
			defer server.Close()

			client := NewClient(testAPIKey, WithBaseURL(server.URL))
			_, err := client.Screenshot(context.Background(), ScreenshotOptions{
				URL: "https://example.com",
			})
			if err == nil {
				t.Fatal("expected error")
			}
			tt.assert(t, err)
		})
	}
}
