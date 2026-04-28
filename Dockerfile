FROM golang:1.22-alpine AS dev
RUN apk add --no-cache git
WORKDIR /app

FROM golang:1.22-alpine AS build
RUN apk add --no-cache git
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /lrv ./cmd/lrv/

FROM alpine:3.19 AS dist
RUN apk add --no-cache git ca-certificates
COPY --from=build /lrv /usr/local/bin/lrv
ENTRYPOINT ["lrv"]
