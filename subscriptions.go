package zlefpay

import "context"

// SubscriptionsResource exposes subscription operations.
type SubscriptionsResource struct{ c *Client }

// AccountResource exposes the authenticated merchant account.
type AccountResource struct{ c *Client }

// SubscriptionParams are the inputs to Subscriptions.Create.
type SubscriptionParams struct {
	// Amount per cycle in major units. Required.
	Amount float64 `json:"amount"`
	// Currency for display: "EUR" (default) or "USD".
	Currency string `json:"currency,omitempty"`
	// Interval: "day", "week", "month" (default) or "year".
	Interval string `json:"interval,omitempty"`
	// IntervalCount multiplies the interval (e.g. every 3 months). Default 1.
	IntervalCount int `json:"interval_count,omitempty"`
	// Description shown to the buyer.
	Description string `json:"description,omitempty"`
	// Collect declares which buyer info to gather.
	Collect *Collect `json:"collect,omitempty"`
	// SuccessURL / CancelURL for the first (and each renewal) checkout.
	SuccessURL string `json:"success_url,omitempty"`
	CancelURL  string `json:"cancel_url,omitempty"`
	// Reference is your own identifier.
	Reference string `json:"reference,omitempty"`
	// Metadata echoed back to you.
	Metadata map[string]string `json:"metadata,omitempty"`
}

// Create opens a subscription and its first checkout payment. The returned
// Subscription includes PaymentURL — send the buyer there to start the cycle.
func (r *SubscriptionsResource) Create(ctx context.Context, p *SubscriptionParams) (*Subscription, error) {
	var out Subscription
	if err := r.c.do(ctx, "POST", "/api/v1/subscriptions", p, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Get retrieves a subscription by id.
func (r *SubscriptionsResource) Get(ctx context.Context, id string) (*Subscription, error) {
	var out Subscription
	if err := r.c.do(ctx, "GET", "/api/v1/subscriptions/"+id, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Cancel stops a subscription from renewing.
func (r *SubscriptionsResource) Cancel(ctx context.Context, id string) (*Subscription, error) {
	var out Subscription
	if err := r.c.do(ctx, "POST", "/api/v1/subscriptions/"+id+"/cancel", nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Get returns the authenticated merchant account.
func (r *AccountResource) Get(ctx context.Context) (*Account, error) {
	var out Account
	if err := r.c.do(ctx, "GET", "/api/v1/account", nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
