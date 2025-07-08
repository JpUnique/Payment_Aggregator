package database

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/hasura/go-graphql-client"
	"github.com/subdialia/fiat-ramp-service/pkg/models"
	"go.uber.org/zap"
)

// GraphQLClient represents a client for database operations.
type GraphQLClient struct {
	client *graphql.Client
	logger *zap.Logger
}

// NewGraphQLClient creates a new GraphQL client with the provided endpoint and admin secret.
func NewGraphQLClient(endpoint, adminSecret string, logger *zap.Logger) *GraphQLClient {
	// Create a custom HTTP client to include the admin secret header
	httpClient := &http.Client{
		Transport: &headerTransport{
			adminSecret: adminSecret,
		},
	}

	// Initialize the GraphQL client with the custom HTTP client.
	client := graphql.NewClient(endpoint, httpClient)
	return &GraphQLClient{
		client: client,
		logger: logger,
	}
}

// headerTransport is a custom HTTP transport to add the admin secret header.
type headerTransport struct {
	adminSecret string
}

// RoundTrip adds the x-hasura-admin-secret header to the request.
func (t *headerTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Set("X-Hasura-Admin-Secret", t.adminSecret)
	return http.DefaultTransport.RoundTrip(req)
}

// ExecuteQuery executes a GraphQL query and returns the result.
func (c *GraphQLClient) ExecuteQuery(ctx context.Context, query string, variables map[string]interface{}, result interface{}) error {
	return c.client.Query(ctx, result, variables)
}

// ExecuteMutation executes a GraphQL mutation and returns the result.
func (c *GraphQLClient) ExecuteMutation(ctx context.Context, mutation string, variables map[string]interface{}, result interface{}) error {
	return c.client.Exec(ctx, mutation, result, variables)
}

func (c *GraphQLClient) UpsertOnramperTransaction(
	ctx context.Context,
	onrampTx *models.WebhookPayload,
	userID string,
) (updatedUserID string, err error) {
	// Prepare variables
	variables := map[string]interface{}{
		"object": map[string]interface{}{
			"user_id":               userID,
			"country":               onrampTx.Country,
			"in_amount":             onrampTx.InAmount,
			"out_amount":            onrampTx.OutAmount,
			"payment_method":        onrampTx.PaymentMethod,
			"source_currency":       onrampTx.SourceCurrency,
			"target_currency":       onrampTx.TargetCurrency,
			"transaction_type":      strings.ToUpper(onrampTx.TransactionType),
			"transaction_status":    onrampTx.Status,
			"transaction_hash":      onrampTx.TransactionHash,
			"partner_context":       onrampTx.PartnerContext,
			"wallet_address":        onrampTx.WalletAddress,
			"onramp_transaction_id": onrampTx.OnrampTransactionID,
			"transaction_id":        onrampTx.TransactionID,
		},
	}
	// GraphQL mutation.
	query := `mutation UpsertFiatTransaction($object: terrace_schema_fiat_transactions_insert_input!) {
  	insert_terrace_schema_fiat_transactions_one(
    object: $object
    on_conflict: {
      constraint: fiat_transactions_uk_transaction_id
      update_columns: [
        country
        in_amount
        out_amount
        payment_method
        source_currency
        target_currency
        transaction_status
        transaction_type
        transaction_hash
        partner_context
        wallet_address
        onramp_transaction_id
      ]
    }
  ) {
    user_id
    transaction_id
    transaction_status
  }
}`
	// Define result structure.
	type resultResponse struct {
		InsertTerraceSchemaFiatTransactionsOne struct {
			UserID            string `json:"user_id"`
			TransactionID     string `json:"transaction_id"`
			TransactionStatus string `json:"transaction_status"`
		} `json:"insert_terrace_schema_fiat_transactions_one"`
	}

	// Use a map to store the raw result.
	result := resultResponse{}
	// Execute with proper query and result handling
	raw, err := c.client.ExecRaw(ctx, query, variables)

	if err != nil {
		err = fmt.Errorf("failed to execute mutation: %w", err)
		return updatedUserID, err
	}
	err = json.Unmarshal(raw, &result)
	if err != nil {
		return updatedUserID, err
	}
	// Verify response
	if result.InsertTerraceSchemaFiatTransactionsOne.UserID == "" {
		err = errors.New("database returned empty user ID")
		return updatedUserID, err
	}
	updatedUserID = result.InsertTerraceSchemaFiatTransactionsOne.UserID
	return updatedUserID, err
}
func (c *GraphQLClient) GetUserIDFromTransaction(
	ctx context.Context,
	transactionID, onrampTxID, walletAddress string,
) (detail string, err error) {

	var raw []byte

	variables := map[string]interface{}{
		"transaction_id": transactionID,
		"onramp_tx_id":   onrampTxID,
		"wallet_address": walletAddress,
	}
	query := `query GetUserIDFromTransaction(
        $transaction_id: String!,
        $onramp_tx_id: String!,
        $wallet_address: String!
    ) {
        terrace_schema_fiat_transactions(
            where: {
                _or: [
                    {transaction_id: {_eq: $transaction_id}},
                    {onramp_transaction_id: {_eq: $onramp_tx_id}},
                    {wallet_address: {_eq: $wallet_address}}
                ]
            }
            limit: 1
        ) {
            user_id
        }
    }`
	type resultResponse struct {
		TerraceSchemaFiatTransactions struct {
			UserID string `json:"user_id"`
		} `json:"terrace_schema_fiat_transactions"`
	}
	result := resultResponse{}
	raw, err = c.client.ExecRaw(ctx, query, variables)
	if err != nil {
		err = errors.New("failed to query the database")
		return detail, err
	}
	err = json.Unmarshal(raw, &result)
	if err != nil {
		err = errors.New("unable to execute Query")
		return detail, err
	}
	if result.TerraceSchemaFiatTransactions.UserID == "" {
		err = errors.New("no transaction found")
		return detail, err
	}
	return result.TerraceSchemaFiatTransactions.UserID, nil
}

func (c *GraphQLClient) UpdateKYCStatus(
	ctx context.Context,
	userID, newStatus string,
) (updatedStatus string, err error) {

	// Prepare variables
	variables := map[string]interface{}{
		"user_id":    userID,
		"new_status": newStatus,
	}

	// GraphQL mutation with conflict resolution.
	query := `mutation UpsertKYCStatus($user_id: uuid!, $new_status: String!) {
        insert_terrace_schema_id_verification_sessions_one(
            object: {
                user_id: $user_id
                status: $new_status
            }
            on_conflict: {
                constraint: id_verification_sessions_user_id_key
                update_columns: [status, updated_at]
                where: {status: {_neq: "APPROVED"}}
            }
        ) {
            verification_session_id
            status
            updated_at
        }
    }`

	// Response structure.
	type resultResponse struct {
		InsertSession struct {
			SessionID string `json:"verification_session_id"`
			Status    string `json:"status"`
		} `json:"insert_terrace_schema_id_verification_sessions_one"`
	}

	var (
		result resultResponse
		raw    []byte
	)
	raw, err = c.client.ExecRaw(ctx, query, variables)
	if err != nil {
		err = fmt.Errorf("graphql execution failed: %w", err)
		return newStatus, err
	}
	err = json.Unmarshal(raw, &result)
	if err != nil {
		err = fmt.Errorf("response parsing failed: %w", err)
		return newStatus, err
	}
	// Validate response
	if result.InsertSession.Status == "" {
		err = errors.New("empty status in response")
		return newStatus, err
	}
	c.logger.Debug("KYC operation processed",
		zap.String("userID", userID),
		zap.String("status", result.InsertSession.Status))

	return result.InsertSession.Status, nil
}
