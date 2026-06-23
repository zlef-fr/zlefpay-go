package zlefpay

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"strconv"
	"strings"
	"time"
)

// Event is a webhook event delivered to a merchant endpoint.
type Event struct {
	ID      string          `json:"id"`
	Type    string          `json:"type"`
	Created int64           `json:"created"`
	Data    struct {
		Object json.RawMessage `json:"object"`
	} `json:"data"`
}

// Payment decodes the event's object as a Payment (for payment.* events).
func (e *Event) Payment() (*Payment, error) {
	var p Payment
	if err := json.Unmarshal(e.Data.Object, &p); err != nil {
		return nil, err
	}
	return &p, nil
}

// Subscription decodes the event's object as a Subscription (for subscription.* events).
func (e *Event) Subscription() (*Subscription, error) {
	var s Subscription
	if err := json.Unmarshal(e.Data.Object, &s); err != nil {
		return nil, err
	}
	return &s, nil
}

// Errors returned by signature verification.
var (
	ErrInvalidSignature = errors.New("zlefpay: webhook signature mismatch")
	ErrSignatureFormat  = errors.New("zlefpay: malformed Zlefpay-Signature header")
	ErrSignatureExpired = errors.New("zlefpay: webhook timestamp outside tolerance")
)

// DefaultTolerance is the accepted clock skew for webhook timestamps.
const DefaultTolerance = 5 * time.Minute

// ConstructEvent verifies the signature header against the raw request body
// using the merchant's webhook signing secret, then parses the event. This is
// the recommended way to handle incoming webhooks.
func ConstructEvent(payload []byte, sigHeader, secret string) (*Event, error) {
	if err := VerifySignatureTol(payload, sigHeader, secret, DefaultTolerance); err != nil {
		return nil, err
	}
	var e Event
	if err := json.Unmarshal(payload, &e); err != nil {
		return nil, err
	}
	return &e, nil
}

// VerifySignature checks a webhook signature with the default tolerance.
func VerifySignature(payload []byte, sigHeader, secret string) error {
	return VerifySignatureTol(payload, sigHeader, secret, DefaultTolerance)
}

// VerifySignatureTol checks the "t=…,v1=…" signature header. The signed message
// is `${t}.${rawBody}` (HMAC-SHA256, hex). A tolerance of 0 skips the freshness
// check.
func VerifySignatureTol(payload []byte, sigHeader, secret string, tolerance time.Duration) error {
	var ts, sig string
	for _, part := range strings.Split(sigHeader, ",") {
		kv := strings.SplitN(strings.TrimSpace(part), "=", 2)
		if len(kv) != 2 {
			continue
		}
		switch kv[0] {
		case "t":
			ts = kv[1]
		case "v1":
			sig = kv[1]
		}
	}
	if ts == "" || sig == "" {
		return ErrSignatureFormat
	}
	if tolerance > 0 {
		t, err := strconv.ParseInt(ts, 10, 64)
		if err != nil {
			return ErrSignatureFormat
		}
		if d := time.Since(time.Unix(t, 0)); d > tolerance || d < -tolerance {
			return ErrSignatureExpired
		}
	}
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(ts + "." + string(payload)))
	expected := hex.EncodeToString(mac.Sum(nil))
	if !hmac.Equal([]byte(expected), []byte(sig)) {
		return ErrInvalidSignature
	}
	return nil
}
