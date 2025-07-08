package onramper

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/subdialia/fiat-ramp-service/pkg/database"
	rmp "github.com/subdialia/fiat-ramp-service/pkg/onrampclient"
	"go.uber.org/zap"
)

// SetupRouter initializes API routes for the Fiat Ramp Service.
func SetupRouter(client *rmp.Client, dbClient *database.GraphQLClient, webhookSecret string) (*gin.Engine, error) {
	router := gin.New()
	logger := zap.L()

	// Add middleware
	router.Use(gin.Recovery()) // Default panic recovery
	router.Use(func(c *gin.Context) {
		start := time.Now()
		c.Next()
		logger.Info("Request",
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.Int("status", c.Writer.Status()),
			zap.Duration("duration", time.Since(start)),
		)
	})

	// Add Prometheus metrics endpoint
	router.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// CORS Middleware
	router.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	})

	// Health Check Endpoint
	router.GET("/health", func(c *gin.Context) {
		logger.Info("Health check requested")
		c.JSON(http.StatusOK, gin.H{"message": "Fiat Ramp Service is running"})
	})

	// Create OnramperManager
	onramperManager := NewOnramperManager(
		client,        // APIClient (*rmp.Client)
		dbClient,      // dbClient
		logger,        // logger
		webhookSecret, // webhookSecret
		client,        // onramperClient (rmp.OnRamperClient interface)
	)

	// Define API routes
	router.GET("/supported", onramperManager.GetCurrencies)
	router.GET("/supported/payment-types", onramperManager.GetPaymentTypes)
	router.GET("supported/payment-types/:source", onramperManager.GetPaymentsByCurrency)
	router.GET("supported/defaults/:all", onramperManager.GetDefaults)
	router.POST("checkout/intent", onramperManager.InitiateTransaction)
	router.GET("/transactions_list", onramperManager.ListTransactions)
	router.GET("/transactions/:transaction_id", onramperManager.GetTransactionByID)
	router.GET("/quotes/:source/:destination", onramperManager.GetQuotes)
	router.GET("/supported/assets", onramperManager.GetAssets)
	router.GET("/supported/onramps", onramperManager.GetOnramps)
	router.GET("/supported/onramps/all", onramperManager.GetOnrampMetadata)
	router.GET("/supported/crypto", onramperManager.GetCryptoByFiat)
	router.POST("/transactions/confirm", onramperManager.ConfirmSellTransaction)
	router.POST("/webhook/onramper", onramperManager.WebhookHandler)

	return router, nil
}
