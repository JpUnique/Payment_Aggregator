# Payment Aggregators System

## Overview
This project is a **Golang-based fiat on/off-ramp service** that integrates with **Onramper** for transaction processing also authentication and KYC Synchronization with TerraceKYC based on the  Transaction status. It ensures secure and reliable fiat-to-crypto transactions using webhook handling, authentication, and database integration.

## Features
- **OnrmapClient Handler**: Processes the API endpoint request and response.
- **Endpoint Manager**: Handles the algorithm and logic to make request.
- **Webhook Handling**: Processes Onramper webhook events for transactions.
- **KYC Synchronization**: Updates Terrace KYC based on webhook events received.
- **Secure Authentication**: Manages user authentication with API key-based security.
- **HMAC Signature Verification**: Ensures webhook requests are authenticated.
- **Database Integration**: Uses Hasura(GraphQL), Kubernetes, Docker to access terrace db.
- **Modular Design**: Organized into `cmd`,`docs`, `pkg` (`database`, `models`, `onrampclient`, `onramper`, `utils`).
- **Comprehensive Testing**: Uses **Testify** for unit and Local testing.
- **RESTful API Endpoints**: Provides multiple endpoints to interact with Onramper services.

## Technologies Used
- **Golang**
- **Gin** (for API handling)
- **Hasura/PostgreSQL/Kubernetes** (Local testing)
- **Testify** (for testing framework)
- **Ngrok** (for webhook tunneling during development)

## Installation & Setup
### Prerequisites
Ensure you have **Go 1.18+** installed.

```sh
# Clone the repository
git clone https://github.com/subdialia/fiat-ramp-service.git
cd fiat-ramp-service

# Install dependencies
go mod tidy
```

### Environment Variables
Create a **.env** file and configuration:
```
ONRAMPER_API_KEY=<your_onramper_api_key>
ONRAMPER_BASE_URL=https://api.onramper.com
WEBHOOK_URL=<terrace base url/onramper/webhook>
DATABASE_URL=<terrace database>
ENVIRONMENT=staging
```

## Running the Service
### Using Makefile
```sh
make run
make test
make fmt
make tidy
make build

```

## Database Setup

### 1. Local Development with Hasura
```bash
# Run Hasura GraphQL Engine with Docker
docker run -d \
  -p 8080:8080 \
  -e HASURA_GRAPHQL_DATABASE_URL=postgres://postgres:postgrespassword@terrace-pg-rw:5432/postgres \
  -e HASURA_GRAPHQL_ENABLE_CONSOLE=true \
  hasura/graphql-engine:v2.32.0

# Access Hasura console at
http://localhost:8080

# Port-forward PostgreSQL service to local machine
kubectl port-forward svc/terrace-pg-rw 5432:5432

# Now you can connect to the database locally:
psql -h localhost -U postgres -d postgres -p 5432

DATABASE_URL=postgres://postgres:postgrespassword@localhost:5432/postgres?sslmode=disable
HASURA_ENDPOINT=http://localhost:8080/v1/graphql

---
 **Developed by Johnpaul**
