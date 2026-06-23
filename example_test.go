package zlefpay_test

import (
	"context"
	"fmt"
	"net/http"

	zlefpay "github.com/zlef-fr/zlefpay-go"
)

// Create a one-time payment and redirect the customer to the hosted checkout.
func Example_createPayment() {
	client := zlefpay.New("zk_live_your_key")

	pay, err := client.Payments.Create(context.Background(), &zlefpay.PaymentParams{
		Amount:      19.90,
		Description: "Pro plan — lifetime",
		Collect:     &zlefpay.Collect{Email: true},
		SuccessURL:  "https://shop.example/thanks",
		Metadata:    map[string]string{"order_id": "1234"},
	})
	if err != nil {
		panic(err)
	}
	fmt.Println("Send the customer to:", pay.URL)
}

// Verify and handle an incoming webhook.
func Example_webhook() {
	const secret = "whsec_from_your_dashboard"

	http.HandleFunc("/webhooks/zlefpay", func(w http.ResponseWriter, r *http.Request) {
		body := make([]byte, r.ContentLength)
		r.Body.Read(body)

		event, err := zlefpay.ConstructEvent(body, r.Header.Get("Zlefpay-Signature"), secret)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if event.Type == "payment.paid" {
			pay, _ := event.Payment()
			fmt.Printf("Order %s is paid: %s\n", pay.Reference, pay.ID)
			// fulfil the order …
		}
		w.WriteHeader(http.StatusOK)
	})
}
