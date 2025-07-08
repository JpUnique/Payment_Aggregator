package onramper

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/subdialia/fiat-ramp-service/pkg/models"
	"go.uber.org/zap"
)

// Minimal mock implementing ONLY needed methods.
type MockOnramperClient struct {
	mock.Mock
}

func TestGetCurrencies(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockResponse := json.RawMessage(`{"message":{"crypto":[],"fiat":[]}}`)
	t.Run("success", func(t *testing.T) {
		mockClient := new(MockOnramperClient)
		mockClient.On("GetCurrencies", context.Background(), "US", "NY", "buy").
			Return(mockResponse, nil)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/supported?type=buy&country=US&subdivision=NY", nil)
		assert.Equal(t, http.StatusOK, w.Code)
	})
}
func TestGetPaymentTypes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockResponse := json.RawMessage(`{"paymentMethods":["credit_card"]}`)
	t.Run("success", func(t *testing.T) {
		mockClient := new(MockOnramperClient)
		mockClient.On("GetPaymentTypes", context.Background(), "buy", true, "US").
			Return(mockResponse, nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/supported/payment-types?type=buy&country=US&isRecurringPayment=true", nil)
		assert.Equal(t, http.StatusOK, w.Code)
	})
}
func TestGetPaymentsByCurrency(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockResponse := json.RawMessage(`{"message":{"payments":["credit_card"]}}`)

	t.Run("success", func(t *testing.T) {
		mockClient := new(MockOnramperClient)
		mockClient.On("GetPaymentsByCurrency",
			context.Background(),
			"USD", // sourceCurrency
			"buy", // transactionType
			false, // isRecurring
			"BTC", // destination
			"US",  // country
			"NY",  // subdivision
		).Return(mockResponse, nil)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/payment-types/USD?type=buy&destination=BTC&country=US&subdivision=NY", nil)
		c.Params = gin.Params{{Key: "source", Value: "USD"}}

		assert.Equal(t, http.StatusOK, w.Code)
	})
}
func TestGetDefaults(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockResponse := json.RawMessage(`{"defaults":{"currency":"USD"}}`)
	t.Run("success", func(t *testing.T) {
		mockClient := new(MockOnramperClient)
		mockClient.On("GetDefaults",
			context.Background(),
			"sell", // transactionType
			"CA",   // country
			"ON",   // subdivision
		).Return(mockResponse, nil)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/supported/defaults?type=sell&country=CA&subdivision=ON", nil)
		assert.Equal(t, http.StatusOK, w.Code)
	})
}
func TestConfirmSellTransaction(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockResponse := json.RawMessage(`{"status":"confirmed"}`)

	t.Run("success", func(t *testing.T) {
		mockClient := new(MockOnramperClient)
		mockClient.On("ConfirmSellTransaction", context.Background(), "sell").
			Return(mockResponse, nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{{Key: "type", Value: "sell"}}
		c.Request = httptest.NewRequest(http.MethodPost, "/transction/confirm/offramp", nil)
		assert.Equal(t, http.StatusOK, w.Code)
	})
}
func TestListTransactions(t *testing.T) {
	gin.SetMode(gin.TestMode)
	statusDate := time.Date(2023, 3, 3, 9, 5, 3, 806000000, time.UTC)

	mockResponse := &models.TransactionListResponse{
		Transactions: []models.TransactionItem{
			{
				TargetCurrency:        "eth",
				Onramp:                "moonpay",
				StatusDate:            statusDate,
				TxType:                "onramp",
				TxHash:                "0x99c1b81f682b77023a68165fce4d95bee29fcbc03208dd6788c880d4709e7aa5",
				Status:                "completed",
				APIKey:                "pk_prod_01GTC8JT9MDSW8G11HPPKSVBTJ",
				SourceCurrency:        "eur",
				Country:               "nl",
				InAmount:              100,
				OutAmount:             0.0644,
				TxID:                  "01GTKAZ20PCES058TDY7WJY2PZ",
				ExternalTransactionID: "b305dc1c-2784-4cd6-9ceb-541fab881378",
				SK:                    "2023-03-03T09:05:03.806Z",
				Wallet:                "0xf04f6e033ac995007b0ab7cc570d130aee4b7c52",
				PaymentMethod:         "credit_debit_card",
			},
		},
		Limit: 12,
	}
	t.Run("success", func(t *testing.T) {
		mockClient := new(MockOnramperClient)
		expectedQuery := models.TransactionListQuery{
			Limit:  12,
			Cursor: "",
		}
		mockClient.On("ListTransactions", mock.Anything, expectedQuery).Return(mockResponse, nil)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/transactions?limit=12", nil)
		assert.Equal(t, http.StatusOK, w.Code)
	})
}
func TestGetCryptoByFiat(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockResponse := json.RawMessage(`{"crypto":["BTC","ETH"]}`)

	t.Run("success", func(t *testing.T) {
		mockClient := new(MockOnramperClient)
		mockClient.On("GetCryptoByFiat", context.Background(), "USD", "US").
			Return(mockResponse, nil)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/crypto?source=USD&country=US", nil)
		assert.Equal(t, http.StatusOK, w.Code)
	})
	t.Run("missing source", func(t *testing.T) {
		manager := &OnramperManager{Logger: zap.NewNop()}

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/crypto", nil)

		manager.GetCryptoByFiat(c)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}
func TestGetQuotesBuy(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockResponse := json.RawMessage(`{"quotes":[{"rate":1.2}]}`)
	t.Run("success", func(t *testing.T) {
		mockClient := new(MockOnramperClient)
		mockClient.On("GetQuotes",
			context.Background(),
			"USD", // fiat
			"BTC", // crypto
			&models.QuoteQueryParams{
				Amount:             10000.50,
				PaymentMethod:      "credit_card",
				UUID:               "user-12345",
				ClientName:         "johnpaul",
				Type:               "buy",
				WalletAddress:      "bc1qxy2kgdygjrsqtzq2n0yrf2493p83kkfjhx0wlh",
				IsRecurringPayment: true,
				Input:              "destination",
				Country:            "US",
				TxInitiation:       true,
			},
		).Return(mockResponse, nil)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/quotes/USD/BTC?amount=100&isRecurring=true&paymentMethod=credit_card", nil)
		c.Params = gin.Params{
			{Key: "source", Value: "USD"},
			{Key: "destination", Value: "BTC"},
		}
		assert.Equal(t, http.StatusOK, w.Code)
	})
	t.Run("missing params", func(t *testing.T) {
		manager := &OnramperManager{Logger: zap.NewNop()}

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/quotes/USD/BTC", nil)

		manager.GetQuotes(c)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}
func TestGetQuotesSell(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockResponse := json.RawMessage(`{"quotes":[{"rate":0.8}]}`)
	t.Run("success", func(t *testing.T) {
		mockClient := new(MockOnramperClient)
		mockClient.On("GetQuotes",
			context.Background(),
			"USD", // fiat
			"BTC", // crypto
			&models.QuoteQueryParams{
				Amount:             10,
				PaymentMethod:      "credit_card",
				UUID:               "user-12345",
				ClientName:         "johnpaul",
				Type:               "sell",
				WalletAddress:      "bc1qxy2kgdygjrsqtzq2n0yrf2493p83kkfjhx0wlh",
				IsRecurringPayment: true,
				Input:              "destination",
				Country:            "US",
				TxInitiation:       true,
			},
		).Return(mockResponse, nil)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/quotes/BTC/USD?amount=10&isRecurring=true&paymentMethod=credit_card", nil)
		c.Params = gin.Params{
			{Key: "source", Value: "BTC"},
			{Key: "destination", Value: "USD"},
		}
		assert.Equal(t, http.StatusOK, w.Code)
	})
	t.Run("missing params", func(t *testing.T) {
		manager := &OnramperManager{Logger: zap.NewNop()}

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/quotes/BTC/USD", nil)

		manager.GetQuotes(c)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}
func TestGetOnrampMetadata(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockResponse := json.RawMessage(`{"onramps":{"limits":5000}}`)
	t.Run("success with default type", func(t *testing.T) {
		mockClient := new(MockOnramperClient)
		mockClient.On("GetOnrampMetadata", context.Background(), "buy").
			Return(mockResponse, nil)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/supported/onramps/all?type=buy", nil)
		assert.Equal(t, http.StatusOK, w.Code)
	})
	t.Run("success with custom type", func(t *testing.T) {
		mockClient := new(MockOnramperClient)
		mockClient.On("GetOnrampMetadata", context.Background(), "sell").
			Return(mockResponse, nil)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/supported/onramps/all?type=sell", nil)
		assert.Equal(t, http.StatusOK, w.Code)
	})
}
func TestGetAssets(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockResponse := json.RawMessage(`{"assets":["BTC","USD"]}`)
	t.Run("success with full parameters", func(t *testing.T) {
		mockClient := new(MockOnramperClient)
		mockClient.On("GetAssets", mock.Anything, &models.AssetRequest{
			Type:           "sell",
			Source:         "USD",
			Country:        "US",
			Subdivision:    "NY",
			Onramps:        "moonpay,transak",
			PaymentMethods: "credit_card",
		}).Return(mockResponse, nil)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet,
			"/assets?type=sell&source=USD&country=US&subdivision=NY&onramps=moonpay,transak&paymentMethods=credit_card",
			nil)
		c.Request = httptest.NewRequest(http.MethodGet,
			"/assets?type=buy&source=USD&country=US&subdivision=NY&onramps=moonpay,transak&paymentMethods=credit_card",
			nil)
		assert.Equal(t, http.StatusOK, w.Code)
	})
	t.Run("success with minimal parameters", func(t *testing.T) {
		mockClient := new(MockOnramperClient)
		mockClient.On("GetAssets", mock.Anything, &models.AssetRequest{
			Type: "buy", // Default value
		}).Return(mockResponse, nil)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/assets?type=buy&source=USD", nil)
		assert.Equal(t, http.StatusOK, w.Code)
	})
}
func TestGetTransactionByID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockResponse := json.RawMessage(`{
  "country": "us",
  "inAmount": "68",
  "onramp": "transfi",
  "onrampTransactionId": "OR-2123428075629314",
  "outAmount": "0.00202087",
  "paymentMethod": "creditcard",
  "sourceCurrency": "usd",
  "status": "completed",
  "statusDate": "2023-07-28T07:56:42.012Z",
  "targetCurrency": "btc",
  "transactionId": "01H6DQWMRC8FA9MBM0HS5NABCD",
  "transactionType": "onramp",
  "transactionHash": "ef76220d3cfd028a7f324ce8744b7a6AWSFKp62f8f94c4dae5149bb41afd7e279",
  "walletAddress": "bc1qp56l3l2w2vdle8cfABCDEFlnlgc7ye7q0lenu3"
}`)
	t.Run("success", func(t *testing.T) {
		mockClient := new(MockOnramperClient)
		mockClient.On("GetTransactionByID", context.Background(), "01H6DQWMRC8FA9MBM0TESTJ72").
			Return(mockResponse, nil)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/transactions/01H6DQWMRC8FA9MBM0TESTJ72", nil)
		c.Params = gin.Params{{Key: "transaction_id", Value: "01H6DQWMRC8FA9MBM0TESTJ72"}}
		assert.Equal(t, http.StatusOK, w.Code)
	})
	t.Run("missing transaction ID", func(t *testing.T) {
		manager := &OnramperManager{Logger: zap.NewNop()}

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/transactions/", nil)

		manager.GetTransactionByID(c)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
	t.Run("client error", func(t *testing.T) {
		mockClient := new(MockOnramperClient)
		mockClient.On("GetTransactionByID", context.Background(), "tx_12345").
			Return(nil, errors.New("api error"))
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/transactions/tx_12345", nil)
		c.Params = gin.Params{{Key: "transaction_id", Value: "tx_12345"}}
	})
}
func TestGetOnramps(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockResponse := json.RawMessage(`{"onramps":["moonpay"]}`)

	t.Run("success with valid parameters", func(t *testing.T) {
		mockClient := new(MockOnramperClient)
		mockClient.On("GetOnramps", mock.Anything, &models.OnrampsQuery{
			TransactionType: "buy",
			Source:          "USD",
			Destination:     "BTC",
		}).Return(mockResponse, nil)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/onramps?type=buy&source=USD&destination=BTC", nil)
		assert.Equal(t, http.StatusOK, w.Code)
	})
	t.Run("invalid parameters", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/onramps?type=invalid", nil)
	})
	t.Run("client error", func(t *testing.T) {
		mockClient := new(MockOnramperClient)
		mockClient.On("GetOnramps", mock.Anything, mock.Anything).
			Return(nil, errors.New("api error"))

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/onramps?type=sell", nil)
	})
}
func TestInitiateTransaction(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockResponse := struct {
		Message struct {
			Status                 string `json:"status"`
			TransactionInformation struct {
				TransactionID string `json:"transactionId"`
				URL           string `json:"url"`
			} `json:"transactionInformation"`
		} `json:"message"`
	}{
		Message: struct {
			Status                 string `json:"status"`
			TransactionInformation struct {
				TransactionID string `json:"transactionId"`
				URL           string `json:"url"`
			} `json:"transactionInformation"`
		}{
			Status: "in_progress",
			TransactionInformation: struct {
				TransactionID string `json:"transactionId"`
				URL           string `json:"url"`
			}{
				TransactionID: "01H9KBT5C21JY0BAX4VTW9EP3V",
				URL:           "https://buy.moonpay.com/...",
			},
		},
	}
	t.Run("success", func(t *testing.T) {
		mockClient := new(MockOnramperClient)

		mockClient.On("InitiateTransaction", mock.Anything, mock.Anything).Return(mockResponse, nil)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/?user_id=user_456", bytes.NewBufferString(`{"wallet":{"address":"0x123"}}`))
		c.Request.Header.Set("Content-Type", "application/json")
		assert.Equal(t, http.StatusOK, w.Code)
	})
}
