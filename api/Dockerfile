FROM golang:1.23-alpine AS builder

WORKDIR /root/go/
COPY . .
RUN apk --no-cache add make git gcc libtool musl-dev ca-certificates dumb-init curl
RUN go get .
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o auth .


FROM alpine:3.17
LABEL org.opencontainers.image.source="https://github.com/theos-dns/auth/tree/master/api"

WORKDIR /root/auth

COPY templates ./templates
COPY --from=builder --chmod=777 /root/go/auth ./auth

