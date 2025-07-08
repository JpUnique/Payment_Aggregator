package models

import "time"

// WebhookPayload represents the webhook payload received from Onramper.
type WebhookPayload struct {
	Country             string    `json:"country"`
	InAmount            float64   `json:"inAmount"`
	Onramp              string    `json:"onramp"`
	OnrampTransactionID string    `json:"onrampTransactionId"`
	OutAmount           float64   `json:"outAmount"`
	PaymentMethod       string    `json:"paymentMethod"`
	PartnerContext      string    `json:"partnerContext"`
	SourceCurrency      string    `json:"sourceCurrency"`
	Status              string    `json:"status"`
	StatusDate          time.Time `json:"statusDate"`
	TargetCurrency      string    `json:"targetCurrency"`
	TransactionID       string    `json:"transactionId"`
	TransactionType     string    `json:"transactionType"`
	TransactionHash     string    `json:"transactionHash"`
	WalletAddress       string    `json:"walletAddress"`
}
