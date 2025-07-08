FROM golang:1.23 AS builder

WORKDIR /app

# Setup Git for SSH
RUN git config --global url."git@github.com:".insteadOf "https://github.com/"
RUN mkdir ~/.ssh
RUN ssh-keyscan -H github.com >> ~/.ssh/known_hosts
RUN go env -w GOPRIVATE="github.com/subdialia/*"

COPY go.mod go.sum ./
RUN --mount=type=ssh go mod download

COPY . .

RUN --mount=type=ssh  go build -o fiat-ramp-service .

FROM gcr.io/distroless/base-debian12:nonroot

COPY --from=builder --chown=nonroot:nonroot /app/fiat-ramp-service /app/

ENTRYPOINT ["/app/fiat-ramp-service", "fiat-ramp-service"]