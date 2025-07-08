package database

import (
	"context"

	"github.com/subdialia/fiat-ramp-service/pkg/models"
)

// QueryClient provides methods for upserting transactions and updating KYC status.
type QueryClient interface {
	// UpsertOnramperTransaction inserts or updates a fiat transaction and returns the user_id.
	UpsertOnramperTransaction(ctx context.Context, onrampTx *models.WebhookPayload, userID string) (updatedUserserID string, err error)
	// UpdateKYCStatus updates the KYC status of a user in the id_verification_sessions table.
	UpdateKYCStatus(ctx context.Context, userID, transactionStatus string) (string, error)
	GetUserIDFromTransaction(ctx context.Context, transactionID, onrampTxID, walletAddress string) (string, error)
}
