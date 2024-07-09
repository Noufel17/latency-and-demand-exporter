# Build phase
FROM golang:1.22 AS builder
WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download && go mod verify
COPY . .
WORKDIR /build/cmd
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o latency-and-demand-exporter

# Production phase
FROM networkstatic/iperf3
WORKDIR /app
COPY --from=builder /build/cmd/latency-and-demand-exporter .
ENTRYPOINT [ "/app/latency-and-demand-exporter"]