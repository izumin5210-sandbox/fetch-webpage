# Buidler stage
FROM golang:1.18.0-bullseye as builder

WORKDIR /app

COPY go.mod go.sum .
RUN go mod download
COPY . .

RUN go build -o ./bin/fetch-webpage ./cmd/fetch-webpage

# Runtime stage
FROM debian:bullseye-slim

WORKDIR /app

RUN apt-get update && apt-get install -y \
    ca-certificates \
  && apt-get clean \
  && rm -rf /var/lib/apt/lists/*

COPY --from=builder /app/bin /usr/local/bin/

ENTRYPOINT ["/usr/local/bin/fetch-webpage"]
