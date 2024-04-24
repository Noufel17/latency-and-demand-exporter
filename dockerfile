# Build phase
FROM golang:1.22 AS builder
# Next line is just for debug
RUN ldd --version
WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download && go mod verify
COPY . .
WORKDIR /build/cmd
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o latency-exporter

# Production phase
FROM alpine:3.14
# Next line is just for debug
RUN ldd; exit 0
WORKDIR /app
COPY --from=builder /build/cmd/latency-exporter .
ENTRYPOINT [ "/app/latency-exporter"]