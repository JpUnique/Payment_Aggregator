package onramper

import (
	"context"
	"net/http"
	"strings"

	"time"

	"github.com/gin-gonic/gin"
	"github.com/subdialia/fiat-ramp-service/pkg/database"
	"github.com/subdialia/fiat-ramp-service/pkg/models"
	rmp "github.com/subdialia/fiat-ramp-service/pkg/onrampclient"
	"github.com/subdialia/fiat-ramp-service/pkg/utils"
	"go.uber.org/zap"
)

type OnramperManager struct {
	// API client for external requests
	APIClient *rmp.Client

	// Logger for structured logging
	Logger *zap.Logger

	// Database client (dependency injection)
	dbClient *database.GraphQLClient
	// Webhook secret.
	WebhookSecret string
	// Onramper API Client.
	onramperClient rmp.OnRamperClient
}

func NewOnramperManager(
	apiClient *rmp.Client,
	dbClient *database.GraphQLClient,
	logger *zap.Logger,
	webhookSecret string,
	onramperClient rmp.OnRamperClient,
) *OnramperManager {
	if logger == nil {
		panic("logger cannot be nil")
	}
	return &OnramperManager{
		APIClient:      apiClient,
		dbClient:       dbClient,
		Logger:         logger,
		WebhookSecret:  webhookSecret,
		onramperClient: onramperClient,
	}
}

// GetCurrencies fetches supported currencies from Onramper API.
func (h *OnramperManager) GetCurrencies(c *gin.Context) {
	transactionType := c.DefaultQuery("type", "buy")
	country := c.Query("country")
	subdivision := c.Query("subdivision")

	h.Logger.Info("Query parameters",
		zap.String("type", transactionType),
		zap.String("country", country),
		zap.String("subdivision", subdivision),
	)

	response, err := h.onramperClient.GetCurrencies(c.Request.Context(), country, subdivision, transactionType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
		return
	}
	// Return JSON response
	c.JSON(http.StatusOK, response)
}
func (h *OnramperManager) GetPaymentTypes(c *gin.Context) {
	transactionType := c.DefaultQuery("type", "buy")
	country := c.Query("country")
	isRecurringParam := c.Query("isRecurringPayment")
	isRecurringPayment := utils.ParseBoolOrDefault(h.Logger, isRecurringParam, false)

	h.Logger.Info("Query parameters",
		zap.String("type", transactionType),
		zap.String("country", country),
		zap.Bool("isRecurringPayment", isRecurringPayment),
	)
	response, err := h.onramperClient.GetPaymentTypes(c.Request.Context(), transactionType, isRecurringPayment, country)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
		return
	}
	// Return JSON response
	c.JSON(http.StatusOK, response)
}
func (h *OnramperManager) GetPaymentsByCurrency(c *gin.Context) {
	sourceCurrency := c.Param("source")

	// Parse query parameters
	transactionType := c.DefaultQuery("type", "buy")
	isRecurringParam := c.DefaultQuery("isRecurringPayment", "false")
	destination := c.Query("destination")
	country := c.Query("country")
	subdivision := c.Query("subdivision")
	isRecurringPayment := utils.ParseBoolOrDefault(h.Logger, isRecurringParam, false)

	h.Logger.Info("Query parameters",
		zap.String("transactionType", transactionType),
		zap.String("sourceCurrency", sourceCurrency),
		zap.Bool("isRecurringPayment", isRecurringPayment),
		zap.String("destination", destination),
		zap.String("country", country),
		zap.String("subdivision", subdivision),
	)

	response, err := h.onramperClient.GetPaymentsByCurrency(
		c.Request.Context(),
		sourceCurrency,
		transactionType,
		isRecurringPayment,
		destination,
		country,
		subdivision,
	)
	if err != nil {
		if strings.Contains(err.Error(), "access forbidden") {
			h.Logger.Error("Access forbidden: invalid API key or insufficient permissions", zap.Error(err))
			c.JSON(http.StatusForbidden, gin.H{"error": "Access forbidden: invalid API key or insufficient permissions"})
		} else {
			h.Logger.Error("Failed to fetch payment types", zap.Error(err))
			c.JSON(http.StatusBadGateway, gin.H{"error": "Failed to fetch payment methods"})
		}
		return
	}

	// Check for errors in the PaymentResponse model
	if response.Error != "" {
		h.Logger.Error("Onramper API returned an error", zap.String("error", response.Error))
		c.JSON(http.StatusBadGateway, gin.H{"error": response.Error})
		return
	}

	// Log the response for debugging
	h.Logger.Info("Payment types response", zap.Any("response", response))
	c.JSON(http.StatusOK, response.Message)
}
func (h *OnramperManager) GetDefaults(c *gin.Context) {
	transactionType := c.DefaultQuery("type", "buy")
	country := c.Query("country")
	subdivision := c.Query("subdivision")

	h.Logger.Info("Query parameters",
		zap.String("type", transactionType),
		zap.String("country", country),
		zap.String("subdivision", subdivision),
	)

	response, err := h.onramperClient.GetDefaults(c.Request.Context(), transactionType, country, subdivision)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
		return
	}
	c.JSON(http.StatusOK, response)
}
func (h *OnramperManager) GetAssets(c *gin.Context) {
	h.Logger.Info("Raw query parameters", zap.String("query", c.Request.URL.RawQuery))
	var params models.AssetRequest
	err := c.ShouldBindQuery(&params)
	if err != nil {
		h.Logger.Error("Invalid query parameters", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid query parameters"})
		return
	}

	h.Logger.Info("Query parameters",
		zap.String("type", string(params.Type)),
		zap.String("source", params.Source),
		zap.String("country", params.Country),
		zap.String("subdivision", params.Subdivision),
		zap.String("onramps", params.Onramps),
		zap.String("paymentMethods", params.PaymentMethods),
	)

	response, err := h.onramperClient.GetAssets(c.Request.Context(), &params)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "Failed to fetch supported assets"})
		return
	}
	c.JSON(http.StatusOK, response)
}
func (h *OnramperManager) GetOnramps(c *gin.Context) {
	h.Logger.Info("Raw query parameters", zap.String("query", c.Request.URL.RawQuery))

	var query models.OnrampsQuery
	err := c.ShouldBindQuery(&query)
	if err != nil {
		h.Logger.Error("Invalid query parameters",
			zap.Error(err),
			zap.Any("received_params", map[string]string{
				"type":        c.Query("type"),
				"source":      c.Query("source"),
				"destination": c.Query("destination"),
			}),
		)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid query parameters"})
		return
	}
	response, err := h.onramperClient.GetOnramps(c.Request.Context(), &query)
	if err != nil {
		h.Logger.Error("Failed to fetch supported onramps", zap.Error(err))
		c.JSON(http.StatusBadGateway, gin.H{"error": "Failed to fetch supported onramps"})
		return
	}
	c.JSON(http.StatusOK, response)
}
func (h *OnramperManager) GetOnrampMetadata(c *gin.Context) {
	h.Logger.Info("Raw query parameters", zap.String("query", c.Request.URL.RawQuery))
	transactionType := c.DefaultQuery("type", "buy")
	h.Logger.Info("Query parameters", zap.String("type", transactionType))

	response, err := h.onramperClient.GetOnrampMetadata(c.Request.Context(), transactionType)
	if err != nil {
		h.Logger.Error("Failed to fetch onramp metadata", zap.Error(err))
		c.JSON(http.StatusBadGateway, gin.H{"error": "Failed to fetch onramp metadata"})
		return
	}
	c.JSON(http.StatusOK, response)
}
func (h *OnramperManager) GetCryptoByFiat(c *gin.Context) {
	source := c.Query("source")
	country := c.Query("country")

	if source == "" {
		h.Logger.Error("Missing required query parameter: source")
		c.JSON(http.StatusBadRequest, gin.H{"error": "source is required"})
		return
	}

	h.Logger.Info("Query parameters",
		zap.String("source", source),
		zap.String("country", country),
	)

	response, err := h.onramperClient.GetCryptoByFiat(c.Request.Context(), source, country)
	if err != nil {
		h.Logger.Error("Failed to fetch crypto currencies", zap.Error(err))
		c.JSON(http.StatusBadGateway, gin.H{"error": "Failed to fetch crypto currencies"})
		return
	}
	c.JSON(http.StatusOK, response)
}
func (h *OnramperManager) GetQuotes(c *gin.Context) {
	fiat := c.Param("source")
	crypto := c.Param("destination")

	if fiat == "" || crypto == "" {
		h.Logger.Error("Missing fiat or crypto parameter")
		c.JSON(http.StatusBadRequest, gin.H{"error": "fiat and crypto are required"})
		return
	}

	var queryParams models.QuoteQueryParams
	err := c.ShouldBindQuery(&queryParams)
	if err != nil {
		h.Logger.Error("Invalid query parameters", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid query parameters"})
		return
	}

	h.Logger.Info("Quote query parameters", zap.Any("params", queryParams))

	quotes, err := h.onramperClient.GetQuotes(c.Request.Context(), fiat, crypto, &queryParams)
	if err != nil {
		h.Logger.Error("Failed to fetch quotes", zap.Error(err))
		c.JSON(http.StatusBadGateway, gin.H{"error": "Failed to fetch quotes"})
		return
	}
	c.JSON(http.StatusOK, quotes)
}
func (h *OnramperManager) GetTransactionByID(c *gin.Context) {
	transactionID := c.Param("transaction_id")

	if transactionID == "" {
		h.Logger.Error("Missing transaction ID")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Transaction ID is required"})
		return
	}

	response, err := h.onramperClient.GetTransactionByID(c.Request.Context(), transactionID)
	if err != nil {
		h.Logger.Error("Failed to fetch transaction", zap.Error(err))
		c.JSON(http.StatusBadGateway, gin.H{"error": "Failed to fetch transaction"})
		return
	}
	c.JSON(http.StatusOK, response)
}
func (h *OnramperManager) ListTransactions(c *gin.Context) {
	var query models.TransactionListQuery
	err := c.ShouldBindQuery(&query)
	if err != nil {
		h.Logger.Error("Invalid query parameters", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid query parameters"})
		return
	}
	// Default limit if not provided
	if query.Limit == 0 {
		query.Limit = 50
	}
	response, err := h.onramperClient.ListTransactions(c.Request.Context(), query)
	if err != nil {
		h.Logger.Error("Failed to list transactions", zap.Error(err))
		c.JSON(http.StatusBadGateway, gin.H{"error": "Failed to list transactions"})
		return
	}
	c.JSON(http.StatusOK, response)
}
func (h *OnramperManager) ConfirmSellTransaction(c *gin.Context) {
	txType := c.Param("type")

	h.Logger.Info("Received confirm sell transaction request",
		zap.String("txType", txType),
	)

	response, err := h.onramperClient.ConfirmSellTransaction(c.Request.Context(), txType)
	if err != nil {
		h.Logger.Error("Failed to confirm sell transaction", zap.Error(err))
		c.JSON(http.StatusBadGateway, gin.H{"error": "Failed to confirm sell transaction"})
		return
	}
	c.JSON(http.StatusOK, response)
}
func (h *OnramperManager) InitiateTransaction(c *gin.Context) {
	userID := c.Query("user_id")
	if userID == "" {
		h.Logger.Error("Missing user_id")
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_id is required"})
		return
	}
	// Parse request body
	var payload models.InitiateTransactionRequest
	err := c.ShouldBindJSON(&payload)
	if err != nil {
		h.Logger.Error("Invalid request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}
	if payload.Wallet.Address == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "wallet address required"})
		return
	}
	// Call client to initiate transaction
	response, err := h.onramperClient.InitiateTransaction(c.Request.Context(), payload)
	if err != nil {
		h.Logger.Error("Failed to initiate transaction", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to initiate transaction"})
		return
	}
	txInfo := response.Message.TransactionInformation
	sess := response.Message.SessionInformation

	if txInfo.TransactionID == "" {
		h.Logger.Error("Empty transaction ID in Onramper response")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Missing transaction ID in response"})
	}

	// Build payload for DB
	onrampTx := &models.WebhookPayload{
		Country:             sess.Country,
		InAmount:            sess.Amount,
		Onramp:              sess.Onramp,
		OnrampTransactionID: txInfo.TransactionID,
		OutAmount:           0.0,
		PaymentMethod:       sess.PaymentMethod,
		SourceCurrency:      sess.Source,
		Status:              utils.MapTransactionStatus(response.Message.Status),
		StatusDate:          time.Now().UTC(),
		TargetCurrency:      sess.Destination,
		TransactionID:       txInfo.TransactionID,
		TransactionType:     strings.ToUpper(sess.Type),
		TransactionHash:     "",
		WalletAddress:       sess.Wallet.Address,
	}

	// Insert into DB
	if h.dbClient == nil {
		h.Logger.Error("Database client is nil")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	returnedUserID, err := h.dbClient.UpsertOnramperTransaction(context.Background(), onrampTx, userID)
	if err != nil {
		h.Logger.Error("Failed to insert transaction", zap.Error(err),
			zap.String("user_id", userID),
			zap.String("transaction_status", response.Message.Status),
			zap.String("transaction_id", txInfo.TransactionID),
		)
		c.JSON(http.StatusInternalServerError, gin.H{"Error": "Failed to save transaction"})
		return
	}
	// verify user ID match
	if returnedUserID != userID {
		h.Logger.Warn("User ID mismatch after upsert",
			zap.String("user", userID),
			zap.String("return", returnedUserID),
		)
	}
	h.Logger.Info("Transaction initiated successfully",
		zap.String("transaction_id", txInfo.TransactionID),
		zap.String("user_id", userID),
		zap.String("status", response.Message.Status),
	)
	// Return response
	c.JSON(http.StatusOK, gin.H{
		"status":         response.Message.Status,
		"transaction_id": txInfo.TransactionID,
		"user_id":        userID,
		"redirect_url":   txInfo.URL,
	})
}
