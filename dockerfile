FROM golang:latest AS builder

WORKDIR /go/src/app

COPY go.mod ./
RUN go mod download

COPY . .
RUN go build -o latency-node-exporter

FROM alpine:latest AS runner

WORKDIR /app

COPY --from=builder /go/src/app/latency-node-exporter ./

ENTRYPOINT ["latency-node-exporter"]
