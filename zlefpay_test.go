package zlefpay

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// mockServer implements just enough of the ZlefPay API to exercise the SDK.
func mockServer(t *testing.T) *httptest.Server {
	mux := http.NewServeMux()

	auth := func(w http.ResponseWriter, r *http.Request) bool {
		if r.Header.Get("Authorization") != "Bearer zk_test_key" {
			w.WriteHeader(401)
			io.WriteString(w, `{"error":{"code":"unauthorized","message":"bad key"}}`)
			return false
		}
		return true
	}

	mux.HandleFunc("/api/v1/account", func(w http.ResponseWriter, r *http.Request) {
		if !auth(w, r) {
			return
		}
		json.NewEncoder(w).Encode(Account{Object: "account", Mode: "sandbox", User: "demo", BusinessName: "Acme", AcceptedCoins: []string{"btc", "eth"}})
	})

	mux.HandleFunc("/api/v1/payments", func(w http.ResponseWriter, r *http.Request) {
		if !auth(w, r) {
			return
		}
		if r.Method == "POST" {
			var p PaymentParams
			json.NewDecoder(r.Body).Decode(&p)
			if p.Amount <= 0 {
				w.WriteHeader(400)
				io.WriteString(w, `{"error":{"code":"invalid_amount","message":"amount must be positive"}}`)
				return
			}
			w.WriteHeader(201)
			json.NewEncoder(w).Encode(Payment{
				ID: "pay_123", Object: "payment", Status: "pending", Amount: p.Amount,
				AmountCents: int(p.Amount * 100), Currency: "EUR", Description: p.Description,
				Metadata: p.Metadata, URL: "https://pay-sandbox.zlef.fr/pay/pay_123",
				Collect: p.Collect,
			})
			return
		}
		// list
		json.NewEncoder(w).Encode(PaymentList{Object: "list", Data: []Payment{{ID: "pay_123", Status: "paid"}}, HasMore: false})
	})

	mux.HandleFunc("/api/v1/payments/pay_123", func(w http.ResponseWriter, r *http.Request) {
		if !auth(w, r) {
			return
		}
		json.NewEncoder(w).Encode(Payment{ID: "pay_123", Object: "payment", Status: "paid", Amount: 19.9})
	})
	mux.HandleFunc("/api/v1/payments/pay_123/cancel", func(w http.ResponseWriter, r *http.Request) {
		if !auth(w, r) {
			return
		}
		json.NewEncoder(w).Encode(Payment{ID: "pay_123", Object: "payment", Status: "canceled"})
	})
	mux.HandleFunc("/api/v1/payments/pay_missing", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(404)
		io.WriteString(w, `{"error":{"code":"not_found","message":"no such payment"}}`)
	})

	mux.HandleFunc("/api/v1/subscriptions", func(w http.ResponseWriter, r *http.Request) {
		if !auth(w, r) {
			return
		}
		var p SubscriptionParams
		json.NewDecoder(r.Body).Decode(&p)
		w.WriteHeader(201)
		json.NewEncoder(w).Encode(Subscription{
			ID: "sub_1", Object: "subscription", Status: "active", Amount: p.Amount,
			Interval: orDefault(p.Interval, "month"), IntervalCount: 1,
			PaymentURL: "https://pay-sandbox.zlef.fr/pay/pay_999",
			LatestPayment: &Payment{ID: "pay_999", Status: "pending"},
		})
	})

	return httptest.NewServer(mux)
}

func orDefault(s, d string) string {
	if s == "" {
		return d
	}
	return s
}

func newTestClient(t *testing.T, url string) *Client {
	return New("zk_test_key", WithBaseURL(url))
}

func TestCreatePayment(t *testing.T) {
	srv := mockServer(t)
	defer srv.Close()
	c := newTestClient(t, srv.URL)

	pay, err := c.Payments.Create(context.Background(), &PaymentParams{
		Amount:      19.90,
		Description: "Pro plan",
		Collect:     &Collect{Email: true, Fields: []Field{{Key: "company", Label: "Company"}}},
		Metadata:    map[string]string{"order": "A1"},
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if pay.ID != "pay_123" || pay.Status != "pending" {
		t.Fatalf("unexpected payment: %+v", pay)
	}
	if pay.Amount != 19.90 {
		t.Fatalf("amount roundtrip: got %v", pay.Amount)
	}
	if pay.Metadata["order"] != "A1" {
		t.Fatalf("metadata not echoed: %+v", pay.Metadata)
	}
	if pay.URL == "" {
		t.Fatal("missing checkout url")
	}
}

func TestCreatePaymentValidationError(t *testing.T) {
	srv := mockServer(t)
	defer srv.Close()
	c := newTestClient(t, srv.URL)

	_, err := c.Payments.Create(context.Background(), &PaymentParams{Amount: 0})
	if err == nil {
		t.Fatal("expected error")
	}
	apiErr, ok := err.(*Error)
	if !ok {
		t.Fatalf("expected *Error, got %T", err)
	}
	if apiErr.Code != "invalid_amount" || apiErr.StatusCode != 400 {
		t.Fatalf("unexpected error: %+v", apiErr)
	}
}

func TestGetCancelList(t *testing.T) {
	srv := mockServer(t)
	defer srv.Close()
	c := newTestClient(t, srv.URL)
	ctx := context.Background()

	pay, err := c.Payments.Get(ctx, "pay_123")
	if err != nil || !pay.Paid() {
		t.Fatalf("get: %v %+v", err, pay)
	}
	canceled, err := c.Payments.Cancel(ctx, "pay_123")
	if err != nil || canceled.Status != "canceled" {
		t.Fatalf("cancel: %v %+v", err, canceled)
	}
	list, err := c.Payments.List(ctx, &PaymentListParams{Limit: 5, Status: "paid"})
	if err != nil || len(list.Data) != 1 {
		t.Fatalf("list: %v %+v", err, list)
	}
}

func TestNotFound(t *testing.T) {
	srv := mockServer(t)
	defer srv.Close()
	c := newTestClient(t, srv.URL)
	_, err := c.Payments.Get(context.Background(), "pay_missing")
	apiErr, ok := err.(*Error)
	if !ok || apiErr.StatusCode != 404 || apiErr.Code != "not_found" {
		t.Fatalf("expected 404 not_found, got %v", err)
	}
}

func TestUnauthorized(t *testing.T) {
	srv := mockServer(t)
	defer srv.Close()
	c := New("zk_test_wrong", WithBaseURL(srv.URL))
	_, err := c.Account.Get(context.Background())
	apiErr, ok := err.(*Error)
	if !ok || apiErr.StatusCode != 401 {
		t.Fatalf("expected 401, got %v", err)
	}
}

func TestSubscription(t *testing.T) {
	srv := mockServer(t)
	defer srv.Close()
	c := newTestClient(t, srv.URL)
	sub, err := c.Subscriptions.Create(context.Background(), &SubscriptionParams{Amount: 9.99, Interval: "month"})
	if err != nil {
		t.Fatalf("sub create: %v", err)
	}
	if sub.ID != "sub_1" || sub.PaymentURL == "" || sub.LatestPayment == nil {
		t.Fatalf("unexpected subscription: %+v", sub)
	}
}

func TestSandboxDefaultsFromKey(t *testing.T) {
	c := New("zk_test_abc")
	if c.baseURL != SandboxBaseURL {
		t.Fatalf("test key should default to sandbox, got %s", c.baseURL)
	}
	c2 := New("zk_live_abc")
	if c2.baseURL != LiveBaseURL {
		t.Fatalf("live key should default to live, got %s", c2.baseURL)
	}
}

// ── webhook signature: must match the server's `${t}.${body}` HMAC-SHA256 ──

func signLikeServer(secret, ts, body string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(ts + "." + body))
	return hex.EncodeToString(mac.Sum(nil))
}

func TestConstructEvent(t *testing.T) {
	secret := "whsec_test"
	body := `{"id":"evt_1","type":"payment.paid","created":1,"data":{"object":{"id":"pay_123","status":"paid","amount":19.9}}}`
	ts := fmt.Sprintf("%d", time.Now().Unix())
	header := "t=" + ts + ",v1=" + signLikeServer(secret, ts, body)

	ev, err := ConstructEvent([]byte(body), header, secret)
	if err != nil {
		t.Fatalf("construct: %v", err)
	}
	if ev.Type != "payment.paid" {
		t.Fatalf("type: %s", ev.Type)
	}
	pay, err := ev.Payment()
	if err != nil || pay.ID != "pay_123" || !pay.Paid() {
		t.Fatalf("payment decode: %v %+v", err, pay)
	}
}

func TestConstructEventBadSignature(t *testing.T) {
	body := `{"type":"payment.paid"}`
	ts := fmt.Sprintf("%d", time.Now().Unix())
	header := "t=" + ts + ",v1=" + strings.Repeat("0", 64)
	if _, err := ConstructEvent([]byte(body), header, "whsec_test"); err != ErrInvalidSignature {
		t.Fatalf("expected ErrInvalidSignature, got %v", err)
	}
}

func TestConstructEventExpired(t *testing.T) {
	secret := "whsec_test"
	body := `{"type":"payment.paid"}`
	ts := fmt.Sprintf("%d", time.Now().Add(-10*time.Minute).Unix())
	header := "t=" + ts + ",v1=" + signLikeServer(secret, ts, body)
	if _, err := ConstructEvent([]byte(body), header, secret); err != ErrSignatureExpired {
		t.Fatalf("expected ErrSignatureExpired, got %v", err)
	}
}

func TestConstructEventMalformed(t *testing.T) {
	if _, err := ConstructEvent([]byte(`{}`), "garbage", "s"); err != ErrSignatureFormat {
		t.Fatalf("expected ErrSignatureFormat, got %v", err)
	}
}
