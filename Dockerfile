# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /pgbqe
RUN apk add --no-cache git ca-certificates make


COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN make build

# Final stage
FROM alpine:latest

RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S pgbqe -G appgroup

WORKDIR /root/

COPY --from=builder /pgbqe/bin/pgbqe .

RUN chown pgbqe:appgroup pgbqe

USER pgbqe

EXPOSE 8080

CMD ["./pgbqe"] 