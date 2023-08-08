FROM golang:1.23 AS builder
WORKDIR /app
COPY . ./
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=on go build -ldflags="-extldflags=-static"  -a -o ./cma-backend ./cmd/main.go

FROM scratch
WORKDIR /app
COPY --from=builder /app/cma-backend ./
COPY --from=alpine:latest /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

ENTRYPOINT [ "./cma-backend" ]
