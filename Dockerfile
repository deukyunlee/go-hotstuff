# Build stage for amd64
FROM golang:alpine AS builder-amd64

ENV GO111MODULE=on \
    CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64

WORKDIR /build
COPY . .

RUN go mod download
RUN go build -o main .

# Build stage for arm64
FROM golang:alpine AS builder-arm64

ENV GO111MODULE=on \
    CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=arm64

WORKDIR /build
COPY . .

RUN go mod download
RUN go build -o main .

FROM scratch

COPY --from=builder-amd64 /build/main /amd64/main
COPY --from=builder-arm64 /build/main /arm64/main

ENTRYPOINT ["/arm64/main"]
