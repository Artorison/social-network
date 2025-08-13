FROM golang:1.24.2-alpine3.21 AS builder
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN apk add --no-cache make && \
    CGO_ENABLED=0 GOFLAGS="-trimpath" make build

FROM alpine:3.21 AS runner
WORKDIR /app

RUN apk add --no-cache ca-certificates tzdata && update-ca-certificates

COPY --from=builder /app/bin/social_network ./social_network
COPY static/ ./static/

EXPOSE 3000
CMD ["./social_network"]
