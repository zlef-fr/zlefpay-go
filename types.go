package zlefpay

// Field is a custom piece of information a merchant asks the buyer to provide
// on the hosted checkout page.
type Field struct {
	Key      string   `json:"key"`
	Label    string   `json:"label,omitempty"`
	Type     string   `json:"type,omitempty"` // text | email | number | tel | select | textarea
	Required bool     `json:"required,omitempty"`
	Options  []string `json:"options,omitempty"`
}

// Collect declares which buyer information the checkout should gather.
type Collect struct {
	Name    bool    `json:"name,omitempty"`
	Email   bool    `json:"email,omitempty"`
	Address bool    `json:"address,omitempty"`
	Phone   bool    `json:"phone,omitempty"`
	Fields  []Field `json:"fields,omitempty"`
}

// LineItem is one line of an itemised payment (display only).
type LineItem struct {
	Name     string  `json:"name"`
	Amount   float64 `json:"amount"`
	Quantity int     `json:"quantity"`
}

// Address is a postal address collected from the buyer.
type Address struct {
	Line1      string `json:"line1,omitempty"`
	Line2      string `json:"line2,omitempty"`
	City       string `json:"city,omitempty"`
	PostalCode string `json:"postal_code,omitempty"`
	State      string `json:"state,omitempty"`
	Country    string `json:"country,omitempty"`
}

// Customer holds the information collected from the buyer at checkout.
type Customer struct {
	Name    string            `json:"name,omitempty"`
	Email   string            `json:"email,omitempty"`
	Phone   string            `json:"phone,omitempty"`
	Address *Address          `json:"address,omitempty"`
	Fields  map[string]string `json:"fields,omitempty"`
}

// Crypto carries the on-chain details once a buyer selects a coin.
type Crypto struct {
	Coin          string  `json:"coin"`
	Address       string  `json:"address"`
	AmountCoin    float64 `json:"amount_coin"`
	ReceivedCoin  float64 `json:"received_coin"`
	TxID          string  `json:"txid"`
	Confirmations int     `json:"confirmations"`
	URI           string  `json:"uri"`
}

// Payment is a single charge created by a merchant.
type Payment struct {
	ID            string            `json:"id"`
	Object        string            `json:"object"`
	Mode          string            `json:"mode"` // live | sandbox
	Status        string            `json:"status"`
	Amount        float64           `json:"amount"`
	AmountCents   int               `json:"amount_cents"`
	Currency      string            `json:"currency"`
	Description   string            `json:"description"`
	LineItems     []LineItem        `json:"line_items"`
	Type          string            `json:"type"` // one_time | subscription
	Subscription  string            `json:"subscription,omitempty"`
	Reference     string            `json:"reference,omitempty"`
	Metadata      map[string]string `json:"metadata"`
	Collect       *Collect          `json:"collect"`
	Customer      *Customer         `json:"customer"`
	Crypto        *Crypto           `json:"crypto"`
	PaidFromBal   bool              `json:"paid_from_balance"`
	Payer         string            `json:"payer,omitempty"`
	URL           string            `json:"url"`
	SuccessURL    string            `json:"success_url,omitempty"`
	CancelURL     string            `json:"cancel_url,omitempty"`
	Created       int64             `json:"created"`
	ExpiresAt     int64             `json:"expires_at"`
	PaidAt        *int64            `json:"paid_at"`
	CanceledAt    *int64            `json:"canceled_at"`
}

// Paid reports whether the payment has settled.
func (p *Payment) Paid() bool { return p.Status == "paid" }

// Subscription is a recurring crypto payment plan.
type Subscription struct {
	ID               string            `json:"id"`
	Object           string            `json:"object"`
	Mode             string            `json:"mode"`
	Status           string            `json:"status"`
	Amount           float64           `json:"amount"`
	AmountCents      int               `json:"amount_cents"`
	Currency         string            `json:"currency"`
	Description      string            `json:"description"`
	Interval         string            `json:"interval"`
	IntervalCount    int               `json:"interval_count"`
	Reference        string            `json:"reference,omitempty"`
	Metadata         map[string]string `json:"metadata"`
	Collect          *Collect          `json:"collect"`
	Payer            string            `json:"payer,omitempty"`
	AutoRenew        bool              `json:"auto_renew"`
	CurrentPeriodEnd *int64            `json:"current_period_end"`
	Created          int64             `json:"created"`
	CanceledAt       *int64            `json:"canceled_at"`
	LatestPayment    *Payment          `json:"latest_payment,omitempty"`
	PaymentURL       string            `json:"payment_url,omitempty"`
	Payments         []string          `json:"payments"`
}

// Account describes the authenticated merchant.
type Account struct {
	Object        string   `json:"object"`
	Mode          string   `json:"mode"`
	User          string   `json:"user"`
	BusinessName  string   `json:"business_name"`
	BalanceCents  int      `json:"balance_cents"`
	AcceptedCoins []string `json:"accepted_coins"`
	WebhookSet    bool     `json:"webhook_configured"`
}

// PaymentList is a page of payments.
type PaymentList struct {
	Object  string    `json:"object"`
	Data    []Payment `json:"data"`
	HasMore bool      `json:"has_more"`
}
