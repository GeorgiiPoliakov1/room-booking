ARG GO_VERSION=1.25.7
ARG ALPINE_VERSION=3.19

FROM golang:${GO_VERSION}-alpine AS builder

RUN apk add --no-cache git ca-certificates

WORKDIR /build

COPY go.mod ./
RUN go mod download && go mod verify

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s" \
    -o /build/room-booking \
    ./cmd/api

FROM alpine:${ALPINE_VERSION}

LABEL go_version="${GO_VERSION}"

RUN apk --no-cache add ca-certificates

RUN adduser -D -u 1000 appuser

WORKDIR /app

COPY --from=builder /build/room-booking .
RUN chown -R appuser:appuser /app
USER appuser

EXPOSE 8080

CMD ["./room-booking"]