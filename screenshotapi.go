// Package screenshotapi provides a Go client for the ScreenshotAPI service.
package screenshotapi

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

const DefaultBaseURL = "https://screenshotapi.to"
const DefaultTimeout = 60 * time.Second

type ImageType string

const (
	PNG  ImageType = "png"
	JPEG ImageType = "jpeg"
	WebP ImageType = "webp"
	PDF  ImageType = "pdf"
)

type ColorScheme string

const (
	Light ColorScheme = "light"
	Dark  ColorScheme = "dark"
)

type WaitUntil string

const (
	Load             WaitUntil = "load"
	DOMContentLoaded WaitUntil = "domcontentloaded"
	NetworkIdle0     WaitUntil = "networkidle0"
	NetworkIdle2     WaitUntil = "networkidle2"
)

type MockupDevice string

const (
	BrowserMockup MockupDevice = "browser"
	IPhoneMockup  MockupDevice = "iphone"
	MacBookMockup MockupDevice = "macbook"
)

// Geolocation configures browser geolocation emulation.
type Geolocation struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Accuracy  float64 `json:"accuracy,omitempty"`
}

// Client is a ScreenshotAPI client.
type Client struct {
	apiKey     string
	baseURL    string
	httpClient *http.Client
}

// Option configures a Client.
type Option func(*Client)

// WithBaseURL sets a custom base URL for the API.
func WithBaseURL(baseURL string) Option {
	return func(c *Client) {
		c.baseURL = strings.TrimRight(baseURL, "/")
	}
}

// WithTimeout sets the HTTP client timeout.
func WithTimeout(timeout time.Duration) Option {
	return func(c *Client) {
		if c.httpClient == nil {
			c.httpClient = &http.Client{}
		}
		c.httpClient.Timeout = timeout
	}
}

// WithHTTPClient sets a custom HTTP client.
func WithHTTPClient(httpClient *http.Client) Option {
	return func(c *Client) {
		if httpClient != nil {
			c.httpClient = httpClient
		}
	}
}

// NewClient creates a new ScreenshotAPI client.
func NewClient(apiKey string, opts ...Option) *Client {
	c := &Client{
		apiKey:  apiKey,
		baseURL: DefaultBaseURL,
		httpClient: &http.Client{
			Timeout: DefaultTimeout,
		},
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// ScreenshotOptions configures a screenshot request.
type ScreenshotOptions struct {
	URL                 string
	HTML                string
	Width               int
	Height              int
	FullPage            bool
	Type                ImageType
	Quality             int
	ColorScheme         ColorScheme
	WaitUntil           WaitUntil
	WaitForSelector     string
	Delay               int
	BlockAds            bool
	RemoveCookieBanners bool
	CSSInject           string
	JSInject            string
	StealthMode         bool
	DevicePixelRatio    int
	Timezone            string
	Locale              string
	CacheTTL            int
	PreloadFonts        bool
	RemoveElements      []string
	RemovePopups        bool
	MockupDevice        MockupDevice
	Geolocation         *Geolocation
}

// Metadata contains response metadata from a screenshot request.
type Metadata struct {
	CreditsRemaining int
	ScreenshotID     string
	DurationMs       int
}

// Result contains the screenshot image data and metadata.
type Result struct {
	Image       []byte
	ContentType string
	Metadata    Metadata
}

// Screenshot takes a screenshot with the given options.
func (c *Client) Screenshot(ctx context.Context, opts ScreenshotOptions) (*Result, error) {
	if opts.URL == "" && opts.HTML == "" {
		return nil, fmt.Errorf("screenshotapi: URL or HTML is required")
	}

	req, err := c.newScreenshotRequest(ctx, opts)
	if err != nil {
		return nil, fmt.Errorf("screenshotapi: failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("screenshotapi: request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return nil, parseErrorResponse(resp)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("screenshotapi: failed to read response: %w", err)
	}

	creditsRemaining, _ := strconv.Atoi(resp.Header.Get("x-credits-remaining"))
	durationMs, _ := strconv.Atoi(resp.Header.Get("x-duration-ms"))
	contentType := resp.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "image/png"
	}

	return &Result{
		Image:       body,
		ContentType: contentType,
		Metadata: Metadata{
			CreditsRemaining: creditsRemaining,
			ScreenshotID:     resp.Header.Get("x-screenshot-id"),
			DurationMs:       durationMs,
		},
	}, nil
}

// Save takes a screenshot and saves it to the specified file path.
func (c *Client) Save(ctx context.Context, opts ScreenshotOptions, path string) (*Metadata, error) {
	result, err := c.Screenshot(ctx, opts)
	if err != nil {
		return nil, err
	}

	if err := os.WriteFile(path, result.Image, 0644); err != nil {
		return nil, fmt.Errorf("screenshotapi: failed to write file: %w", err)
	}

	return &result.Metadata, nil
}

type screenshotRequestBody struct {
	URL                 string       `json:"url,omitempty"`
	HTML                string       `json:"html,omitempty"`
	Width               int          `json:"width,omitempty"`
	Height              int          `json:"height,omitempty"`
	FullPage            bool         `json:"fullPage,omitempty"`
	Type                ImageType    `json:"type,omitempty"`
	Quality             int          `json:"quality,omitempty"`
	ColorScheme         ColorScheme  `json:"colorScheme,omitempty"`
	WaitUntil           WaitUntil    `json:"waitUntil,omitempty"`
	WaitForSelector     string       `json:"waitForSelector,omitempty"`
	Delay               int          `json:"delay,omitempty"`
	BlockAds            bool         `json:"blockAds,omitempty"`
	RemoveCookieBanners bool         `json:"removeCookieBanners,omitempty"`
	CSSInject           string       `json:"cssInject,omitempty"`
	JSInject            string       `json:"jsInject,omitempty"`
	StealthMode         bool         `json:"stealthMode,omitempty"`
	DevicePixelRatio    int          `json:"devicePixelRatio,omitempty"`
	Timezone            string       `json:"timezone,omitempty"`
	Locale              string       `json:"locale,omitempty"`
	CacheTTL            int          `json:"cacheTtl,omitempty"`
	PreloadFonts        bool         `json:"preloadFonts,omitempty"`
	RemoveElements      []string     `json:"removeElements,omitempty"`
	RemovePopups        bool         `json:"removePopups,omitempty"`
	MockupDevice        MockupDevice `json:"mockupDevice,omitempty"`
	Geolocation         *Geolocation `json:"geoLocation,omitempty"`
}

func (c *Client) newScreenshotRequest(ctx context.Context, opts ScreenshotOptions) (*http.Request, error) {
	endpoint := c.baseURL + "/api/v1/screenshot"
	if opts.HTML != "" {
		payload := screenshotRequestBody{
			URL:                 opts.URL,
			HTML:                opts.HTML,
			Width:               opts.Width,
			Height:              opts.Height,
			FullPage:            opts.FullPage,
			Type:                opts.Type,
			Quality:             opts.Quality,
			ColorScheme:         opts.ColorScheme,
			WaitUntil:           opts.WaitUntil,
			WaitForSelector:     opts.WaitForSelector,
			Delay:               opts.Delay,
			BlockAds:            opts.BlockAds,
			RemoveCookieBanners: opts.RemoveCookieBanners,
			CSSInject:           opts.CSSInject,
			JSInject:            opts.JSInject,
			StealthMode:         opts.StealthMode,
			DevicePixelRatio:    opts.DevicePixelRatio,
			Timezone:            opts.Timezone,
			Locale:              opts.Locale,
			CacheTTL:            opts.CacheTTL,
			PreloadFonts:        opts.PreloadFonts,
			RemoveElements:      opts.RemoveElements,
			RemovePopups:        opts.RemovePopups,
			MockupDevice:        opts.MockupDevice,
			Geolocation:         opts.Geolocation,
		}
		body, err := json.Marshal(payload)
		if err != nil {
			return nil, err
		}
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
		if err != nil {
			return nil, err
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("x-api-key", c.apiKey)
		return req, nil
	}

	params := buildScreenshotParams(opts)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint+"?"+params.Encode(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("x-api-key", c.apiKey)
	return req, nil
}

func buildScreenshotParams(opts ScreenshotOptions) url.Values {
	params := url.Values{}
	params.Set("url", opts.URL)
	setIntParam(params, "width", opts.Width)
	setIntParam(params, "height", opts.Height)
	setBoolParam(params, "fullPage", opts.FullPage)
	if opts.Type != "" {
		params.Set("type", string(opts.Type))
	}
	setIntParam(params, "quality", opts.Quality)
	if opts.ColorScheme != "" {
		params.Set("colorScheme", string(opts.ColorScheme))
	}
	if opts.WaitUntil != "" {
		params.Set("waitUntil", string(opts.WaitUntil))
	}
	if opts.WaitForSelector != "" {
		params.Set("waitForSelector", opts.WaitForSelector)
	}
	setIntParam(params, "delay", opts.Delay)
	setBoolParam(params, "blockAds", opts.BlockAds)
	setBoolParam(params, "removeCookieBanners", opts.RemoveCookieBanners)
	if opts.CSSInject != "" {
		params.Set("cssInject", opts.CSSInject)
	}
	if opts.JSInject != "" {
		params.Set("jsInject", opts.JSInject)
	}
	setBoolParam(params, "stealthMode", opts.StealthMode)
	setIntParam(params, "devicePixelRatio", opts.DevicePixelRatio)
	if opts.Timezone != "" {
		params.Set("timezone", opts.Timezone)
	}
	if opts.Locale != "" {
		params.Set("locale", opts.Locale)
	}
	setIntParam(params, "cacheTtl", opts.CacheTTL)
	setBoolParam(params, "preloadFonts", opts.PreloadFonts)
	if len(opts.RemoveElements) > 0 {
		params.Set("removeElements", strings.Join(opts.RemoveElements, ","))
	}
	setBoolParam(params, "removePopups", opts.RemovePopups)
	if opts.MockupDevice != "" {
		params.Set("mockupDevice", string(opts.MockupDevice))
	}
	if opts.Geolocation != nil {
		params.Set("geoLatitude", formatFloat(opts.Geolocation.Latitude))
		params.Set("geoLongitude", formatFloat(opts.Geolocation.Longitude))
		if opts.Geolocation.Accuracy > 0 {
			params.Set("geoAccuracy", formatFloat(opts.Geolocation.Accuracy))
		}
	}
	return params
}

func setBoolParam(params url.Values, key string, value bool) {
	if value {
		params.Set(key, "true")
	}
}

func setIntParam(params url.Values, key string, value int) {
	if value > 0 {
		params.Set(key, strconv.Itoa(value))
	}
}

func formatFloat(value float64) string {
	return strconv.FormatFloat(value, 'f', -1, 64)
}

func parseErrorResponse(resp *http.Response) error {
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return &APIError{
			StatusCode: resp.StatusCode,
			Code:       "unknown_error",
			Message:    fmt.Sprintf("HTTP %d", resp.StatusCode),
		}
	}

	var errResp struct {
		Error         string `json:"error"`
		Message       string `json:"message"`
		Balance       *int   `json:"balance"`
		CreditBalance *int   `json:"creditBalance"`
	}
	if err := json.Unmarshal(body, &errResp); err != nil {
		msg := strings.TrimSpace(string(body))
		if msg == "" {
			msg = resp.Status
		}
		return &APIError{
			StatusCode: resp.StatusCode,
			Code:       "unknown_error",
			Message:    fmt.Sprintf("HTTP %d: %s", resp.StatusCode, msg),
		}
	}

	msg := errResp.Error
	if msg == "" {
		msg = errResp.Message
	}
	if msg == "" {
		msg = resp.Status
	}

	switch resp.StatusCode {
	case 401:
		return &AuthenticationError{APIError{StatusCode: 401, Code: "authentication_error", Message: msg}}
	case 402:
		balance := 0
		if errResp.CreditBalance != nil {
			balance = *errResp.CreditBalance
		} else if errResp.Balance != nil {
			balance = *errResp.Balance
		}
		return &InsufficientCreditsError{
			APIError: APIError{StatusCode: 402, Code: "insufficient_credits", Message: msg},
			Balance:  balance,
		}
	case 403:
		return &InvalidAPIKeyError{APIError{StatusCode: 403, Code: "invalid_api_key", Message: msg}}
	case 500:
		detail := errResp.Message
		if detail == "" {
			detail = msg
		}
		if detail == "" {
			detail = "Screenshot failed"
		}
		return &ScreenshotFailedError{APIError{StatusCode: 500, Code: "screenshot_failed", Message: detail}}
	default:
		return &APIError{StatusCode: resp.StatusCode, Code: "unknown_error", Message: msg}
	}
}
