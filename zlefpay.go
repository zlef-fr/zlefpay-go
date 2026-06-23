// Package zlefpay is the official Go SDK for ZlefPay — accept crypto payments
// for products or subscriptions with a few lines of code.
//
//	client := zlefpay.New("zk_live_…")
//	pay, err := client.Payments.Create(ctx, &zlefpay.PaymentParams{
//	    Amount:      19.90,
//	    Description: "Pro plan",
//	    SuccessURL:  "https://shop.example/thanks",
//	})
//	// redirect your customer to pay.URL
//
// Sandbox (simulated funds, no real crypto):
//
//	client := zlefpay.New("zk_test_…", zlefpay.WithSandbox())
//
// Webhooks are verified with the merchant's signing secret:
//
//	event, err := zlefpay.ConstructEvent(body, r.Header.Get("Zlefpay-Signature"), secret)
package zlefpay

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const (
	// Version of this SDK (sent as part of the User-Agent).
	Version = "1.0.0"
	// LiveBaseURL is the production ZlefPay API.
	LiveBaseURL = "https://pay.zlef.fr"
	// SandboxBaseURL is the test instance — simulated funds, separate data.
	SandboxBaseURL = "https://pay-sandbox.zlef.fr"
)

// Client is a ZlefPay API client. Create one with New and reuse it.
type Client struct {
	apiKey     string
	baseURL    string
	httpClient *http.Client
	userAgent  string

	// Resources.
	Payments      *PaymentsResource
	Subscriptions *SubscriptionsResource
	Account       *AccountResource
}

// Option configures a Client.
type Option func(*Client)

// WithSandbox points the client at the sandbox instance (test keys, fake funds).
func WithSandbox() Option { return func(c *Client) { c.baseURL = SandboxBaseURL } }

// WithBaseURL overrides the API base URL (useful for self-hosting or tests).
func WithBaseURL(u string) Option { return func(c *Client) { c.baseURL = strings.TrimRight(u, "/") } }

// WithHTTPClient supplies a custom *http.Client (timeouts, proxies, …).
func WithHTTPClient(h *http.Client) Option { return func(c *Client) { c.httpClient = h } }

// New returns a Client authenticated with the given API key. Keys beginning
// with "zk_test_" automatically default to the sandbox instance.
func New(apiKey string, opts ...Option) *Client {
	c := &Client{
		apiKey:     apiKey,
		baseURL:    LiveBaseURL,
		httpClient: &http.Client{Timeout: 30 * time.Second},
		userAgent:  "zlefpay-go/" + Version,
	}
	if strings.HasPrefix(apiKey, "zk_test_") {
		c.baseURL = SandboxBaseURL
	}
	for _, o := range opts {
		o(c)
	}
	c.Payments = &PaymentsResource{c}
	c.Subscriptions = &SubscriptionsResource{c}
	c.Account = &AccountResource{c}
	return c
}

// Error is a structured API error returned by ZlefPay.
type Error struct {
	Code          string `json:"code"`
	Message       string `json:"message"`
	StatusCode    int    `json:"-"`
	ShortfallCnts int    `json:"shortfall_cents,omitempty"`
	Fields        []string `json:"fields,omitempty"`
}

func (e *Error) Error() string {
	return fmt.Sprintf("zlefpay: %s (%s, http %d)", e.Message, e.Code, e.StatusCode)
}

type errorEnvelope struct {
	Error *Error `json:"error"`
}

// do performs an authenticated request and decodes the JSON response into out.
func (c *Client) do(ctx context.Context, method, path string, body any, out any) error {
	var rdr io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return err
		}
		rdr = bytes.NewReader(b)
	}
	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, rdr)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("User-Agent", c.userAgent)
	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(io.LimitReader(resp.Body, 5<<20))
	if err != nil {
		return err
	}
	if resp.StatusCode >= 400 {
		var env errorEnvelope
		if json.Unmarshal(data, &env) == nil && env.Error != nil {
			env.Error.StatusCode = resp.StatusCode
			return env.Error
		}
		return &Error{Code: "http_error", Message: string(data), StatusCode: resp.StatusCode}
	}
	if out != nil && len(data) > 0 {
		return json.Unmarshal(data, out)
	}
	return nil
}

// ── query helpers ──

func encodeQuery(values map[string]string) string {
	if len(values) == 0 {
		return ""
	}
	q := url.Values{}
	for k, v := range values {
		if v != "" {
			q.Set(k, v)
		}
	}
	if len(q) == 0 {
		return ""
	}
	return "?" + q.Encode()
}

func itoa(i int) string { return strconv.Itoa(i) }
