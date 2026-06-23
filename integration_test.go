package zlefpay

import (
	"context"
	"os"
	"testing"
)

// TestLiveBackend runs the SDK against a real ZlefPay instance. It is skipped
// unless ZLEFPAY_TEST_BASE and ZLEFPAY_TEST_KEY are set:
//
//	ZLEFPAY_TEST_BASE=http://127.0.0.1:10064 ZLEFPAY_TEST_KEY=zk_test_… \
//	    go test -run TestLiveBackend -v
func TestLiveBackend(t *testing.T) {
	base := os.Getenv("ZLEFPAY_TEST_BASE")
	key := os.Getenv("ZLEFPAY_TEST_KEY")
	if base == "" || key == "" {
		t.Skip("set ZLEFPAY_TEST_BASE and ZLEFPAY_TEST_KEY to run the integration test")
	}
	c := New(key, WithBaseURL(base))
	ctx := context.Background()

	acct, err := c.Account.Get(ctx)
	if err != nil {
		t.Fatalf("account: %v", err)
	}
	t.Logf("account: %s (%s) accepts %v", acct.User, acct.Mode, acct.AcceptedCoins)

	pay, err := c.Payments.Create(ctx, &PaymentParams{
		Amount:      24.99,
		Description: "Go SDK integration",
		Collect:     &Collect{Email: true},
		SuccessURL:  "https://example.test/ok",
		Metadata:    map[string]string{"sdk": "go"},
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if pay.Status != "pending" || pay.URL == "" || pay.AmountCents != 2499 {
		t.Fatalf("unexpected payment: %+v", pay)
	}
	t.Logf("created %s → %s", pay.ID, pay.URL)

	got, err := c.Payments.Get(ctx, pay.ID)
	if err != nil || got.ID != pay.ID {
		t.Fatalf("get mismatch: %v %+v", err, got)
	}

	canceled, err := c.Payments.Cancel(ctx, pay.ID)
	if err != nil || canceled.Status != "canceled" {
		t.Fatalf("cancel: %v %+v", err, canceled)
	}

	sub, err := c.Subscriptions.Create(ctx, &SubscriptionParams{Amount: 5, Interval: "month", Description: "Go sub"})
	if err != nil || sub.PaymentURL == "" {
		t.Fatalf("subscription: %v %+v", err, sub)
	}
	t.Logf("subscription %s, first payment %s", sub.ID, sub.PaymentURL)

	list, err := c.Payments.List(ctx, &PaymentListParams{Limit: 10})
	if err != nil || len(list.Data) == 0 {
		t.Fatalf("list: %v %+v", err, list)
	}
}
