# zlefpay-go

![views](https://assets.zlef.fr/badge/views/zlef-fr/zlefpay-go.svg)

Official Go SDK for **[ZlefPay](https://pay.zlef.fr)** — accept crypto payments for
products or subscriptions in a few lines. Non-custodial: funds settle straight to
your own wallet addresses. No KYC.

```bash
go get github.com/zlef-fr/zlefpay-go
```

## Quick start

```go
package main

import (
    "context"
    "fmt"

    zlefpay "github.com/zlef-fr/zlefpay-go"
)

func main() {
    client := zlefpay.New("zk_live_your_key")

    pay, err := client.Payments.Create(context.Background(), &zlefpay.PaymentParams{
        Amount:      19.90,
        Description: "Pro plan",
        Collect:     &zlefpay.Collect{Email: true},
        SuccessURL:  "https://shop.example/thanks",
        Metadata:    map[string]string{"order_id": "1234"},
    })
    if err != nil {
        panic(err)
    }
    fmt.Println("Redirect your customer to:", pay.URL)
}
```

## Sandbox

Test keys (`zk_test_…`) automatically target the sandbox instance — simulated
funds, separate data, a "Simulate payment" button on the checkout page.

```go
client := zlefpay.New("zk_test_your_key")        // → pay-sandbox.zlef.fr
// or force it explicitly
client := zlefpay.New(key, zlefpay.WithSandbox())
```

## Subscriptions

```go
sub, _ := client.Subscriptions.Create(ctx, &zlefpay.SubscriptionParams{
    Amount:      9.99,
    Interval:    "month",
    Description: "Membership",
})
fmt.Println("First payment:", sub.PaymentURL)
```

## Collecting buyer information

```go
Collect: &zlefpay.Collect{
    Name:    true,
    Email:   true,
    Address: true,
    Fields: []zlefpay.Field{
        {Key: "vat", Label: "VAT number", Required: false},
        {Key: "plan", Label: "Plan", Type: "select", Options: []string{"basic", "pro"}},
    },
},
```

## Webhooks

ZlefPay signs every webhook with your endpoint's signing secret. Always verify:

```go
event, err := zlefpay.ConstructEvent(body, r.Header.Get("Zlefpay-Signature"), secret)
if err != nil { /* reject */ }

switch event.Type {
case "payment.paid":
    pay, _ := event.Payment()
    // fulfil pay.Reference
case "subscription.renewed":
    sub, _ := event.Subscription()
    _ = sub
}
```

Events: `payment.created`, `payment.processing`, `payment.paid`,
`payment.expired`, `payment.canceled`, `subscription.created`,
`subscription.payment_due`, `subscription.renewed`, `subscription.canceled`.

## Errors

API errors are returned as `*zlefpay.Error` with `Code`, `Message` and `StatusCode`.

```go
if e, ok := err.(*zlefpay.Error); ok {
    fmt.Println(e.Code, e.StatusCode)
}
```

## License

MIT
