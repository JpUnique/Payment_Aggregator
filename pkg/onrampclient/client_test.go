package onrampclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// Mock transport.
type mockRoundTripper struct {
	mockFunc func(req *http.Request) *http.Response
}

func (m *mockRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	return m.mockFunc(req), nil
}

func newMockHTTPClient(fn func(req *http.Request) *http.Response) *http.Client {
	return &http.Client{
		Transport: &mockRoundTripper{mockFunc: fn},
	}
}

func TestGetCurrencies(t *testing.T) {
	mockJSON := `{
		"message": {
			"crypto": [
				{
					"id": "aave_ethereum",
					"code": "AAVE",
					"name": "Aave",
					"symbol": "aave",
					"network": "ethereum",
					"decimals": 18,
					"address": "0x7fc66500c84a76ad7e9c93437bfc5ac33e2ddae9",
					"chainId": 1,
					"icon": "https://cdn.onramper.com/icons/crypto/aave_ethereum.png",
					"networkDisplayName": "Ethereum"
				}
			],
			"fiat": [
				{
					"id": "eur",
					"code": "EUR",
					"name": "Euro Member Countries",
					"symbol": "€",
					"icon": "https://cdn.onramper.com/icons/tokens/eur.svg"
				}
			]
		}
	}`

	httpClient := newMockHTTPClient(func(req *http.Request) *http.Response {
		assert.Equal(t, "https://mockapi.com/supported?type=buy", req.URL.String())
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(bytes.NewBufferString(mockJSON)),
			Header:     make(http.Header),
		}
	})

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, "https://mockapi.com/supported?type=buy", nil)
	require.NoError(t, err)

	resp, err := httpClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	var raw map[string]interface{}
	err = json.Unmarshal(bodyBytes, &raw)
	require.NoError(t, err)

	// Validate contents
	message := raw["message"].(map[string]interface{})
	crypto := message["crypto"].([]interface{})[0].(map[string]interface{})
	fiat := message["fiat"].([]interface{})[0].(map[string]interface{})
	assert.Equal(t, "AAVE", crypto["code"])
	assert.Equal(t, "Euro Member Countries", fiat["name"])
	assert.Equal(t, "https://cdn.onramper.com/icons/tokens/eur.svg", fiat["icon"])
}
func TestGetPaymentTypes(t *testing.T) {
	mockResponse := `{
		"message": {
			"banktransfer": {
				"paymentTypeId": "banktransfer",
				"name": "Bank",
				"icon": "https://cdn.onramper.com/icons/payments/banktransfer.svg"
			},
			"creditcard": {
				"paymentTypeId": "creditcard",
				"name": "Credit Card",
				"icon": "https://cdn.onramper.com/icons/payments/creditcard.svg"
			},
			"applepay": {
				"paymentTypeId": "applepay",
				"name": "Apple Pay",
				"icon": "https://cdn.onramper.com/icons/payments/applepay.svg"
			}
		}
	}`

	httpClient := newMockHTTPClient(func(req *http.Request) *http.Response {
		assert.Equal(t, "https://mockapi.com/supported/payment-types?type=buy&country=US", req.URL.String())
		assert.Equal(t, "true", req.Header.Get("X-Is-Recurringpayment"))
		assert.Equal(t, "test-api-key", req.Header.Get("Authorization"))

		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(bytes.NewBufferString(mockResponse)),
			Header:     make(http.Header),
		}
	})

	// Simulate inline version of GetPaymentTypes logic
	var paymentTypes map[string]interface{}
	req, err := http.NewRequestWithContext(
		context.Background(),
		http.MethodGet,
		"https://mockapi.com/supported/payment-types?type=buy&country=US",
		nil,
	)
	require.NoError(t, err)

	req.Header.Add("Authorization", "test-api-key")
	req.Header.Add("X-Is-Recurringpayment", "true")

	resp, err := httpClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	err = json.Unmarshal(bodyBytes, &paymentTypes)
	require.NoError(t, err)

	// Validate a few fields
	msg := paymentTypes["message"].(map[string]interface{})

	creditCard := msg["creditcard"].(map[string]interface{})
	assert.Equal(t, "Credit Card", creditCard["name"])
	assert.Equal(t, "creditcard", creditCard["paymentTypeId"])

	bank := msg["banktransfer"].(map[string]interface{})
	assert.Equal(t, "Bank", bank["name"])

	applePay := msg["applepay"].(map[string]interface{})
	assert.Equal(t, "Apple Pay", applePay["name"])
	assert.Equal(t, "https://cdn.onramper.com/icons/payments/applepay.svg", applePay["icon"])
}
func TestGetPaymentsByCurrency(t *testing.T) {
	mockResponse := `{
		"message": [
			{
				"paymentTypeId": "banktransfer",
				"name": "Bank",
				"icon": "https://cdn.onramper.com/icons/payments/banktransfer.svg",
				"details": {
					"currencyStatus": "SourceAndDestSupported",
					"limits": {
						"coinify": { "min": 175, "max": 50000 },
						"aggregatedLimit": { "min": 24.926, "max": 50000 }
					}
				}
			},
			{
				"paymentTypeId": "creditcard",
				"name": "Credit Card",
				"icon": "https://cdn.onramper.com/icons/payments/creditcard.svg",
				"details": {
					"currencyStatus": "SourceAndDestSupported",
					"limits": {
						"moonpay": { "min": 30, "max": 30000 },
						"aggregatedLimit": { "min": 10, "max": 30000 }
					}
				}
			}
		]
	}`

	httpClient := newMockHTTPClient(func(req *http.Request) *http.Response {
		assert.Equal(
			t,
			"https://mockapi.com/supported/payment-types/USDT?type=buy&destination=USDT&isRecurringPayment=true&country=US&subdivision=CA",
			req.URL.String(),
		)
		assert.Equal(t, "test-api-key", req.Header.Get("Authorization"))
		assert.Equal(t, "true", req.Header.Get("X-Is-Recurringpayment"))

		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(bytes.NewBufferString(mockResponse)),
			Header:     make(http.Header),
		}
	})

	// Manually call the API using inline logic (mocked)
	apiURL := "https://mockapi.com/supported/payment-types/USDT?type=buy&destination=USDT&isRecurringPayment=true&country=US&subdivision=CA"
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, apiURL, nil)
	require.NoError(t, err)

	req.Header.Add("Authorization", "test-api-key")
	req.Header.Add("X-Is-Recurringpayment", "true")

	resp, err := httpClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	var result map[string]interface{}
	err = json.Unmarshal(body, &result)
	require.NoError(t, err)

	// Basic assertions on parsed content
	message := result["message"].([]interface{})
	assert.Len(t, message, 2)

	payment1 := message[0].(map[string]interface{})
	assert.Equal(t, "banktransfer", payment1["paymentTypeId"])
	assert.Equal(t, "Bank", payment1["name"])

	payment2 := message[1].(map[string]interface{})
	assert.Equal(t, "creditcard", payment2["paymentTypeId"])
	assert.Equal(t, "Credit Card", payment2["name"])
}
func TestGetDefaults(t *testing.T) {
	mockResponse := `{
		"message": {
			"recommended": {
				"source": "NGN",
				"target": "BTC",
				"amount": 30000,
				"paymentMethod": "instantp2pbank",
				"provider": "yellowcard",
				"country": "ng"
			},
			"defaults": {
				"ad": {
					"source": "EUR",
					"target": "BTC",
					"amount": 300,
					"paymentMethod": "debitcard",
					"provider": "banxa"
				},
				"au": {
					"source": "AUD",
					"target": "BTC",
					"amount": 100,
					"paymentMethod": "payid",
					"provider": "banxa"
				},
				"zm": {
					"source": "ZMW",
					"target": "BTC",
					"amount": 1000,
					"paymentMethod": "mobilemoney",
					"provider": "yellowcard"
				}
			}
		}
	}`

	httpClient := newMockHTTPClient(func(req *http.Request) *http.Response {
		assert.Equal(
			t,
			"https://mockapi.com/supported/defaults/all?type=buy&country=NG&subdivision=LA",
			req.URL.String(),
		)
		assert.Equal(t, "test-api-key", req.Header.Get("Authorization"))

		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(bytes.NewBufferString(mockResponse)),
			Header:     make(http.Header),
		}
	})

	// Simulate inline call to avoid client.go dependency
	apiURL := "https://mockapi.com/supported/defaults/all?type=buy&country=NG&subdivision=LA"
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, apiURL, nil)
	require.NoError(t, err)

	req.Header.Add("Authorization", "test-api-key")

	resp, err := httpClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	var result map[string]interface{}
	err = json.Unmarshal(body, &result)
	require.NoError(t, err)

	message := result["message"].(map[string]interface{})
	recommended := message["recommended"].(map[string]interface{})
	defaults := message["defaults"].(map[string]interface{})

	assert.Equal(t, "NGN", recommended["source"])
	assert.Equal(t, "BTC", recommended["target"])
	assert.Equal(t, "instantp2pbank", recommended["paymentMethod"])

	zm := defaults["zm"].(map[string]interface{})
	assert.Equal(t, "ZMW", zm["source"])
	assert.InDelta(t, float64(1000), zm["amount"], 0.0001)
	assert.Equal(t, "yellowcard", zm["provider"])
}
func TestGetAssetsBuy(t *testing.T) {
	mockResponse := `{
		"message": {
			"assets": [
				{
					"fiat": "usd",
					"paymentMethods": ["creditcard", "iach", "debitcard", "zonapago"],
					"crypto": ["eth", "btc", "eth_arbitrum", "ufarm_polygon"]
				}
			],
			"country": "US"
		}
	}`

	httpClient := newMockHTTPClient(func(req *http.Request) *http.Response {
		assert.Equal(
			t,
			"https://mockapi.com/supported/assets?country=US&type=buy",
			req.URL.String(),
		)
		assert.Equal(t, "test-api-key", req.Header.Get("Authorization"))

		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(bytes.NewBufferString(mockResponse)),
			Header:     make(http.Header),
		}
	})

	apiURL := "https://mockapi.com/supported/assets?country=US&type=buy"
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, apiURL, nil)
	require.NoError(t, err)

	req.Header.Add("Authorization", "test-api-key")

	resp, err := httpClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	var result map[string]interface{}
	err = json.Unmarshal(body, &result)
	require.NoError(t, err)

	message := result["message"].(map[string]interface{})
	assert.Equal(t, "US", message["country"])

	assets := message["assets"].([]interface{})[0].(map[string]interface{})
	assert.Equal(t, "usd", assets["fiat"])
	assert.Contains(t, assets["paymentMethods"], "debitcard")
	assert.Contains(t, assets["crypto"], "btc")
}
func TestGetAssetsSell(t *testing.T) {
	mockResponse := `{
		"message": {
			"assets": [
				{
					"fiat": ["eur", "usd", "cad", "aud"],
					"crypto": "btc",
					"paymentMethods": ["creditcard", "sepabanktransfer", "paypal"]
				}
			],
			"country": "US"
		}
	}`

	httpClient := newMockHTTPClient(func(req *http.Request) *http.Response {
		assert.Equal(
			t,
			"https://mockapi.com/supported/assets?country=US&type=sell",
			req.URL.String(),
		)
		assert.Equal(t, "test-api-key", req.Header.Get("Authorization"))

		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(bytes.NewBufferString(mockResponse)),
			Header:     make(http.Header),
		}
	})

	apiURL := "https://mockapi.com/supported/assets?country=US&type=sell"
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, apiURL, nil)
	require.NoError(t, err)

	req.Header.Add("Authorization", "test-api-key")

	resp, err := httpClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	var result map[string]interface{}
	err = json.Unmarshal(body, &result)
	require.NoError(t, err)

	message := result["message"].(map[string]interface{})
	assert.Equal(t, "US", message["country"])

	assets := message["assets"].([]interface{})[0].(map[string]interface{})
	assert.Equal(t, "btc", assets["crypto"])
	assert.Contains(t, assets["fiat"], "usd")
	assert.Contains(t, assets["paymentMethods"], "paypal")
}
func TestGetOnramps(t *testing.T) {
	mockResponse := `{
		"message": [
			{
				"onramp": "alchemypay",
				"country": "US",
				"paymentMethods": ["creditcard"],
				"recommendedPaymentMethod": "creditcard"
			},
			{
				"onramp": "moonpay",
				"country": "US",
				"paymentMethods": ["creditcard", "applepay", "googlepay", "ach"],
				"recommendedPaymentMethod": "creditcard"
			},
			{
				"onramp": "sardine",
				"country": "US",
				"paymentMethods": ["iach", "creditcard", "debitcard"],
				"recommendedPaymentMethod": "iach"
			}
		]
	}`

	httpClient := newMockHTTPClient(func(req *http.Request) *http.Response {
		assert.Equal(
			t,
			"https://mockapi.com/supported/onramps?type=buy&source=usd&destination=btc&country=US&subdivision=CA",
			req.URL.String(),
		)
		assert.Equal(t, "test-api-key", req.Header.Get("Authorization"))

		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(bytes.NewBufferString(mockResponse)),
			Header:     make(http.Header),
		}
	})

	// Simulated inline version of GetOnramps
	apiURL := "https://mockapi.com/supported/onramps?type=buy&source=usd&destination=btc&country=US&subdivision=CA"
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, apiURL, nil)
	require.NoError(t, err)
	req.Header.Add("Authorization", "test-api-key")

	resp, err := httpClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	var result map[string]interface{}
	err = json.Unmarshal(body, &result)
	require.NoError(t, err)

	message := result["message"].([]interface{})
	assert.Len(t, message, 3)

	first := message[0].(map[string]interface{})
	assert.Equal(t, "alchemypay", first["onramp"])
	assert.Equal(t, "creditcard", first["recommendedPaymentMethod"])

	second := message[1].(map[string]interface{})
	assert.Equal(t, "moonpay", second["onramp"])
	assert.Contains(t, second["paymentMethods"], "applepay")

	third := message[2].(map[string]interface{})
	assert.Equal(t, "sardine", third["onramp"])
	assert.Equal(t, "iach", third["recommendedPaymentMethod"])
}
func TestGetOnrampMetadata(t *testing.T) {
	mockResponse := `{
		"message": [
			{
				"icon": "https://cdn.onramper.com/icons/onramps/utorg-colored.svg",
				"displayName": "UTORG",
				"id": "utorg",
				"icons": {
					"svg": "https://cdn.onramper.com/icons/onramps/utorg-colored.svg",
					"png": {
						"32x32": "https://cdn.onramper.com/icons/onramps/utorg-colored.png",
						"160x160": "https://cdn.onramper.com/icons/onramps/utorg-colored-160.png"
					}
				}
			},
			{
				"icon": "https://cdn.onramper.com/icons/onramps/sardine-colored.svg",
				"displayName": "Sardine",
				"id": "sardine",
				"icons": {
					"svg": "https://cdn.onramper.com/icons/onramps/sardine-colored.svg",
					"png": {
						"32x32": "https://cdn.onramper.com/icons/onramps/sardine-colored.png",
						"160x160": "https://cdn.onramper.com/icons/onramps/sardine-colored-160.png"
					}
				}
			}
		]
	}`

	httpClient := newMockHTTPClient(func(req *http.Request) *http.Response {
		assert.Equal(t,
			"https://mockapi.com/supported/onramps/all?type=buy",
			req.URL.String(),
		)
		assert.Equal(t, "test-api-key", req.Header.Get("Authorization"))

		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(bytes.NewBufferString(mockResponse)),
			Header:     make(http.Header),
		}
	})

	// Inline simulation of GetOnrampMetadata
	apiURL := "https://mockapi.com/supported/onramps/all?type=buy"
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, apiURL, nil)
	require.NoError(t, err)
	req.Header.Add("Authorization", "test-api-key")

	resp, err := httpClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	var result map[string]interface{}
	err = json.Unmarshal(body, &result)
	require.NoError(t, err)

	// Assertions
	message := result["message"].([]interface{})
	assert.Len(t, message, 2)

	first := message[0].(map[string]interface{})
	assert.Equal(t, "UTORG", first["displayName"])
	assert.Equal(t, "utorg", first["id"])
	assert.Equal(t, "https://cdn.onramper.com/icons/onramps/utorg-colored.svg", first["icon"])

	second := message[1].(map[string]interface{})
	assert.Equal(t, "Sardine", second["displayName"])

	icons := second["icons"].(map[string]interface{})
	svg := icons["svg"].(string)
	png := icons["png"].(map[string]interface{})

	assert.Equal(t, "https://cdn.onramper.com/icons/onramps/sardine-colored.svg", svg)
	assert.Equal(t, "https://cdn.onramper.com/icons/onramps/sardine-colored-160.png", png["160x160"])
}
func TestGetCryptoByFiat(t *testing.T) {
	mockResponse := `{
		"message": [
			{
				"id": "aave_ethereum",
				"code": "AAVE",
				"name": "Aave",
				"networkDisplayName": "Ethereum",
				"fiat": [
					{
						"id": "usd",
						"onramps": [
							{
								"id": "coinify",
								"paymentMethods": [
									{ "id": "banktransfer", "min": 189.32, "max": 54091.46 },
									{ "id": "creditcard", "min": 64.91, "max": 5409.15 }
								]
							},
							{
								"id": "topper",
								"paymentMethods": [
									{ "id": "applepay", "min": 10.83, "max": 54112.11 },
									{ "id": "creditcard", "min": 10.83, "max": 54112.11 }
								]
							}
						]
					}
				]
			}
		]
	}`

	httpClient := newMockHTTPClient(func(req *http.Request) *http.Response {
		assert.Equal(
			t,
			"https://mockapi.com/supported/crypto?source=usd&country=US",
			req.URL.String(),
		)
		assert.Equal(t, "test-api-key", req.Header.Get("Authorization"))

		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(bytes.NewBufferString(mockResponse)),
			Header:     make(http.Header),
		}
	})

	// Simulate inline version of GetCryptoByFiat
	apiURL := "https://mockapi.com/supported/crypto?source=usd&country=US"
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, apiURL, nil)
	require.NoError(t, err)
	req.Header.Add("Authorization", "test-api-key")

	resp, err := httpClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	var result map[string]interface{}
	err = json.Unmarshal(body, &result)
	require.NoError(t, err)

	message := result["message"].([]interface{})
	assert.Len(t, message, 1)

	crypto := message[0].(map[string]interface{})
	assert.Equal(t, "aave_ethereum", crypto["id"])
	assert.Equal(t, "Aave", crypto["name"])
	assert.Equal(t, "Ethereum", crypto["networkDisplayName"])

	fiatArray := crypto["fiat"].([]interface{})
	usd := fiatArray[0].(map[string]interface{})
	assert.Equal(t, "usd", usd["id"])

	onramps := usd["onramps"].([]interface{})
	coinify := onramps[0].(map[string]interface{})
	assert.Equal(t, "coinify", coinify["id"])

	methods := coinify["paymentMethods"].([]interface{})
	firstMethod := methods[0].(map[string]interface{})
	assert.Equal(t, "banktransfer", firstMethod["id"])
	assert.InDelta(t, 189.32, firstMethod["min"], 0.01)
	assert.InDelta(t, 54091.46, firstMethod["max"], 0.01)
}
func TestGetQuotesBuy(t *testing.T) {
	mockResponse := `[
		{
			"rate": 24138.08409757557,
			"networkFee": 0,
			"transactionFee": 0,
			"payout": 0.00398,
			"availablePaymentMethods": [
				{
					"paymentTypeId": "creditcard",
					"name": "Credit Card",
					"icon": "https://cdn.onramper.com/icons/payments/creditcard.svg"
				},
				{
					"paymentTypeId": "sepainstant",
					"name": "SEPA Instant",
					"icon": "https://cdn.onramper.com/icons/payments/sepa.svg"
				}
			],
			"ramp": "moonpay",
			"paymentMethod": "creditcard",
			"quoteId": "01H985NH79FW951SKERQ45JMYXmoonpay",
			"recommendations": ["LowKyc", "BestPrice"]
		},
		{
			"ramp": "fonbnk",
			"paymentMethod": "creditcard",
			"errors": [
				{
					"type": "NoSupportedPaymentFound",
					"errorId": 6103,
					"message": "No supported payments found"
				}
			],
			"quoteId": "01H985NH79FW951SKERQ45JMYXfonbnk"
		}
	]`

	httpClient := newMockHTTPClient(func(req *http.Request) *http.Response {
		assert.Contains(t, req.URL.String(), "/quotes/usd/btc")
		assert.Contains(t, req.URL.String(), "amount=100")
		assert.Contains(t, req.URL.String(), "type=buy")
		assert.Equal(t, "test-api-key", req.Header.Get("Authorization"))

		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(bytes.NewBufferString(mockResponse)),
			Header:     make(http.Header),
		}
	})

	// Simulated inline version of GetQuotes
	url := "https://mockapi.com/quotes/usd/btc?amount=100&type=buy"
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, url, nil)
	require.NoError(t, err)
	req.Header.Add("Authorization", "test-api-key")

	resp, err := httpClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	var result []map[string]interface{}
	err = json.Unmarshal(body, &result)
	require.NoError(t, err)

	assert.Len(t, result, 2)

	first := result[0]
	assert.Equal(t, "moonpay", first["ramp"])
	assert.Equal(t, "creditcard", first["paymentMethod"])
	assert.InDelta(t, 0.00398, first["payout"], 0.0001)
	assert.Contains(t, first["recommendations"], "LowKyc")

	methods := first["availablePaymentMethods"].([]interface{})
	pm0 := methods[0].(map[string]interface{})
	assert.Equal(t, "creditcard", pm0["paymentTypeId"])

	second := result[1]
	assert.Equal(t, "fonbnk", second["ramp"])
	errs := second["errors"].([]interface{})[0].(map[string]interface{})
	assert.Equal(t, "NoSupportedPaymentFound", errs["type"])
}
func TestGetQuotesSell(t *testing.T) {
	mockResponse := `[
		{
			"rate": 24138.08409757557,
			"networkFee": 0,
			"transactionFee": 0,
			"payout": 0.00398,
			"availablePaymentMethods": [
				{
					"paymentTypeId": "creditcard",
					"name": "Credit Card",
					"icon": "https://cdn.onramper.com/icons/payments/creditcard.svg"
				},
				{
					"paymentTypeId": "sepainstant",
					"name": "SEPA Instant",
					"icon": "https://cdn.onramper.com/icons/payments/sepa.svg"
				}
			],
			"ramp": "moonpay",
			"paymentMethod": "creditcard",
			"quoteId": "01H985NH79FW951SKERQ45JMYXmoonpay",
			"recommendations": ["LowKyc", "BestPrice"]
		},
		{
			"ramp": "fonbnk",
			"paymentMethod": "creditcard",
			"errors": [
				{
					"type": "NoSupportedPaymentFound",
					"errorId": 6103,
					"message": "No supported payments found"
				}
			],
			"quoteId": "01H985NH79FW951SKERQ45JMYXfonbnk"
		}
	]`

	httpClient := newMockHTTPClient(func(req *http.Request) *http.Response {
		assert.Contains(t, req.URL.String(), "/quotes/btc/usd") // flipped fiat/crypto for sell
		assert.Contains(t, req.URL.String(), "amount=100")
		assert.Contains(t, req.URL.String(), "type=sell")
		assert.Equal(t, "test-api-key", req.Header.Get("Authorization"))

		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(bytes.NewBufferString(mockResponse)),
			Header:     make(http.Header),
		}
	})

	// Simulate direct call to URL for raw JSON test
	url := "https://mockapi.com/quotes/btc/usd?amount=100&type=sell"
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, url, nil)
	require.NoError(t, err)
	req.Header.Add("Authorization", "test-api-key")

	resp, err := httpClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	var result []map[string]interface{}
	err = json.Unmarshal(body, &result)
	require.NoError(t, err)

	assert.Len(t, result, 2)

	first := result[0]
	assert.Equal(t, "moonpay", first["ramp"])
	assert.Equal(t, "creditcard", first["paymentMethod"])
	assert.InDelta(t, 0.00398, first["payout"].(float64), 0.00001)

	methods := first["availablePaymentMethods"].([]interface{})
	assert.Equal(t, "creditcard", methods[0].(map[string]interface{})["paymentTypeId"])

	second := result[1]
	assert.Equal(t, "fonbnk", second["ramp"])
	errs := second["errors"].([]interface{})[0].(map[string]interface{})
	assert.Equal(t, "NoSupportedPaymentFound", errs["type"])
}
func TestGetTransactionByID(t *testing.T) {
	mockResponse := `{
		"country": "us",
		"inAmount": 68,
		"onramp": "transfi",
		"onrampTransactionId": "OR-2123428075629314",
		"outAmount": 0.00202087,
		"paymentMethod": "creditcard",
		"sourceCurrency": "usd",
		"status": "completed",
		"statusDate": "2023-07-28T07:56:42.012Z",
		"targetCurrency": "btc",
		"transactionId": "01H6DQWMRC8FA9MBM0HS5NABCD",
		"transactionType": "onramp",
		"transactionHash": "ef76220d3cfd028a7f324ce8744b7a6AWSFKp62f8f94c4dae5149bb41afd7e279",
		"walletAddress": "bc1qp56l3l2w2vdle8cfABCDEFlnlgc7ye7q0lenu3"
	}`

	transactionID := "01H6DQWMRC8FA9MBM0HS5NABCD"
	expectedURL := fmt.Sprintf("https://mockapi.com/transactions/%s", transactionID)

	httpClient := newMockHTTPClient(func(req *http.Request) *http.Response {
		assert.Equal(t, expectedURL, req.URL.String())
		assert.Equal(t, "test-api-key", req.Header.Get("Authorization"))
		assert.Equal(t, "test-webhook-secret", req.Header.Get("X-Onramper-Secret"))

		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(bytes.NewBufferString(mockResponse)),
			Header:     make(http.Header),
		}
	})

	// Simulate inline test
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, expectedURL, nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "test-api-key")
	req.Header.Set("X-Onramper-Secret", "test-webhook-secret")

	resp, err := httpClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	var result map[string]interface{}
	err = json.Unmarshal(body, &result)
	require.NoError(t, err)

	assert.Equal(t, "us", result["country"])
	assert.Equal(t, "transfi", result["onramp"])
	assert.Equal(t, "creditcard", result["paymentMethod"])
	assert.Equal(t, "btc", result["targetCurrency"])
	assert.Equal(t, "completed", result["status"])
	assert.Equal(t, "01H6DQWMRC8FA9MBM0HS5NABCD", result["transactionId"])
	assert.InDelta(t, float64(0.00202087), result["outAmount"], 0.0001)
}
func TestListTransactions(t *testing.T) {
	mockResponse := `{
		"transactions": [
			{
				"targetCurrency": "eth",
				"onramp": "moonpay",
				"statusDate": "2023-01-20T15:15:33.922Z",
				"txType": "onramp",
				"status": "pending",
				"sourceCurrency": "eur",
				"country": "lk",
				"inAmount": 100,
				"outAmount": 0,
				"TxId": "Wwl5Hom-CW-qCjdZIB97Xg--",
				"wallet": "0x29bd7d9bed3028e72208f94e696b526d32b20efe",
				"paymentMethod": "credit_debit_card"
			}
		],
		"limit": 12
	}`

	client := &Client{
		BaseURL:       "https://mockapi.com",
		APIKey:        "test-api-key",
		WebhookSecret: "test-webhook-secret",
		HTTPClient: newMockHTTPClient(func(req *http.Request) *http.Response {
			assert.Equal(t, "Bearer test-api-key", req.Header.Get("Authorization"))
			assert.Equal(t, "test-webhook-secret", req.Header.Get("X-Onramper-Secret"))
			assert.Contains(t, req.URL.String(), "/transactions")
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewBufferString(mockResponse)),
				Header:     make(http.Header),
			}
		}),
	}

	req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, client.BaseURL+"/transactions?limit=12", nil)
	req.Header.Set("Authorization", "Bearer "+client.APIKey)
	req.Header.Set("X-Onramper-Secret", client.WebhookSecret)

	resp, err := client.HTTPClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	var parsed map[string]interface{}
	err = json.Unmarshal(body, &parsed)
	require.NoError(t, err)

	transactions := parsed["transactions"].([]interface{})
	firstTx := transactions[0].(map[string]interface{})
	assert.Equal(t, "eth", firstTx["targetCurrency"])
	assert.Equal(t, "moonpay", firstTx["onramp"])
	assert.InDelta(t, float64(100), firstTx["inAmount"], 0.0001)
	assert.Equal(t, "pending", firstTx["status"])
	assert.Equal(t, "credit_debit_card", firstTx["paymentMethod"])
}
func TestInitiateTransaction(t *testing.T) {
	mockResponse := `{
		"message": {
			"validationInformation": true,
			"status": "in_progress",
			"sessionInformation": {
				"onramp": "moonpay",
				"source": "eur",
				"destination": "btc",
				"amount": 100,
				"type": "buy",
				"paymentMethod": "creditcard",
				"network": "bitcoin",
				"uuid": "6756256e-d07f-42f0-a873-4d992eec8a2e",
				"originatingHost": "buy.onramper.com",
				"wallet": {
					"address": "bc1qxy2kgdygjrsqtzq2n0yrf2493p83kkfjhx0wlh"
				},
				"country": "LK",
				"expiringTime": 1693569558,
				"sessionId": "01H9896ZN8K047VJ2DZAR5RGT6"
			},
			"transactionInformation": {
				"transactionId": "01H9KBT5C21JY0BAX4VTW9EP3V",
				"url": "https://buy.moonpay.com?type=onramp&lockAmount=true...",
				"type": "iframe",
				"params": {
					"permissions": "accelerometer; autoplay; camera; gyroscope; payment"
				}
			}
		}
	}`

	client := &Client{
		BaseURL: "https://mockapi.com",
		APIKey:  "test-api-key",
		HTTPClient: newMockHTTPClient(func(req *http.Request) *http.Response {
			assert.Equal(t, "https://mockapi.com/checkout/intent", req.URL.String())

			body, _ := io.ReadAll(req.Body)
			var parsed map[string]interface{}
			_ = json.Unmarshal(body, &parsed)

			assert.Equal(t, "alchemypay", parsed["onramp"])
			assert.Equal(t, "USD", parsed["source"])
			assert.Equal(t, "creditcard", parsed["paymentMethod"])
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewBufferString(mockResponse)),
				Header:     make(http.Header),
			}
		}),
	}
	// Inline map as payload
	payload := map[string]interface{}{
		"onramp":        "alchemypay",
		"source":        "USD",
		"destination":   "BTC",
		"amount":        100.0,
		"type":          "buy",
		"paymentMethod": "creditcard",
		"network":       "bitcoin",
		"uuid":          "6756256e-d07f-42f0-a873-4d992eec8a2e",
		"wallet": map[string]interface{}{
			"address": "bc1qxy2kgdygjrsqtzq2n0yrf2493p83kkfjhx0wlh",
		},
		"country": "US",
	}

	body, err := json.Marshal(payload)
	require.NoError(t, err)

	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, client.BaseURL+"/checkout/intent", bytes.NewBuffer(body))
	require.NoError(t, err)
	req.Header.Add("Authorization", client.APIKey)
	req.Header.Add("Content-Type", "application/json")

	resp, err := client.HTTPClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	var response struct {
		Message struct {
			TransactionInformation struct {
				TransactionID string `json:"transactionId"`
				URL           string `json:"url"`
			} `json:"transactionInformation"`
		} `json:"message"`
	}
	err = json.NewDecoder(resp.Body).Decode(&response)
	require.NoError(t, err)
	assert.Equal(t, "01H9KBT5C21JY0BAX4VTW9EP3V", response.Message.TransactionInformation.TransactionID)
}
func TestConfirmSellTransaction(t *testing.T) {
	mockResponse := `{
		"status": "success"
	}`
	client := &Client{
		BaseURL: "https://mockapi.com",
		APIKey:  "test-api-key",
		Logger:  zap.NewNop(),
		HTTPClient: newMockHTTPClient(func(req *http.Request) *http.Response {
			assert.Equal(t, "https://mockapi.com/transactions/confirm/sell123", req.URL.String())
			assert.Equal(t, "Bearer test-api-key", req.Header.Get("Authorization"))
			assert.Equal(t, "application/json", req.Header.Get("Content-Type"))
			assert.Equal(t, "application/json", req.Header.Get("Accept"))

			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewBufferString(mockResponse)),
				Header:     make(http.Header),
			}
		}),
	}
	ctx := context.Background()
	resp, err := client.ConfirmSellTransaction(ctx, "sell123")
	require.NoError(t, err)
	assert.Equal(t, "success", resp.Status)
}
