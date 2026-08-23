FROM golang:1.26 AS builder

WORKDIR /app

COPY go.mod ./
COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/cloudscale-node ./cmd/node
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/cloudscale-gateway ./cmd/gateway


FROM debian:bookworm-slim

WORKDIR /app

COPY --from=builder /out/cloudscale-node /app/cloudscale-node
COPY --from=builder /out/cloudscale-gateway /app/cloudscale-gateway

EXPOSE 8001 8002 8003 9000

CMD ["/app/cloudscale-node"]

