FROM golang:1.25.4-alpine3.22 AS builder

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY . ./
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -a -o main ./cmd/app/main.go

FROM alpine:3.22

WORKDIR /app

COPY --from=builder /build/api/openapi ./docs
COPY --from=builder /build/main .

RUN addgroup -S appgroup && adduser -S appuser -G appgroup

ENTRYPOINT [ "./main" ]
