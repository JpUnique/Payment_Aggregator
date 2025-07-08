package models

import (
	"time"

	"github.com/go-playground/validator/v10"
)

// SupportedCurrenciesResponse represents the response from the /supported endpoint.
type SupportedCurrenciesResponse struct {
	Message SupportedCurrencies `json:"message"`
}

// SupportedCurrencies contains lists of supported crypto and fiat currencies.
type SupportedCurrencies struct {
	Crypto []CryptoCurrency `json:"crypto"`
	Fiat   []FiatCurrency   `json:"fiat"`
}

// CryptoCurrency represents a supported cryptocurrency.
type CryptoCurrency struct {
	ID                 string `json:"id"`
	Code               string `json:"code"`
	Name               string `json:"name"`
	Symbol             string `json:"symbol"`
	Network            string `json:"network"`
	Decimals           int    `json:"decimals"`
	Address            string `json:"address"`
	ChainID            int    `json:"chainId"`
	Icon               string `json:"icon"`
	NetworkDisplayName string `json:"networkDisplayName"`
}

// FiatCurrency represents a supported fiat currency.
type FiatCurrency struct {
	ID     string `json:"id"`
	Code   string `json:"code"`
	Name   string `json:"name"`
	Symbol string `json:"symbol"`
	Icon   string `json:"icon"`
}

// PaymentTypesResponse represents the response from the /supported/payment-types endpoint.
type PaymentTypesResponse struct {
	Message map[string]PaymentType `json:"message"`
}

// PaymentType represents a single payment type.
type PaymentType struct {
	PaymentTypeID string `json:"paymentTypeId"`
	Name          string `json:"name"`
	Icon          string `json:"icon"`
}

// PaymentRequest represents the query parameters for fetching payment types.
type PaymentRequest struct {
	TransactionType string `form:"type" json:"type" binding:"omitempty,oneof=buy sell"`
	IsRecurring     bool   `form:"isRecurringPayment" json:"isRecurringPayment"`
	Destination     string `form:"destination" json:"destination" binding:"required"`
	Country         string `form:"country" json:"country,omitempty"`
	Subdivision     string `form:"subdivision" json:"subdivision,omitempty"`
}

// PaymentLimit represents the minimum and maximum limits for a specific payment provider.
type PaymentLimit struct {
	Min float64 `json:"min"`
	Max float64 `json:"max"`
}

// PaymentDetails represents the details about a payment method, including currency status and provider-specific limits.
type PaymentDetails struct {
	CurrencyStatus string                  `json:"currencyStatus"`
	Limits         map[string]PaymentLimit `json:"limits"`
}

// PaymentMethod represents a single supported payment method.
type PaymentMethod struct {
	PaymentTypeID string         `json:"paymentTypeId"`
	Name          string         `json:"name"`
	Icon          string         `json:"icon"`
	Details       PaymentDetails `json:"details"`
}

// PaymentResponse represents the API response for supported payment methods based on source currency.
type PaymentResponse struct {
	Message []PaymentMethod `json:"message"`
	Error   string          `json:"error,omitempty"`
}

// DefaultsResponse represents the top-level response structure.
type DefaultsResponse struct {
	Message DefaultsMessage `json:"message"`
}

// DefaultsMessage represents the "message" object in the response.
type DefaultsMessage struct {
	Recommended DefaultSetting  `json:"recommended"`
	Defaults    CountryDefaults `json:"defaults"`
}

// CountryDefaults is a map of country codes to their default settings.
type CountryDefaults map[string]DefaultSetting

// DefaultSetting represents the default settings for a country or the recommended settings.
type DefaultSetting struct {
	Source        string  `json:"source"`
	Target        string  `json:"target"`
	Amount        float64 `json:"amount"`
	PaymentMethod string  `json:"paymentMethod"`
	Provider      string  `json:"provider"`
	Country       string  `json:"country,omitempty"`
}

type OnramperClient struct {
	validate *validator.Validate
}

func NewCustomer() *OnramperClient {
	return &OnramperClient{
		validate: validator.New(),
	}
}

// TransactionListQuery represents the query parameters for listing transactions.
type TransactionListQuery struct {
	StartDateTime  string `form:"startDateTime"`
	EndDateTime    string `form:"endDateTime"`
	Limit          int    `form:"limit"`
	TransactionIDs string `form:"transactionIds"`
	Cursor         string `form:"cursor"`
}

// TransactionItem represents a single transaction record.
type TransactionItem struct {
	TargetCurrency        string    `json:"targetCurrency"`
	Onramp                string    `json:"onramp"`
	StatusDate            time.Time `json:"statusDate"`
	TxType                string    `json:"txType"`
	TxHash                string    `json:"txHash,omitempty"`
	Status                string    `json:"status"`
	APIKey                string    `json:"ApiKey"`
	SourceCurrency        string    `json:"sourceCurrency"`
	Country               string    `json:"country"`
	Info                  string    `json:"info,omitempty"`
	InAmount              float64   `json:"inAmount"`
	OutAmount             float64   `json:"outAmount,omitempty"`
	TxID                  string    `json:"TxId"`
	ExternalTransactionID string    `json:"externalTransactionId,omitempty"`
	SK                    string    `json:"sk"`
	Wallet                string    `json:"wallet,omitempty"`
	PaymentMethod         string    `json:"paymentMethod"`
}

type TransactionListResponse struct {
	Transactions []TransactionItem `json:"transactions"`
	Limit        int               `json:"limit"`
}

// TransactionResponse represents the response for a single transaction.
type TransactionResponse struct {
	Country             string    `json:"country"`
	InAmount            float64   `json:"inAmount"`
	Onramp              string    `json:"onramp"`
	OnrampTransactionID string    `json:"onrampTransactionId"`
	OutAmount           float64   `json:"outAmount"`
	PaymentMethod       string    `json:"paymentMethod"`
	SourceCurrency      string    `json:"sourceCurrency"`
	Status              string    `json:"status"`
	StatusDate          time.Time `json:"statusDate"`
	TargetCurrency      string    `json:"targetCurrency"`
	TransactionID       string    `json:"transactionId"`
	TransactionType     string    `json:"transactionType"`
	TransactionHash     string    `json:"transactionHash,omitempty"`
	WalletAddress       string    `json:"walletAddress"`
}

// QuoteQueryParams represents the query parameters for the /quotes/{fiat}/{crypto} endpoint.
type QuoteQueryParams struct {
	Amount             float64 `form:"amount"`
	PaymentMethod      string  `form:"paymentMethod"`
	UUID               string  `form:"uuid"`
	ClientName         string  `form:"clientName"`
	Type               string  `form:"type"`
	WalletAddress      string  `form:"walletAddress"`
	IsRecurringPayment bool    `form:"isRecurringPayment"`
	Input              string  `form:"input"`
	Country            string  `form:"country"`
	TxInitiation       bool    `form:"txInitiation"`
}

// QuoteResponse represents a single quote from the /quotes/{fiat}/{crypto} endpoint.
type QuoteResponse struct {
	Rate                    float64              `json:"rate"`
	NetworkFee              float64              `json:"networkFee"`
	TransactionFee          float64              `json:"transactionFee"`
	Payout                  float64              `json:"payout"`
	AvailablePaymentMethods []QuotePaymentMethod `json:"availablePaymentMethods"`
	Ramp                    string               `json:"ramp"`
	PaymentMethod           string               `json:"paymentMethod"`
	QuoteID                 string               `json:"quoteId"`
	Recommendations         []string             `json:"recommendations"`
	Errors                  []QuoteError         `json:"errors,omitempty"`
}

// QuotePaymentMethod represents a payment method.
type QuotePaymentMethod struct {
	PaymentTypeID string              `json:"paymentTypeId"`
	Name          string              `json:"name"`
	Icon          string              `json:"icon"`
	Details       QuotePaymentDetails `json:"details"`
}

// QuotePaymentDetails represents the details of a payment method.
type QuotePaymentDetails struct {
	CurrencyStatus string        `json:"currencyStatus"`
	Limits         PaymentLimits `json:"limits"`
}

// PaymentLimits represents the limits for a payment method.
type PaymentLimits struct {
	ProviderLimits  map[string]LimitRange `json:"-"`
	AggregatedLimit LimitRange            `json:"aggregatedLimit"`
}

// LimitRange represents the minimum and maximum limits for a payment method.
type LimitRange struct {
	Min float64 `json:"min"`
	Max float64 `json:"max"`
}

// QuoteError represents an error in a quote.
type QuoteError struct {
	Type    string `json:"type"`
	ErrorID int    `json:"errorId"`
	Message string `json:"message"`
}

type SupportedAssetsResponse struct { //nolint:revive // Renaming would break API compatibility.
	Message struct { //nolint:revive // Renaming would break API compatibility.
		Assets  []interface{} `json:"assets"`
		Country string        `json:"country"`
	} `json:"message"`
}

type BuyAsset struct {
	Fiat           string   `json:"fiat"`
	PaymentMethods []string `json:"paymentMethods"`
	Crypto         []string `json:"crypto"`
}
type SellAsset struct {
	Fiat           []string `json:"fiat"`
	Crypto         string   `json:"crypto"`
	PaymentMethods []string `json:"paymentMethods"`
}

// TransactionType enum for buy/sell.
type TransactionType string

const (
	BuyTransaction  TransactionType = "buy"
	SellTransaction TransactionType = "sell"
)

// AssetRequest represents the parameters for querying supported assets.
type AssetRequest struct {
	Source         string          `json:"source"`
	Country        string          `json:"country"`
	Type           TransactionType `json:"type"`
	Onramps        string          `json:"onramps"`
	PaymentMethods string          `json:"paymentMethods"`
	Subdivision    string          `json:"subdivision"`
}

// OnrampsQuery represents the structure of the response from the Onramper API.
type OnrampsQuery struct {
	TransactionType string `form:"type"`
	Source          string `form:"source"`
	Destination     string `form:"destination"`
	Country         string `form:"country"`
	Subdivision     string `form:"subdivision"`
}
type Onramp struct {
	Onramp                   string        `json:"onramp"`
	Icon                     string        `json:"icon"`
	Icons                    OnrampIconSet `json:"icons"`
	DisplayName              string        `json:"displayName"`
	Country                  string        `json:"country"`
	PaymentMethods           []string      `json:"paymentMethods"`
	RecommendedPaymentMethod string        `json:"recommendedPaymentMethod"`
	Recommendations          []interface{} `json:"recommendations"`
}

type OnrampIconSet struct { //nolint:revive // Renaming would break API compatibility.
	SVG string           `json:"svg"`
	PNG OnrampImageSizes `json:"png"`
}

type OnrampImageSizes struct {
	Size32x32   string `json:"32x32"`
	Size160x160 string `json:"160x160"`
}

// OnrampResponse represents the structure of the response from Onramper API.
type OnrampResponse struct {
	Message []Onramp `json:"message"`
}

// CryptoFiatResponse represents a fiat currency (e.g., USD) and its associated onramps.
type CryptoFiatResponse struct {
	Message []AssetMessage `json:"message"`
}

type AssetMessage struct {
	ID                 string     `json:"id"`
	Code               string     `json:"code"`
	Name               string     `json:"name"`
	Fiat               []FiatItem `json:"fiat"`
	NetworkDisplayName string     `json:"networkDisplayName"`
}

type FiatItem struct {
	ID      string    `json:"id"`
	Onramps []Onramps `json:"onramps"`
}

type Onramps struct {
	ID             string           `json:"id"`
	PaymentMethods []PaymentMethods `json:"paymentMethods"`
}

type PaymentMethods struct {
	ID  string  `json:"id"`
	Min float64 `json:"min"`
	Max float64 `json:"max"`
}

// OnrampMetadataResponse represents the structure of the response from the Onramper API..
type OnrampMetadataResponse struct {
	Message []OnrampMetadata `json:"message"`
}
type OnrampMetadata struct {
	Icon        string  `json:"icon"`
	DisplayName string  `json:"displayName"`
	ID          string  `json:"id"`
	Icons       IconSet `json:"icons"`
}

type IconSet struct { //nolint:revive // Renaming would break API compatibility.
	SVG string     `json:"svg"`
	PNG ImageSizes `json:"png"`
}

type ImageSizes struct { //nolint:revive // Renaming would break API compatibility.
	Size32x32   string `json:"32x32"`
	Size160x160 string `json:"160x160"`
}

type InitiateTransactionRequest struct {
	Onramp        string  `json:"onramp"`
	Source        string  `json:"source"`
	Destination   string  `json:"destination"`
	Amount        float64 `json:"amount"`
	Type          string  `json:"type"`
	PaymentMethod string  `json:"paymentMethod"`
	Network       string  `json:"network"`
	UUID          string  `json:"uuid"`
	Wallet        struct {
		Address string `json:"address"`
	} `json:"wallet"`
	Country string `json:"country"`
}

type InitiateTransactionResponse struct {
	Message struct {
		ValidationInformation bool   `json:"validationInformation"`
		Status                string `json:"status"`

		SessionInformation struct {
			Onramp          string  `json:"onramp"`
			Source          string  `json:"source"`
			Destination     string  `json:"destination"`
			Amount          float64 `json:"amount"`
			Type            string  `json:"type"`
			PaymentMethod   string  `json:"paymentMethod"`
			Network         string  `json:"network"`
			UUID            string  `json:"uuid"`
			OriginatingHost string  `json:"originatingHost"`
			Wallet          struct {
				Address string `json:"address"`
			} `json:"wallet"`
			SupportedParams struct {
				Theme struct {
					IsDark             bool   `json:"isDark"`
					ThemeName          string `json:"themeName"`
					PrimaryColor       string `json:"primaryColor"`
					SecondaryColor     string `json:"secondaryColor"`
					PrimaryTextColor   string `json:"primaryTextColor"`
					SecondaryTextColor string `json:"secondaryTextColor"`
					CardColor          string `json:"cardColor"`
					BorderRadius       *int   `json:"borderRadius"`
				} `json:"theme"`
				PartnerData struct {
					RedirectURL struct {
						Success string `json:"success"`
					} `json:"redirectUrl"`
				} `json:"partnerData"`
			} `json:"supportedParams"`
			Country      string `json:"country"`
			ExpiringTime int64  `json:"expiringTime"`
			SessionID    string `json:"sessionId"`
		} `json:"sessionInformation"`

		TransactionInformation struct {
			TransactionID string `json:"transactionId"`
			URL           string `json:"url"`
			Type          string `json:"type"`
			Params        struct {
				Permissions string `json:"permissions"`
			} `json:"params"`
		} `json:"transactionInformation"`
	} `json:"message"`
}

type SellTransactionConfirmationResponse struct {
	Status string `json:"status"`
}
