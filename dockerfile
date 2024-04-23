FROM golang:1.20.4-alpine as builder
RUN apk add build-base

WORKDIR /app
# Cache and install dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy app files
COPY . .
# Build app
RUN go build \
  -o custom-node-exporter \
  ./main.go

FROM golang:1.21.6-alpine as runner

COPY --from=builder /app/custom-node-exporter /usr/local/bin/custom-node-exporter

CMD custom-node-exporter
