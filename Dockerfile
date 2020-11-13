# Build
FROM golang:1.13-alpine as builder

ENV CGO_ENABLED 0
ENV GOOS linux
ARG TARGET_APP=./cmd/node

RUN apk update && apk add --no-cache git

RUN mkdir /app
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -a -installsuffix cgo -o main $TARGET_APP

# Runtime
FROM alpine:latest as runtime

RUN apk --no-cache add ca-certificates

COPY --from=builder /app/main /app/main
WORKDIR /app

ENTRYPOINT ["./main"]