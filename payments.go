package zlefpay

import "context"

// PaymentsResource exposes payment operations.
type PaymentsResource struct{ c *Client }

// PaymentParams are the inputs to Payments.Create.
type PaymentParams struct {
	// Amount in major units (e.g. 19.90). Required.
	Amount float64 `json:"amount"`
	// Currency for display: "EUR" (default) or "USD". Settlement is in crypto.
	Currency string `json:"currency,omitempty"`
	// Description shown to the buyer.
	Description string `json:"description,omitempty"`
	// LineItems are an optional itemised breakdown (display only).
	LineItems []LineItem `json:"line_items,omitempty"`
	// Type is "one_time" (default) or "subscription".
	Type string `json:"type,omitempty"`
	// Collect declares which buyer info to gather on the hosted page.
	Collect *Collect `json:"collect,omitempty"`
	// SuccessURL: where to send the buyer after payment (signed query appended).
	SuccessURL string `json:"success_url,omitempty"`
	// CancelURL: where the buyer is sent if they abandon checkout.
	CancelURL string `json:"cancel_url,omitempty"`
	// ExpiresIn is the payment window in seconds (default 3600).
	ExpiresIn int `json:"expires_in,omitempty"`
	// Reference is your own order identifier (returned on the payment + webhooks).
	Reference string `json:"reference,omitempty"`
	// Metadata: arbitrary string key/values echoed back to you.
	Metadata map[string]string `json:"metadata,omitempty"`
}

// Create opens a new payment and returns it, including the hosted checkout URL.
func (r *PaymentsResource) Create(ctx context.Context, p *PaymentParams) (*Payment, error) {
	var out Payment
	if err := r.c.do(ctx, "POST", "/api/v1/payments", p, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Get retrieves a payment by id.
func (r *PaymentsResource) Get(ctx context.Context, id string) (*Payment, error) {
	var out Payment
	if err := r.c.do(ctx, "GET", "/api/v1/payments/"+id, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Cancel cancels a pending payment.
func (r *PaymentsResource) Cancel(ctx context.Context, id string) (*Payment, error) {
	var out Payment
	if err := r.c.do(ctx, "POST", "/api/v1/payments/"+id+"/cancel", nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// PaymentListParams filters the payment list.
type PaymentListParams struct {
	Limit  int    // 1..100, default 20
	Status string // optional: pending|processing|paid|expired|canceled
}

// List returns a page of the merchant's payments, newest first.
func (r *PaymentsResource) List(ctx context.Context, p *PaymentListParams) (*PaymentList, error) {
	q := map[string]string{}
	if p != nil {
		if p.Limit > 0 {
			q["limit"] = itoa(p.Limit)
		}
		q["status"] = p.Status
	}
	var out PaymentList
	if err := r.c.do(ctx, "GET", "/api/v1/payments"+encodeQuery(q), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
