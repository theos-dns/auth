FROM golang:1.23-alpine AS builder

WORKDIR /root/go/
COPY . .
RUN apk --no-cache add make git gcc libtool musl-dev ca-certificates dumb-init curl
RUN go get .
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o nginx-forward-generator .


FROM nginx:1.27-alpine
LABEL org.opencontainers.image.source="https://github.com/theos-dns/auth"

WORKDIR /root/app

COPY --from=builder --chmod=777 /root/go/nginx-forward-generator ./nginx-forward-generator

COPY templates-conf/nginx.conf /etc/nginx/nginx.conf
COPY templates-conf/default.conf /etc/nginx/conf.d/default.conf
COPY templates-conf/allowed-ips.conf /var/nginx/allowed-ips.conf

CMD ["/bin/sh", "-c",  "./nginx-forward-generator -to ${FORWARD_TO:?} -port ${PORTS:?} -nginx-conf-file /etc/nginx/nginx.conf -allowed-ips-file /var/nginx/allowed-ips.conf && nginx -g 'daemon off;'"]
