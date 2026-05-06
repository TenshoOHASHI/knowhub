ARG GO_VERSION=1.25

FROM golang:${GO_VERSION}-alpine AS builder

ARG SERVICE
WORKDIR /src

RUN apk add --no-cache git

COPY proto ./proto
COPY services/pkg ./services/pkg
COPY services/${SERVICE} ./services/${SERVICE}

WORKDIR /src/services/${SERVICE}
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/service .

FROM alpine:3.22

ARG SERVICE
ENV SERVICE=${SERVICE}

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app/services/${SERVICE}
COPY --from=builder /out/service /usr/local/bin/tenhub-service

EXPOSE 50051 50052 50053 50054 8080
CMD ["tenhub-service"]
