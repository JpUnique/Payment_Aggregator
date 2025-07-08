# Onramper API Endpoints
The `endpoint.go` file defines various endpoints to interact with Onramper’s API:

#### Get Supported Currencies
```http
GET /supported
```
#### Query Paramters
```
type=buy&country=us&subdivision=us-ny

```
#### Response Body
```json
{
    "message": {
        "crypto": [
            {
                "id": "1inch_bsc",
                "code": "1INCH",
                "name": "1inch Network",
                "symbol": "1inch",
                "network": "bsc",
                "decimals": 0,
                "address": "0x111111111117dc0aa78b770fa6a738034120c302",
                "chainId": 56,
                "icon": "https://cdn.onramper.com/icons/crypto/webp/1inch_bsc.webp",
                "networkDisplayName": "BNB Smart Chain"
            },
            // ... (truncated list)
        ],
         "fiat": [
            {
                "id": "aed",
                "code": "AED",
                "name": "UAE-Dirham",
                "symbol": "د.إ",
                "icon": "https://cdn.onramper.com/icons/tokens/aed.svg"
            },
            // ... (fiat currency data)
        ]
    }
}
```

#### Get Payment Methods
```http
GET /supported/payment-types
```
#### Query Paramters
```
type=Default(buy)(buy/sell)&country=us&isRecurringPayment=false

```
#### Response Body
```json
{
 {
    "message": {
        "7eleven": {
            "paymentTypeId": "7eleven",
            "name": "7 Eleven",
            "icon": "https://cdn.onramper.com/icons/payments/banktransfer.svg"
        },
        "acb": {
            "paymentTypeId": "acb",
            "name": "ACB",
            "icon": "https://cdn.onramper.com/icons/payments/banktransfer.svg"
        },
        // ... (truncated list)
        }
 }
}
```
#### Get Payment Methods by Currency
```http
GET /supported/payment-types/{currency}
```
#### Query Paramters
```
type=Default(buy)(buy/sell)&country=us&isRecurringPayment=false&destination=btc&us-ny

```
#### Response Body
```json
{
 {
    {
        "paymentTypeId": "applepay",
        "name": "Apple Pay",
        "icon": "https://cdn.onramper.com/icons/payments/applepay.svg",
        "details": {
            "currencyStatus": "SourceAndDestSupported",
            "limits": {
                "aggregatedLimit": {
                    "min": 5,
                    "max": 100000
                },
                "alchemypay": {
                    "min": 100,
                    "max": 2000
                },
                "gatefi": {
                    "min": 10.88,
                    "max": 16200.77
                },
            }
        }
    },
     // ... (truncated list)
  }
 }

```
#### Get Defaults
```http
GET /supported/fefaults/all
```
#### Query Paramters
```
type=Default(buy)(buy/sell)&country=us&subdivision=fus-ny

```
#### Response Body
```json
{
 {
    "message": {
        "recommended": {
            "source": "USD",
            "target": "BTC",
            "amount": 300,
            "paymentMethod": "debitcard",
            "provider": "topper",
            "country": "us"
        },
        "defaults": {
            "ad": {
                "source": "EUR",
                "target": "BTC",
                "amount": 300,
                "paymentMethod": "debitcard",
                "provider": "banxa"
            },
        // ... (truncated list)
        }
    }
 }
}
```
#### Get Available Assets
```http
GET /supported/assets
```
#### Query Paramters
```
source=FIATorCrypto&type=Default(buy)(buy/sell)&paymentMethods=creditcard&onramps=topper&country=us&isRecurringPayment=false&destination=btc&us-ny

```
#### Response Body
```json
{
 {
   "message": {
        "assets": [
            {
                "crypto": [
                    "ach_ethereum",
                    "ada_cardano",
                    "algo_algorand",
                    "alph_alph",
                    "ape_ethereum",
                    "arb_arbitrum",
                    ]
                         // ... (truncated list)
                     "fiat": "pkr",
                "paymentMethods": [
                    "creditcard"
                ]
            },
        ]
   }
  }
 }

```
#### Get Onramps
```http
GET /supported/onramps
```
#### Query Paramters
```
source=usd&type=Default(buy)(buy/sell)&destination=btc&country=us&destination=btc&us-ny&subdivision=us-ny

```
#### Response Body
```json
{
 {
      "message": [
        {
            "onramp": "alchemypay",
            "icon": "https://cdn.onramper.com/icons/onramps/alchemypay-colored.svg",
            "icons": {
                "svg": "https://cdn.onramper.com/icons/onramps/alchemypay-colored.svg",
                "png": {
                    "32x32": "https://cdn.onramper.com/icons/onramps/alchemypay-colored.png",
                    "160x160": "https://cdn.onramper.com/icons/onramps/alchemypay-colored-160.png"
                }
            },
            "displayName": "Alchemy Pay",
            "country": "US",
            "paymentMethods": [
                "applepay",
                "creditcard",
                "googlepay",
                "neteller",
                "skrill"
            ],
            "recommendedPaymentMethod": "applepay",
            "recommendations": []
             // ... (truncated list)
        },
      ]
  }
 }

```
#### Get Onramps Metadata
```http
GET /supported/onramps/all
```
#### Query Paramters
```
type=Default(buy)(buy/sell)

```
#### Response Body
```json
{
 {
      "message": [
           {
            "icon": "https://cdn.onramper.com/icons/onramps/alchemypay-colored.svg",
            "displayName": "Alchemy Pay",
            "id": "alchemypay",
            "icons": {
                "svg": "https://cdn.onramper.com/icons/onramps/alchemypay-colored.svg",
                "png": {
                    "32x32": "https://cdn.onramper.com/icons/onramps/alchemypay-colored.png",
                    "160x160": "https://cdn.onramper.com/icons/onramps/alchemypay-colored-160.png"
                }
            }
             // ... (truncated list)
        },
      ]
  }
 }
```
#### Get Crypto Currencies by Fiat
```http
GET /supported/crypto
```
#### Query Paramters
```
source=buy&country=us

```
#### Response Body
```json
{
    "message": {
        {
            "id": "aave_ethereum",
            "code": "AAVE",
            "name": "Aave",
            "fiat": [
                {
                    "id": "usd",
                    "onramps": [
                        {
                            "id": "coinify",
                            "paymentMethods": [
                                {
                                    "id": "banktransfer",
                                    "min": 189.32,
                                    "max": 54091.46
                                },
                                {
                                    "id": "creditcard",
                                    "min": 64.91,
                                    "max": 5409.15
                                }
                            ]
                        },
                        // ... (truncated list)
                  ]
                }
            ]
        }
    }
}
```
#### Get Buy Quotes
```http
GET /quotes/{fiat}/{crypto}  Buy Quotes
GET /quotes/{crypto}/{fiat}  Sell Quotes
```
#### Query Paramters
```
amount=&paymentMethod=&uuid=&clientName=&type=buy&walletAddress=&isRecurringPayment=input=eur&country=us

amount=&paymentMethod=&uuid=&clientName=&type=sell&walletAddress=&isRecurringPayment=input=eur&country=us

```
#### Response Body
```json
{
  {
        "rate": 78240.2821657536,
        "networkFee": 0.27,
        "transactionFee": 19.51,
        "payout": 0.01278114,
        "availablePaymentMethods": [
            {
                "paymentTypeId": "applepay",
                "name": "Apple Pay",
                "icon": "https://cdn.onramper.com/icons/payments/applepay.svg",
                "details": {
                    "currencyStatus": "SourceAndDestSupported",
                    "limits": {
                        "aggregatedLimit": {
                            "min": 17,
                            "max": 10000
                        }
                    }
                }
            },
             // ... (truncated list)
        ]
  }
}
```
#### Get Transaction
```http
GET /transaction/{transactionId}
```
#### Authentication Header
```go
	req.Header.Set("Authorization", h.APIKey)
	req.Header.Set("x-onramper-secret", h.WebhookSecret)
```
#### Response Body
``` json
{
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
}
```

#### Get List Transaction
```http
GET /transactions
```
#### Authentication Header
```go
	req.Header.Set("Authorization", h.APIKey)
	req.Header.Set("x-onramper-secret", h.WebhookSecret)
```
#### Query Paramters
```
startDateTime=StarttimeinISO8601standard&endDateTime=EndtimeinISO8601standard&limt=50&trasactionIds=&cursor=pagination

```
#### Response Body
```json
{
  "transactions": [
    {
      "targetCurrency": "eth",
      "onramp": "moonpay",
      "statusDate": "2023-01-20T15:15:33.922Z",
      "txType": "onramp",
      "txHash": "",
      "status": "pending",
      "ApiKey": "pk_prod_01GTC8JT9MDSW8G11HPPKSVBTJ",
      "sourceCurrency": "eur",
      "country": "lk",
      "info": "response-pulling-event-testing",
      "inAmount": 100,
      "outAmount": 0,
      "TxId": "Wwl5Hom-CW-qCjdZIB97Xg--",
      "externalTransactionId": "4a28307f-0cc8-47ec-aaf9-278b2ac2f2e4",
      "sk": "2023-01-20T15:15:33.922Z",
      "wallet": "0x29bd7d9bed3028e72208f94e696b526d32b20efe",
      "paymentMethod": "credit_debit_card"
    },
      // ... (truncated list)
  ]
}
    ```

#### Post Initiate Transaction
```http
POST /checkout/intent
```
#### Authentication Header
```
	req.Header.Set("Authorization", h.APIKey)

```
#### Request Body
``` json
{
        "onramp": "gatefi",
        "source": "USD",
        "destination": "BTC",
        "amount": 100.0,
        "type": "buy",
        "paymentMethod": "creditcard",
        "network": "bitcoin",
        "uuid": "6756256e-d07f-42f0-a873-4d992eec8a2e",
        "wallet": {
            "address": "bc1qxy2kgdygjrsqtzq2n0yrf2493p83kkfjhx0wlh"
        },
        "country": "US"
}
```
#### Response Body

```json
{
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
      "supportedParams": {
        "theme": {
          "isDark": false,
          "themeName": "light-theme",
          "primaryColor": "#241D1C",
          "secondaryColor": "#FFFFFF",
          "primaryTextColor": "#141519",
          "secondaryTextColor": "#6B6F80",
          "cardColor": "#F6F7F9",
          "borderRadius": null
        },
        "partnerData": {
          "redirectUrl": {
            "success": "http%3A%2F%2Fredirecturl.com%2F"
          }
        }
      },
      "country": "LK",
      "expiringTime": 1693569558,
      "sessionId": "01H9896ZN8K047VJ2DZAR5RGT6"
    },
    "transactionInformation": {
      "transactionId": "01H9KBT5C21JY0BAX4VTW9EP3V",
      "url": "https://buy.moonpay.com?type=onramp&lockAmount=true&baseCurrencyAmount=100&baseCurrencyCode=EUR&currencyCode=BTC&walletAddress=bc1qxy2kgdygjrsqtzq2n0yrf2493p83kkfjhx0wlh&externalTransactionId=01H9896ZR1WS9RV1G5Y4RY91Z1&colorCode=%23241D1C&theme=light&apiKey=pk_live_MJouyH3Fe3BVaNOaHCt5SvBfJ5Mk7RvQ&redirectURL=http%3A%2F%2Fredirecturl.com%2F&signature=jPFeIEwDJkvpOyhl938ErS69do1K5FJUS%2FvHV4M%2BzW8%3D",
      "type": "iframe",
      "params": {
        "permissions": "accelerometer; autoplay; camera; gyroscope; payment"
      }
    }
  }
}

```

####  Webhook Payload 
```http
POST baseurl/onramper/webhook
```
#### Request Headers:
```
X-Onramper-Signature: <HMAC signature>
Content-Type: application/json
```
#### Request Body (Example):
```json
{
    "country": "us",
    "inAmount": 100,
    "onramp": "gatefi",
    "onrampTransactionId": "8bf94c80-test-aabb-851-143835984d1d",
    "outAmount": 3.83527521,
    "paymentMethod": "creditcard",
    "partnerContext": "",
    "sourceCurrency": "usd",
    "status": "pending/completed/paid/new/failed/canceled",
    "statusDate": "2023-08-09T13:15:18.725Z",
    "targetCurrency": "sol",
    "transactionId": "01H7D547TESTV2RQJ52ZAB7WF7",
    "transactionType": "buy",
    "transactionHash": "",
    "walletAddress": "testG15oy66q7cU6aNige54PxLLEfGZvRsAADjbF7D4"
}
```
#### Response:
```json
{
  "message": "Webhook processed successfully"
}
```
