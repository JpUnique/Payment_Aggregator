package onramper

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"

	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/mock"
	"github.com/subdialia/fiat-ramp-service/pkg/models"
	"go.uber.org/zap"
)

// WebhookManager handles Onramper webhook events.
type WebhookManager struct {
	Logger        *zap.Logger
	WebhookSecret string
	DB            DatabaseService // Dependency for DB service
	KYC           KYCService      // Dependency for KYC service
}

// DatabaseService is an interface for database operations.
type DatabaseService interface {
	UpdateTransaction(payload models.WebhookPayload) (string, error)
}

// KYCService is an interface for KYC operations.
type KYCService interface {
	UpdateKYCStatus(status string, userID string) (bool, error)
}

// WebhookHandler processes incoming webhooks from Onramper.
func (w *WebhookManager) WebhookHandler(c *gin.Context) {
	// Track webhook request count

	// Read request body
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		w.Logger.Error("Failed to read webhook body", zap.Error(err))
		// utils.WebhookRequests.WithLabelValues("invalid_request").Inc()
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	// Restore request body for logging/debugging
	c.Request.Body = io.NopCloser(bytes.NewBuffer(body))

	// Validate HMAC Signature
	if !w.ValidateSignature(c.Request.Header.Get("X-Onramper-Webhook-Signature"), body, w.WebhookSecret) {
		w.Logger.Error("Invalid webhook signature")
		// utils.WebhookRequests.WithLabelValues("invalid_signature").Inc()
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid signature"})
		return
	}

	// Parse the webhook payload
	var payload models.WebhookPayload
	err = json.Unmarshal(body, &payload)
	if err != nil {
		w.Logger.Error("Failed to parse webhook payload", zap.Error(err))
		// utils.WebhookRequests.WithLabelValues("invalid_payload").Inc()
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to parse webhook data"})
		return
	}

	// Upsert the transaction data (and retrieve userID)
	userID, err := w.UpdateTransaction(payload)
	if err != nil {
		w.Logger.Error("Failed to store webhook data in DB", zap.Error(err))
		// utils.WebhookRequests.WithLabelValues("db_failure").Inc()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	// Update KYC status if transaction is canceled or failed
	_, err = w.UpdateKYCStatus(payload.Status, userID)
	if err != nil {
		w.Logger.Error("Failed to update KYC status", zap.Error(err))
	}

	// Respond to Onramper
	c.JSON(http.StatusOK, gin.H{"message": "Webhook received"})
}

// ValidateSignature checks the validity of the webhook signature.
func (w *WebhookManager) ValidateSignature(signature string, body []byte, secret string) bool {
	h := hmac.New(sha256.New, []byte(secret))
	h.Write(body)
	expectedSignature := hex.EncodeToString(h.Sum(nil))

	return signature == expectedSignature
}

// UpdateTransaction stores the transaction into the database (mocked).
func (w *WebhookManager) UpdateTransaction(payload models.WebhookPayload) (string, error) {
	// Mocked DB interaction
	// In a real scenario, this would interact with the database and return the user ID
	return "user123", nil
}

// UpdateKYCStatus updates the KYC status (mocked).
func (w *WebhookManager) UpdateKYCStatus(status string, userID string) (bool, error) {
	// Mocked KYC update
	// In a real scenario, this would update the KYC status in the database
	return true, nil
}

// MockDatabaseService is a mock implementation of the database service.
type MockDatabaseService struct {
	mock.Mock
}

func (m *MockDatabaseService) UpdateTransaction(payload models.WebhookPayload) (string, error) {
	args := m.Called(payload)
	return args.String(0), args.Error(1)
}

// MockKYCService is a mock implementation of the KYC service.
type MockKYCService struct {
	mock.Mock
}

func (m *MockKYCService) UpdateKYCStatus(status string, userID string) (bool, error) {
	args := m.Called(status, userID)
	return args.Bool(0), args.Error(1)
}

func TestWebhookHandler(t *testing.T) {
	// Setup
	logger, _ := zap.NewDevelopment()

	// Create mocked database and KYC services
	mockDB := new(MockDatabaseService)
	mockKYC := new(MockKYCService)

	// Create WebhookManager with injected dependencies (via MockDatabaseService and MockKYCService)
	manager := &WebhookManager{
		Logger:        logger,
		WebhookSecret: "test-secret",
		DB:            mockDB,
		KYC:           mockKYC,
	}

	// Test cases
	tests := []struct {
		name           string
		payload        models.WebhookPayload
		signature      string
		mockDBError    error
		mockKYCError   error
		expectedStatus int
	}{
		{
			name: "Valid Webhook",
			payload: models.WebhookPayload{
				Status: "completed",
			},
			signature:      generateHMACSignature(`{"status":"completed"}`, "test-secret"),
			mockDBError:    nil,
			mockKYCError:   nil,
			expectedStatus: http.StatusOK,
		},
		{
			name: "Invalid Signature",
			payload: models.WebhookPayload{
				Status: "completed",
			},
			signature:      "invalid-signature",
			mockDBError:    nil,
			mockKYCError:   nil,
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name: "Database Error",
			payload: models.WebhookPayload{
				Status: "completed",
			},
			signature:      generateHMACSignature(`{"status":"completed"}`, "test-secret"),
			mockDBError:    errors.New("database error"),
			mockKYCError:   nil,
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name: "KYC Update Error",
			payload: models.WebhookPayload{
				Status: "failed",
			},
			signature:      generateHMACSignature(`{"status":"failed"}`, "test-secret"),
			mockDBError:    nil,
			mockKYCError:   errors.New("kyc error"),
			expectedStatus: http.StatusOK, // KYC error is logged but does not fail the request
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Mock database and KYC service behavior
			mockDB.On("UpdateTransaction", tt.payload).Return("user123", tt.mockDBError)
			mockKYC.On("UpdateKYCStatus", tt.payload.Status, "user123").Return(true, tt.mockKYCError)

			// Create a Gin context
			gin.SetMode(gin.TestMode)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			// Marshal payload to JSON
			payloadBytes, _ := json.Marshal(tt.payload)
			c.Request = httptest.NewRequest(http.MethodPost, "/webhook", bytes.NewBuffer(payloadBytes))
			c.Request.Header.Set("X-Onramper-Webhook-Signature", tt.signature)

			// Call the handler
			manager.WebhookHandler(c)

		})
	}
}

// generateHMACSignature generates a valid HMAC signature for testing.
func generateHMACSignature(payload, secret string) string {
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(payload))
	return hex.EncodeToString(h.Sum(nil))
}
