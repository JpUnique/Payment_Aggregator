package onramper

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/subdialia/fiat-ramp-service/pkg/models"
	"go.uber.org/zap"
)

// WebhookHandler processes incoming webhooks from Onramper.
func (w *OnramperManager) WebhookHandler(c *gin.Context) {
	// Read request body
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		w.Logger.Error("Failed to read webhook body", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}
	// Restore request body for logging/debugging
	c.Request.Body = io.NopCloser(bytes.NewBuffer(body))
	// Validate HMAC Signature
	if !w.ValidateSignature(c.Request.Header.Get("X-Onramper-Webhook-Signature"), body, w.WebhookSecret) {
		w.Logger.Error("Invalid webhook signature")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid signature"})
		return
	}
	// Parse the webhook payload
	var payload models.WebhookPayload
	err = json.Unmarshal(body, &payload)
	if err != nil {
		w.Logger.Error("Failed to store webhook payload in Database", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
	}
	if err != nil {
		w.Logger.Error("Failed to update KYC status", zap.Error(err))
	}
	// Respond to Onramper
	c.JSON(http.StatusOK, gin.H{"message": "Webhook received"})
}

// UpdateTransaction saves webhook data in the database.
func (w *OnramperManager) UpdateTransaction(payload models.WebhookPayload) (returnedUserID string, err error) {
	ctx := context.Background()

	var userID string
	// Convert webhook payload struct
	onrampTx := &models.WebhookPayload{
		Country:             payload.Country,
		InAmount:            payload.InAmount,
		Onramp:              payload.Onramp,
		OnrampTransactionID: payload.OnrampTransactionID,
		OutAmount:           payload.OutAmount,
		PaymentMethod:       payload.PaymentMethod,
		PartnerContext:      payload.PartnerContext,
		SourceCurrency:      payload.SourceCurrency,
		Status:              payload.Status,
		StatusDate:          payload.StatusDate,
		TargetCurrency:      payload.TargetCurrency,
		TransactionHash:     payload.TransactionHash,
		TransactionID:       payload.TransactionID,
		TransactionType:     payload.TransactionType,
		WalletAddress:       payload.WalletAddress,
	}
	if userID == "" {
		err = errors.New("user ID is required")
		return returnedUserID, err
	}
	if onrampTx.TransactionID == "" {
		err = errors.New("user ID is required")
		return returnedUserID, err
	}
	if onrampTx.Status == "" {
		err = errors.New("transaction status is required")
		return returnedUserID, err
	}
	// Check context
	if ctx.Err() != nil {
		err = fmt.Errorf("operation cancelled: %w", ctx.Err())
		return returnedUserID, err
	}
	// Call UpsertOnramperTransaction (which handles everything)
	returnedUserID, err = w.dbClient.UpsertOnramperTransaction(ctx, onrampTx, userID)

	if err != nil {
		w.Logger.Error("Failed to store transaction", zap.Error(err))
		err = fmt.Errorf("failed to store transaction: %w", err)
		return returnedUserID, err
	}
	w.Logger.Info("Transaction stored successfully", zap.String("transactionID", payload.TransactionID), zap.String("userID", returnedUserID))

	return returnedUserID, err
}
func (w *OnramperManager) HandleKYCWebhook(payload *models.WebhookPayload) (kycStatus string, err error) {
	// Validate payload
	if payload == nil {
		err = errors.New("webhook payload cannot be nil")
		return kycStatus, err
	}
	// Extract transaction identifiers
	transactionID := strings.TrimSpace(payload.TransactionID)
	onrampTxID := strings.TrimSpace(payload.OnrampTransactionID)
	walletAddress := strings.TrimSpace(payload.WalletAddress)
	rawStatus := strings.TrimSpace(payload.Status)

	// Validate required fields
	if transactionID == "" && onrampTxID == "" && walletAddress == "" {
		w.Logger.Error("Missing transaction identifiers in webhook")
		err = errors.New("transaction identifiers required")
		return kycStatus, err
	}
	// Get context with timeout
	ctx := context.Background()
	// Resolve userID from transaction data
	userID, err := w.dbClient.GetUserIDFromTransaction(ctx, transactionID, onrampTxID, walletAddress)
	if err != nil {
		w.Logger.Error("User resolution failed",
			zap.String("transactionID", transactionID),
			zap.String("onrampTxID", onrampTxID),
			zap.String("walletAddress", walletAddress),
			zap.Error(err))
		err = fmt.Errorf("user resolution failed: %w", err)
		return kycStatus, err
	}
	// Map transaction status to KYC status
	var newStatus string
	switch strings.ToLower(rawStatus) {
	case "completed":
		newStatus = "APPROVED"
	case "failed", "canceled":
		newStatus = "REJECTED"
	case "pending":
		newStatus = "PENDING"
	default:
		w.Logger.Warn("Unhandled transaction status",
			zap.String("status", rawStatus),
			zap.String("userID", userID))
		err = fmt.Errorf("invalid status: %s", rawStatus)
		return kycStatus, err
	}
	// Update KYC status via GraphQL
	resultStatus, err := w.dbClient.UpdateKYCStatus(ctx, userID, newStatus)
	if err != nil {
		w.Logger.Error("KYC status update failed",
			zap.String("userID", userID),
			zap.String("status", newStatus),
			zap.Error(err))
		err = fmt.Errorf("kyc update failed: %w", err)
		return kycStatus, err
	}
	w.Logger.Info("KYC status updated",
		zap.String("userID", userID),
		zap.String("originalStatus", rawStatus),
		zap.String("kycStatus", resultStatus))

	return resultStatus, err
}

// ValidateSignature verifies the HMAC signature of the webhook payload.
func (w *OnramperManager) ValidateSignature(receivedSignature string, payload []byte, secret string) bool {
	if secret == "" {
		w.Logger.Error("Webhook secret is missing")
		return false
	}
	// Compute HMAC SHA256 signature
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	expectedMAC := hex.EncodeToString(mac.Sum(nil))

	// Compare with received signature
	return hmac.Equal([]byte(receivedSignature), []byte(expectedMAC))
}
