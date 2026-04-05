FROM golang:1.25-alpine AS builder

WORKDIR /src

RUN apk add --no-cache ca-certificates git

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags="-s -w" -o /out/sso ./cmd/sso

FROM alpine:3.21

RUN apk add --no-cache ca-certificates tzdata wget && adduser -D -h /app appuser

WORKDIR /app

COPY --from=builder /out/sso /app/sso

USER appuser

EXPOSE 8080

CMD ["./sso"]
