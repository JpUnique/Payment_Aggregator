package cmd

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"

	"github.com/subdialia/fiat-ramp-service/pkg/database"
	rmp "github.com/subdialia/fiat-ramp-service/pkg/onrampclient"
	"github.com/subdialia/fiat-ramp-service/pkg/onramper"
)

//nolint:gochecknoglobals // cfgFile is a command-line flag, conventionally global for Cobra.
var (
	cfgFile  string
	logLevel string
)

//nolint:gochecknoglobals // rootCmd is the entry point for the Cobra CLI application.
var rootCmd = &cobra.Command{
	Use:   "fiat-ramp-service",
	Short: "Fiat Ramp Service CLI",
	Long:  "The Fiat on/off ramp service integrates with the Onramper API to facilitate transactions.",
	RunE: func(cmd *cobra.Command, args []string) error {
		logger, err := zap.NewProduction()
		if err != nil {
			fmt.Printf("failed to initialize logger: %v\n", err)
			os.Exit(100)
		}

		//defer logger.Sync()
		defer func(l *zap.Logger) {
			syncErr := l.Sync()
			if syncErr != nil {
				// Handle error (log to stderr since logger might be unavailable)
				fmt.Fprintf(os.Stderr, "failed to sync logger: %v\n", syncErr)
			}
		}(logger)
		zap.ReplaceGlobals(logger)

		logger.Info("Starting fiat-ramp-service API server")

		// Load Config from environment variables
		baseURL := viper.GetString("ONRAMPER_BASE_URL")
		if baseURL == "" {
			logger.Fatal("ONRAMPER_BASE_URL is required")
		}

		apiKey := viper.GetString("ONRAMPER_API_KEY")
		if apiKey == "" {
			logger.Fatal("ONRAMPER_API_KEY is required")
		}

		webhookSecret := viper.GetString("ONRAMPER_WEBHOOK_SECRET")
		if webhookSecret == "" {
			logger.Fatal("ONRAMPER_WEBHOOK_SECRET is required")
		}

		apiPort := viper.GetString("API_PORT")
		if apiPort == "" {
			logger.Fatal("API_PORT is required")
		}

		metricsPort := viper.GetString("METRICS_PORT")
		if metricsPort == "" {
			logger.Fatal("METRICS_PORT is required")
		}
		// Initialize Hasura GraphQL Client
		hasuraEndpoint := viper.GetString("HASURA_GRAPHQL_ENDPOINT")
		if hasuraEndpoint == "" {
			logger.Fatal("HASURA_ENDPOINT is required")
		}
		// Initialize Hasura GraphQL Client
		hasuraSecret := viper.GetString("HASURA_GRAPHQL_ADMIN_SECRET")
		if hasuraSecret == "" {
			logger.Fatal("HASURA_GRAPHQL_ADMIN_SECRET is required")
		}
		graphQLClient := database.NewGraphQLClient(hasuraEndpoint, hasuraSecret, logger)

		// Test Hasura Client with a Simple Query
		testQuery := `
			query TestHasuraAccess {
				__typename
			}
		`
		var result struct {
			Data struct {
				Typename string `json:"__typename"`
			} `json:"data"`
		}

		err = graphQLClient.ExecuteQuery(context.Background(), testQuery, nil, &result)
		if err != nil {
			logger.Error("Failed to execute Hasura query", zap.Error(err))
		} else {
			logger.Info("Hasura query executed successfully", zap.String("__typename", result.Data.Typename))
		}

		// Initialize Onramper Client
		client := rmp.NewClient(baseURL, apiKey, webhookSecret, logger)

		onramperAPIClient, ok := client.(*rmp.Client)
		if !ok {
			return fmt.Errorf("internal error: failed to assert OnRamper client type: expected *rmp.Client, got %T", client)
		}
		// Setup router (Pass webhookSecret)
		router, err := onramper.SetupRouter(onramperAPIClient, graphQLClient, webhookSecret)
		if err != nil { // This checks the error from SetupRouter
			return fmt.Errorf("failed to setup router: %w", err)
		}

		// Start the API server on port 9999
		apiServer := &http.Server{
			Addr:         ":" + apiPort,
			Handler:      router,
			ReadTimeout:  15 * time.Second,
			WriteTimeout: 15 * time.Second,
		}

		// Start the Metrics server on port 8080
		metricsServer := &http.Server{
			Addr:         ":" + metricsPort,
			Handler:      promhttp.Handler(), // Serves Prometheus metrics
			ReadTimeout:  5 * time.Second,
			WriteTimeout: 5 * time.Second,
		}

		// Graceful shutdown handling
		stop := make(chan os.Signal, 1)
		signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

		// Start the Metrics Server
		go func() {
			logger.Info("Metrics server started", zap.String("port", metricsPort))
			err = metricsServer.ListenAndServe()
			if err != nil && !errors.Is(err, http.ErrServerClosed) {
				logger.Fatal("Metrics server failed", zap.Error(err))
			}
		}()

		// Start the API Server
		go func() {
			logger.Info("API server started", zap.String("port", apiPort))
			err = apiServer.ListenAndServe()
			if err != nil && !errors.Is(err, http.ErrServerClosed) {
				logger.Fatal("API server failed", zap.Error(err))
			}

		}()

		// Wait for termination signal
		<-stop
		logger.Info("Shutting down servers...")

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		// Shutdown APIserver
		err = apiServer.Shutdown(ctx)
		if err != nil {
			logger.Error("Failed to shut down API server properly", zap.Error(err))
		}

		// Shutdown Metric Server
		err = metricsServer.Shutdown(ctx)
		if err != nil {
			logger.Error("Failed to shut down Metrics server properly", zap.Error(err))
		}

		logger.Info("Servers gracefully stopped")
		return nil
	},
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(100)
		os.Exit(100)
	}
}

//nolint:gochecknoinits // Cobra's recommended way to initialize flags and config.
func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", ".env", "config file (default is .env)")
	rootCmd.PersistentFlags().StringVar(&logLevel, "log-level", "", "logging level (debug, info, warn, error, dpanic, panic, fatal)")
}

func initConfig() {
	viper.SetConfigFile(cfgFile)
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil {
		_, _ = fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}
}
