FROM golang:alpine AS builder

WORKDIR /app

RUN apk add --no-cache git

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /app/bin/all ./cmd/all
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /app/bin/catalog ./cmd/catalog
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /app/bin/auth ./cmd/auth
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /app/bin/collector ./cmd/collector

FROM alpine:3.20

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

COPY --from=builder /app/bin/all /app/all
COPY --from=builder /app/bin/catalog /app/catalog
COPY --from=builder /app/bin/auth /app/auth
COPY --from=builder /app/bin/collector /app/collector
COPY --from=builder /app/migrations /app/migrations

EXPOSE 4242 4243 50051

CMD ["/app/all"]
