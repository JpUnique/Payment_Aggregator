package onrampclient

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"

	"github.com/subdialia/fiat-ramp-service/pkg/models"
	"go.uber.org/zap"
)

type OnRamperClient interface {
	GetCurrencies(ctx context.Context, country string, subdivision string, transactionType string) (currrencies models.SupportedCurrenciesResponse, err error)
	GetPaymentTypes(ctx context.Context, transactionType string, isRecurringPayment bool, country string) (paymentTypes models.PaymentTypesResponse, err error)
	GetPaymentsByCurrency(ctx context.Context, sourceCurrency string, transactionType string, isRecurringPayment bool, destination string, country string, subdivision string) (paymentByCurrency models.PaymentResponse, err error)
	GetDefaults(ctx context.Context, transactionType string, conutry string, subdivision string) (defaults models.DefaultsResponse, err error)
	GetAssets(ctx context.Context, paymentParam *models.AssetRequest) (assets models.SupportedAssetsResponse, err error)
	GetOnramps(ctx context.Context, params *models.OnrampsQuery) (onramps models.OnrampResponse, err error)
	GetOnrampMetadata(ctx context.Context, transactionType string) (metadata models.OnrampMetadataResponse, err error)
	GetCryptoByFiat(ctx context.Context, source string, country string) (cryptofiat models.CryptoFiatResponse, err error)
	GetQuotes(ctx context.Context, fiat string, crypto string, quotesParam *models.QuoteQueryParams) (quotes []models.QuoteResponse, err error)
	GetTransactionByID(ctx context.Context, transactionID string) (transactionid models.TransactionResponse, err error)
	ListTransactions(ctx context.Context, ListTransactions models.TransactionListQuery) (transactionlist models.TransactionListResponse, err error)
	InitiateTransaction(ctx context.Context, payload models.InitiateTransactionRequest) (transaction models.InitiateTransactionResponse, err error)
	ConfirmSellTransaction(ctx context.Context, txType string) (confirmation models.SellTransactionConfirmationResponse, err error)
}

const (
	transactionTypeBuy = "buy"
)

// Client manages communication with the Onramper API.
type Client struct {
	BaseURL       string
	APIKey        string
	WebhookSecret string
	HTTPClient    *http.Client
	Logger        *zap.Logger
}

// NewClient initializes a new Onramper API client.
func NewClient(baseURL, apiKey string, webhookSecret string, logger *zap.Logger) OnRamperClient {

	return &Client{
		BaseURL:       baseURL,
		APIKey:        apiKey,
		WebhookSecret: webhookSecret,
		HTTPClient:    &http.Client{},
		Logger:        logger,
	}
}

func (h Client) GetCurrencies(ctx context.Context, country string, subdivision string, transactionType string) (currrencies models.SupportedCurrenciesResponse, err error) {
	// Construct API request URL with query parameters
	h.Logger.Info("Fetching currencies", zap.String("url", h.BaseURL))
	apiURL := fmt.Sprintf("%s/supported?type=%s", h.BaseURL, transactionType)
	if country != "" {
		apiURL += "&country=" + country
	}
	if subdivision != "" {
		apiURL += "&subdivision=" + subdivision
	}
	h.Logger.Info("Fetching currencies", zap.String("url", apiURL))

	// Prepare Onramper API request
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
	if err != nil {
		h.Logger.Error("Failed to create request", zap.Error(err))
		return currrencies, err
	}

	req.Header.Add("Authorization", h.APIKey)

	resp, err := h.HTTPClient.Do(req)
	if err != nil {
		h.Logger.Error("Failed to fetch currencies", zap.Error(err))
		return currrencies, err
	}
	defer resp.Body.Close()
	h.Logger.Info("Received response", zap.Int("status", resp.StatusCode))

	var body []byte
	if resp.StatusCode != http.StatusOK {
		body, err = io.ReadAll(resp.Body)
		if err != nil {
			h.Logger.Error("Failed to read response body", zap.Error(err))
			return currrencies, err
		}
		err = fmt.Errorf("unable to get currencies with status code: %d - message: %s", resp.StatusCode, string(body))
		return currrencies, err
	}

	err = json.NewDecoder(resp.Body).Decode(&currrencies)
	if err != nil {
		h.Logger.Error("Failed to decode response", zap.Error(err))
		err = fmt.Errorf("failed to decode response: %w", err)
		return currrencies, err
	}
	return currrencies, err
}
func (h Client) GetPaymentTypes(ctx context.Context, transactionType string, isRecurringPayment bool, country string) (paymentTypes models.PaymentTypesResponse, err error) {
	// Construct API request URL with query parameters
	apiURL := fmt.Sprintf("%s/supported/payment-types?type=%s", h.BaseURL, transactionType)
	if country != "" {
		apiURL += "&country=" + country
	}
	h.Logger.Info("Fetching payment types", zap.String("url", apiURL))

	// Prepare Onramper API request
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
	if err != nil {
		h.Logger.Error("Failed to create request", zap.Error(err))
		return paymentTypes, err
	}

	req.Header.Add("Authorization", h.APIKey)
	if transactionType == transactionTypeBuy {
		// Convert bool to string: "true" / "false"
		recurringValue := strconv.FormatBool(isRecurringPayment)
		req.Header.Add("X-Is-Recurringpayment", recurringValue)
	}

	resp, err := h.HTTPClient.Do(req)
	if err != nil {
		h.Logger.Error("Failed to fetch payment types", zap.Error(err))
		return paymentTypes, err
	}
	defer resp.Body.Close()

	h.Logger.Info("Received response for payment types", zap.Int("status", resp.StatusCode))

	// Check for non-OK status
	var body []byte
	if resp.StatusCode != http.StatusOK {
		body, err = io.ReadAll(resp.Body)
		if err != nil {
			h.Logger.Error("Failed to read response body", zap.Error(err))
			return paymentTypes, err
		}
		err = fmt.Errorf(
			"unable to get payment types with status code: %d - message: %s",
			resp.StatusCode, string(body),
		)
		return paymentTypes, err
	}
	// Parse the JSON response
	err = json.NewDecoder(resp.Body).Decode(&paymentTypes)
	if err != nil {
		h.Logger.Error("Failed to decode payment types response", zap.Error(err))
		err = fmt.Errorf("failed to decode response: %w", err)
		return paymentTypes, err
	}

	return paymentTypes, err
}
func (h Client) GetPaymentsByCurrency(ctx context.Context, sourceCurrency string, transactionType string, isRecurringPayment bool, destination string, country string, subdivision string) (paymentByCurrency models.PaymentResponse, err error) {

	apiURL := fmt.Sprintf(
		"%s/supported/payment-types/%s?type=%s&destination=%s&isRecurringPayment=%t",
		h.BaseURL,
		sourceCurrency,
		transactionType,
		destination,
		isRecurringPayment,
	)
	if country != "" {
		apiURL += "&country=" + country
	}
	if subdivision != "" {
		apiURL += "&subdivision=" + subdivision
	}

	h.Logger.Info("Fetching payment types by currency", zap.String("url", apiURL))

	// Prepare Onramper API request
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
	if err != nil {
		h.Logger.Error("Failed to create request", zap.Error(err))
		return paymentByCurrency, err
	}

	// Always add your API key
	req.Header.Add("Authorization", h.APIKey)
	if transactionType == transactionTypeBuy {
		req.Header.Add("X-Is-Recurringpayment", strconv.FormatBool(isRecurringPayment))
	}

	// Perform the request
	resp, err := h.HTTPClient.Do(req)
	if err != nil {
		h.Logger.Error("Failed to fetch payment types by currency", zap.Error(err))
		return paymentByCurrency, err
	}
	defer resp.Body.Close()

	h.Logger.Info("Received response for payment types by currency", zap.Int("status", resp.StatusCode))

	// Handle non-OK status codes
	var body []byte
	if resp.StatusCode != http.StatusOK {
		body, err = io.ReadAll(resp.Body)
		if err != nil {
			h.Logger.Error("Failed to read response body", zap.Error(err))
			return paymentByCurrency, err
		}
		err = fmt.Errorf(
			"unable to get payment types with status code: %d - message: %s",
			resp.StatusCode, string(body),
		)
		return paymentByCurrency, err
	}
	err = json.NewDecoder(resp.Body).Decode(&paymentByCurrency)
	if err != nil {
		h.Logger.Error("Failed to decode payment types response", zap.Error(err))
		err = fmt.Errorf("failed to decode response: %w", err)
		return paymentByCurrency, err
	}
	return paymentByCurrency, err
}
func (h Client) GetDefaults(ctx context.Context, transactionType string, country string, subdivision string) (defaults models.DefaultsResponse, err error) {
	// Construct API request URL with query parameters
	apiURL := fmt.Sprintf("%s/supported/defaults/all?type=%s", h.BaseURL, transactionType)
	if country != "" {
		apiURL += "&country=" + country
	}
	if subdivision != "" {
		apiURL += "&subdivision=" + subdivision
	}
	h.Logger.Info("Fetching currencies", zap.String("url", apiURL))

	// Prepare Onramper API request
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
	if err != nil {
		h.Logger.Error("Failed to create request", zap.Error(err))
		return defaults, err
	}

	req.Header.Add("Authorization", h.APIKey)

	resp, err := h.HTTPClient.Do(req)
	if err != nil {
		h.Logger.Error("Failed to fetch currencies", zap.Error(err))
		return defaults, err
	}
	defer resp.Body.Close()
	h.Logger.Info("Received response", zap.Int("status", resp.StatusCode))

	var body []byte
	if resp.StatusCode != http.StatusOK {
		body, err = io.ReadAll(resp.Body)
		if err != nil {
			h.Logger.Error("Failed to read response body", zap.Error(err))
			return defaults, err
		}
		err = fmt.Errorf("unable to get currencies with status code: %d - message: %s", resp.StatusCode, string(body))
		return defaults, err
	}
	err = json.NewDecoder(resp.Body).Decode(&defaults)
	if err != nil {
		h.Logger.Error("Failed to decode response", zap.Error(err))
		err = fmt.Errorf("failed to decode response: %w", err)
		return defaults, err
	}
	return defaults, err
}
func (h Client) GetAssets(ctx context.Context, paymentParam *models.AssetRequest) (assets models.SupportedAssetsResponse, err error) {
	// Build base URL (using params fields)
	params := url.Values{}
	if paymentParam.Source != "" {
		params.Add("source", paymentParam.Source)
	}
	if paymentParam.Country != "" {
		params.Add("country", paymentParam.Country)
	}
	if paymentParam.Type != "" {
		params.Add("type", string(paymentParam.Type))
	}
	if paymentParam.PaymentMethods != "" {
		params.Add("paymentMethods", paymentParam.PaymentMethods)
	}
	if paymentParam.Onramps != "" {
		params.Add("onramps", paymentParam.Onramps)
	}
	if paymentParam.Subdivision != "" {
		params.Add("subdivision", paymentParam.Subdivision)
	}

	apiURL := fmt.Sprintf("%s/supported/assets?%s", h.BaseURL, params.Encode())

	h.Logger.Info("Fetching supported assets", zap.String("url", apiURL))

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
	if err != nil {
		h.Logger.Error("Failed to create request", zap.Error(err))
		return assets, err
	}

	req.Header.Add("Authorization", h.APIKey)

	resp, err := h.HTTPClient.Do(req)
	if err != nil {
		h.Logger.Error("Failed to fetch supported assets", zap.Error(err))
		return assets, err
	}
	defer resp.Body.Close()

	h.Logger.Info("Received response", zap.Int("status", resp.StatusCode))

	// Handle non-OK status
	var body []byte
	if resp.StatusCode != http.StatusOK {
		body, err = io.ReadAll(resp.Body)
		if err != nil {
			h.Logger.Error("Failed to read response body", zap.Error(err))
			return assets, err
		}
		err = fmt.Errorf("unable to get assets with status code: %d - message: %s", resp.StatusCode, string(body))
		return assets, err
	}
	err = json.NewDecoder(resp.Body).Decode(&assets)
	if err != nil {
		h.Logger.Error("Failed to decode response", zap.Error(err))
		err = fmt.Errorf("failed to decode response: %w", err)
		return assets, err
	}
	return assets, err
}
func (h Client) GetOnramps(ctx context.Context, params *models.OnrampsQuery) (onramps models.OnrampResponse, err error) {
	// Build the API URL from the parameters.
	queryParams := url.Values{}
	queryParams.Add("type", params.TransactionType)
	queryParams.Add("source", params.Source)
	queryParams.Add("destination", params.Destination)

	if params.Country != "" {
		queryParams.Add("country", params.Country)
	}
	if params.Subdivision != "" {
		queryParams.Add("subdivision", params.Subdivision)
	}
	apiURL := fmt.Sprintf("%s/supported/onramps?%s", h.BaseURL, queryParams.Encode())

	// Logging for debug
	h.Logger.Info("Fetching supported onramps", zap.String("url", apiURL))

	// Create the request
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
	if err != nil {
		h.Logger.Error("Failed to create request", zap.Error(err))
		return onramps, err
	}
	// Add authorization header
	req.Header.Add("Authorization", h.APIKey)

	// Execute request
	resp, err := h.HTTPClient.Do(req)
	if err != nil {
		h.Logger.Error("Failed to fetch supported onramps", zap.Error(err))
		return onramps, err
	}
	defer resp.Body.Close()

	// Log status code for debugging
	h.Logger.Info("Received response", zap.Int("status", resp.StatusCode))

	// Handle non-200 status codes

	var body []byte
	if resp.StatusCode != http.StatusOK {
		body, err = io.ReadAll(resp.Body)
		if err != nil {
			h.Logger.Error("Failed to read response body", zap.Error(err))
			return onramps, err
		}
		err = fmt.Errorf("unable to get currencies with status code: %d - message: %s", resp.StatusCode, string(body))
		return onramps, err
	}
	// Decode successful response
	err = json.NewDecoder(resp.Body).Decode(&onramps)
	if err != nil {
		h.Logger.Error("Failed to decode onramps response", zap.Error(err))
		err = fmt.Errorf("failed to decode response: %w", err)
		return onramps, err
	}
	return onramps, err
}
func (h Client) GetOnrampMetadata(ctx context.Context, transactionType string) (metadata models.OnrampMetadataResponse, err error) {
	// Construct API request URL with query parameters
	apiURL := fmt.Sprintf("%s/supported/onramps/all?type=%s", h.BaseURL, transactionType)
	h.Logger.Info("Fetching onramp metadata", zap.String("url", apiURL))

	// Prepare Onramper API request
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
	if err != nil {
		h.Logger.Error("Failed to create request", zap.Error(err))
		return metadata, err
	}

	// Add your API key (if required)
	req.Header.Add("Authorization", h.APIKey)

	// Execute the request
	resp, err := h.HTTPClient.Do(req)
	if err != nil {
		h.Logger.Error("Failed to fetch onramp metadata", zap.Error(err))
		return metadata, err
	}
	defer resp.Body.Close()

	h.Logger.Info("Received response", zap.Int("status", resp.StatusCode))

	// Handle non-200 status codes
	var body []byte
	if resp.StatusCode != http.StatusOK {
		body, err = io.ReadAll(resp.Body)
		if err != nil {
			h.Logger.Error("Failed to read response body", zap.Error(err))
			return metadata, err
		}
		err = fmt.Errorf("unable to get onramp metadata with status code: %d - message: %s",
			resp.StatusCode, string(body),
		)
		return metadata, err
	}
	err = json.NewDecoder(resp.Body).Decode(&metadata)
	if err != nil {
		h.Logger.Error("Failed to decode response", zap.Error(err))
		err = fmt.Errorf("failed to decode response: %s", err.Error())
		return metadata, err
	}
	return metadata, err
}
func (h Client) GetCryptoByFiat(ctx context.Context, source string, country string) (cryptofiat models.CryptoFiatResponse, err error) {

	apiURL := fmt.Sprintf("%s/supported/crypto?source=%s", h.BaseURL, source)
	if country != "" {
		apiURL += "&country=" + country
	}

	// Log for debugging
	h.Logger.Info("Fetching crypto by fiat", zap.String("url", apiURL))

	// Create the HTTP request
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
	if err != nil {
		h.Logger.Error("Failed to create request", zap.Error(err))
		return cryptofiat, err
	}
	req.Header.Add("Authorization", h.APIKey)

	// Make the request
	resp, err := h.HTTPClient.Do(req)
	if err != nil {
		h.Logger.Error("Failed to fetch crypto by fiat", zap.Error(err))
		return cryptofiat, err
	}
	defer resp.Body.Close()

	// Log the status code
	h.Logger.Info("Received response", zap.Int("status", resp.StatusCode))

	// Handle non-200 status codes
	var body []byte
	if resp.StatusCode != http.StatusOK {
		body, err = io.ReadAll(resp.Body)
		if err != nil {
			h.Logger.Error("Failed to read response body", zap.Error(err))
			return cryptofiat, err
		}
		err = fmt.Errorf("unable to get crypto by fiat with status code: %d - message: %s", resp.StatusCode, string(body))
		return cryptofiat, err
	}
	// Decode the JSON into our CryptoFiatResponse model
	err = json.NewDecoder(resp.Body).Decode(&cryptofiat)
	if err != nil {
		h.Logger.Error("Failed to decode response", zap.Error(err))
		err = fmt.Errorf("failed to decode response: %w", err)
		return cryptofiat, err
	}
	return cryptofiat, err
}

// buildGetQuotesURL constructs the URL for the GetQuotes API call.
func (h Client) buildGetQuotesURL(fiat, crypto string, quotesParam *models.QuoteQueryParams) string {
	q := url.Values{}
	if quotesParam.Amount > 0 {
		q.Set("amount", strconv.FormatFloat(quotesParam.Amount, 'f', -1, 64))
	}
	if quotesParam.PaymentMethod != "" {
		q.Set("paymentMethod", quotesParam.PaymentMethod)
	}
	if quotesParam.UUID != "" {
		q.Set("uuid", quotesParam.UUID)
	}
	if quotesParam.ClientName != "" {
		q.Set("clientName", quotesParam.ClientName)
	}
	if quotesParam.Type != "" {
		q.Set("type", quotesParam.Type)
	}
	if quotesParam.WalletAddress != "" {
		q.Set("walletAddress", quotesParam.WalletAddress)
	}
	if quotesParam.IsRecurringPayment {
		q.Set("isRecurringPayment", "true")
	}
	if quotesParam.Input != "" {
		q.Set("input", quotesParam.Input)
	}
	if quotesParam.Country != "" {
		q.Set("country", quotesParam.Country)
	}
	if quotesParam.TxInitiation {
		q.Set("txInitiation", "true")
	}

	// Determine path based on transaction type
	var path string
	if quotesParam.Type == transactionTypeBuy {
		path = fmt.Sprintf("/quotes/%s/%s", fiat, crypto)
	} else {
		path = fmt.Sprintf("/quotes/%s/%s", crypto, fiat)
	}

	apiURL := h.BaseURL + path
	if len(q) > 0 {
		apiURL += "?" + q.Encode()
	}
	return apiURL
}

func (h Client) GetQuotes(ctx context.Context, fiat string, crypto string, quotesParam *models.QuoteQueryParams) (quotes []models.QuoteResponse, err error) {
	if fiat == "" || crypto == "" {
		err = errors.New("both fiat and crypto parameters are required")
		return quotes, err
	}

	apiURL := h.buildGetQuotesURL(fiat, crypto, quotesParam)

	q := url.Values{}
	if len(q) > 0 {
		apiURL += "?" + q.Encode()
	}

	h.Logger.Info("Fetching quotes", zap.String("url", apiURL))

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
	if err != nil {
		h.Logger.Error("Failed to create request", zap.Error(err))
		return quotes, err
	}
	req.Header.Add("Authorization", h.APIKey)
	req.Header.Set("Accept", "application/json")

	resp, err := h.HTTPClient.Do(req)
	if err != nil {
		h.Logger.Error("Failed to fetch quotes", zap.Error(err))
		return quotes, err
	}
	defer resp.Body.Close()

	h.Logger.Info("Received response", zap.Int("status", resp.StatusCode))

	var body []byte
	if resp.StatusCode != http.StatusOK {
		body, err = io.ReadAll(resp.Body)
		if err != nil {
			h.Logger.Error("Failed to read response body", zap.Error(err))
			return quotes, err
		}
		err = fmt.Errorf("unable to get quotes: %d - %s", resp.StatusCode, string(body))
		return quotes, err
	}
	err = json.NewDecoder(resp.Body).Decode(&quotes)
	if err != nil {
		h.Logger.Error("Failed to decode quotes", zap.Error(err))
		err = fmt.Errorf("failed to decode quotes: %w", err)
		return quotes, err
	}

	if len(quotes) == 0 {
		h.Logger.Error("Onramper returned empty quotes")
		err = errors.New("no quotes found")
		return quotes, err
	}

	h.Logger.Info("Quotes response",
		zap.Int("quote_count", len(quotes)))

	return quotes, err
}
func (h Client) GetTransactionByID(ctx context.Context, transactionID string) (transactionid models.TransactionResponse, err error) {
	apiURL := fmt.Sprintf("%s/transactions/%s", h.BaseURL, transactionID)

	h.Logger.Info("Fetching transaction details", zap.String("url", apiURL))

	// Prepare request
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
	if err != nil {
		h.Logger.Error("Failed to create request", zap.Error(err))
		return transactionid, err
	}

	// Use the webhook secret as the Authorization header
	req.Header.Set("Authorization", h.APIKey)
	req.Header.Set("X-Onramper-Secret", h.WebhookSecret)

	// Make the request
	resp, err := h.HTTPClient.Do(req)
	if err != nil {
		h.Logger.Error("Failed to fetch transaction", zap.Error(err))
		return transactionid, err
	}
	defer resp.Body.Close()

	h.Logger.Info("Received response", zap.Int("status", resp.StatusCode))

	// Handle error responses
	var body []byte
	if resp.StatusCode != http.StatusOK {
		body, err = io.ReadAll(resp.Body)
		if err != nil {
			h.Logger.Error("Failed to read response body", zap.Error(err))
			return transactionid, err
		}
		err = fmt.Errorf("unable to get transaction: %d - %s", resp.StatusCode, string(body))
		return transactionid, err
	}
	// var response models.TransactionResponse // This was unused, transactionid is the return variable
	// Decode into struct
	err = json.NewDecoder(resp.Body).Decode(&transactionid)
	if err != nil {
		h.Logger.Error("Failed to decode transaction response", zap.Error(err))
		err = fmt.Errorf("failed to decode transaction response: %w", err)
		return transactionid, err
	}

	h.Logger.Info("Transaction fetched",
		zap.String("transaction_id", transactionid.TransactionID))

	return transactionid, err
}
func (h Client) ListTransactions(ctx context.Context, query models.TransactionListQuery) (transactionlist models.TransactionListResponse, err error) {
	apiURL := fmt.Sprintf("%s/transactions", h.BaseURL)
	// Construct query string
	params := url.Values{}
	if query.StartDateTime != "" {
		params.Add("startDateTime", query.StartDateTime)
	}
	if query.EndDateTime != "" {
		params.Add("endDateTime", query.EndDateTime)
	}
	if query.Limit > 0 {
		params.Add("limit", strconv.Itoa(query.Limit))
	}
	if query.TransactionIDs != "" {
		params.Add("transactionIds", query.TransactionIDs)
	}
	if query.Cursor != "" {
		params.Add("cursor", query.Cursor)
	}

	fullURL := apiURL
	if encoded := params.Encode(); encoded != "" {
		fullURL += "?" + encoded
	}

	h.Logger.Info("Fetching transaction list", zap.String("url", fullURL))

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fullURL, nil)
	if err != nil {
		h.Logger.Error("Failed to create request", zap.Error(err))
		return transactionlist, err
	}
	// Set headers
	req.Header.Set("Authorization", "Bearer "+h.APIKey)
	req.Header.Set("X-Onramper-Secret", h.WebhookSecret)

	// Execute request
	resp, err := h.HTTPClient.Do(req)
	if err != nil {
		h.Logger.Error("Failed to perform request", zap.Error(err))
		return transactionlist, err
	}
	defer resp.Body.Close()

	h.Logger.Info("Received response", zap.Int("status", resp.StatusCode))

	// Handle non-200
	if resp.StatusCode != http.StatusOK {
		var body []byte
		body, err = io.ReadAll(resp.Body)
		if err != nil {
			h.Logger.Error("Failed to read error response body for ListTransactions", zap.Error(err))
			// err is already set from ReadAll, so we can return it
			return transactionlist, fmt.Errorf("unable to list transactions: %d and failed to read body: %w", resp.StatusCode, err)
		}
		err = fmt.Errorf("unable to list transactions: %d - %s", resp.StatusCode, string(body))
		return transactionlist, err
	}

	err = json.NewDecoder(resp.Body).Decode(&transactionlist)
	if err != nil {
		h.Logger.Error("Failed to decode response", zap.Error(err))
		err = fmt.Errorf("failed to decode transaction list: %w", err)
		return transactionlist, err
	}

	h.Logger.Info("Transactions fetched",
		zap.Int("Transaction_List", len(transactionlist.Transactions)),
	)
	return transactionlist, err
}
func (h Client) InitiateTransaction(ctx context.Context, payload models.InitiateTransactionRequest) (transaction models.InitiateTransactionResponse, err error) {
	// Construct API request URL
	apiURL := fmt.Sprintf("%s/checkout/intent", h.BaseURL)
	h.Logger.Info("Initiating transaction", zap.String("url", apiURL))

	// Marshal payload to JSON
	requestBody, err := json.Marshal(payload)
	if err != nil {
		h.Logger.Error("Failed to marshal request body", zap.Error(err))
		return transaction, err
	}

	// Prepare Onramper API request
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, apiURL, bytes.NewBuffer(requestBody))
	if err != nil {
		h.Logger.Error("Failed to create request", zap.Error(err))
		return transaction, err
	}
	fmt.Println("resquest body: ", payload)
	req.Header.Add("Authorization", h.APIKey)
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")

	resp, err := h.HTTPClient.Do(req)
	if err != nil {
		h.Logger.Error("Failed to initiate transaction", zap.Error(err))
		return transaction, err
	}
	defer resp.Body.Close()
	h.Logger.Info("Received response", zap.Int("status", resp.StatusCode))

	var body []byte
	if resp.StatusCode != http.StatusOK {
		body, err = io.ReadAll(resp.Body)
		if err != nil {
			h.Logger.Error("Failed to read response body", zap.Error(err))
			return transaction, err
		}
		err = fmt.Errorf("unable to initiate transaction with status code: %d - message: %s", resp.StatusCode, string(body))
		return transaction, err
	}
	// Parse the JSON response body into the model struct
	err = json.NewDecoder(resp.Body).Decode(&transaction)
	if err != nil {
		h.Logger.Error("Failed to decode response", zap.Error(err))
		err = fmt.Errorf("failed to decode response: %s", err.Error())
		return transaction, err
	}
	// Optional: Validate the response if needed
	if transaction.Message.TransactionInformation.TransactionID == "" {
		h.Logger.Error("Onramper API returned an invalid transaction response")
		err = errors.New("onramper API returned an empty transaction ID")
		return transaction, err
	}
	h.Logger.Info("Transaction response",
		zap.String("transaction_id", transaction.Message.TransactionInformation.TransactionID),
		zap.String("url", transaction.Message.TransactionInformation.URL),
	)
	return transaction, err
}
func (h Client) ConfirmSellTransaction(ctx context.Context, txType string) (confirmation models.SellTransactionConfirmationResponse, err error) {
	// Construct API request URL
	apiURL := fmt.Sprintf("%s/transactions/confirm/%s", h.BaseURL, txType)
	h.Logger.Info("Confirming sell transaction", zap.String("url", apiURL))
	// Prepare Onramper API request
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, apiURL, nil)
	if err != nil {
		h.Logger.Error("Failed to create confirmation request", zap.Error(err))
		return confirmation, err
	}
	req.Header.Add("Authorization", "Bearer "+h.APIKey)
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "application/json")

	resp, err := h.HTTPClient.Do(req)
	if err != nil {
		h.Logger.Error("Failed to send confirmation request", zap.Error(err))
		return confirmation, err
	}
	defer resp.Body.Close()
	h.Logger.Info("Received response", zap.Int("status", resp.StatusCode))

	var body []byte
	if resp.StatusCode != http.StatusOK {
		body, err = io.ReadAll(resp.Body)
		if err != nil {
			h.Logger.Error("Failed to read error response body", zap.Error(err))
			return confirmation, err
		}
		err = fmt.Errorf("failed to confirm sell transaction with status code: %d - message: %s", resp.StatusCode, string(body))
		return confirmation, err
	}
	// Decode the response bod
	err = json.NewDecoder(resp.Body).Decode(&confirmation)
	if err != nil {
		h.Logger.Error("Failed to decode confirmation response", zap.Error(err))
		err = fmt.Errorf("failed to decode confirmation response: %s", err.Error())
		return confirmation, err
	}

	if confirmation.Status == "" {
		h.Logger.Error("Empty status in confirmation response")
		err = errors.New("empty confirmation status returned")
		return confirmation, err
	}

	h.Logger.Info("Sell transaction confirmed",
		zap.String("status", confirmation.Status),
	)
	return confirmation, err
}
