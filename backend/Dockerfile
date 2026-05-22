FROM golang:1.25-alpine AS builder

WORKDIR /src

ARG GOPROXY=https://proxy.golang.org,direct
ENV GOPROXY=${GOPROXY}

RUN apk add --no-cache git ca-certificates

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/dotblue .

FROM alpine:3.20

WORKDIR /app

RUN apk add --no-cache ca-certificates tzdata && adduser -D -h /app appuser

COPY --from=builder /out/dotblue /app/main
COPY resource /app/resource
COPY manifest /app/manifest

RUN chown -R appuser:appuser /app && chmod +x /app/main

USER appuser

EXPOSE 8000

CMD ["./main"]
