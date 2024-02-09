FROM arm64v8/golang:1.22.0-bullseye as builder

WORKDIR /app

COPY go.* ./
RUN go mod download

COPY . ./

RUN go build -mod=readonly -v -o bot

FROM arm64v8/debian:buster-slim
RUN set -x && apt-get update && DEBIAN_FRONTEND=noninteractive apt-get install -y \
    ca-certificates && \
    rm -rf /var/lib/apt/lists/*

COPY --from=builder /app/bot /app/bot

CMD ["/app/bot"]
