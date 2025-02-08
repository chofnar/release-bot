FROM arm64v8/golang:1.23.6-bullseye as builder

WORKDIR /app

COPY go.* ./
RUN go mod download

COPY . ./

RUN go build -mod=readonly -v -o bot

FROM arm64v8/debian:latest
RUN set -x && apt-get update && apt-get upgrade -y && DEBIAN_FRONTEND=noninteractive apt-get install -y \
    ca-certificates && \
    rm -rf /var/lib/apt/lists/*

COPY --from=builder /app/bot /app/bot

CMD ["/app/bot"]
