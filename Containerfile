FROM registry.access.redhat.com/ubi9/go-toolset:1.23 AS builder

WORKDIR /app
COPY . .

RUN CGO_ENABLED=0 go build -o /tmp/bucket-store-server .

FROM registry.access.redhat.com/ubi9/ubi-micro:latest

COPY --from=builder /tmp/bucket-store-server /bucket-store-server

EXPOSE 8080

ENTRYPOINT ["/bucket-store-server"]
