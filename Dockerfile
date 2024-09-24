FROM ghcr.io/theos-dns/auth-nginx:latest
LABEL org.opencontainers.image.source="https://github.com/theos-dns/auth"

WORKDIR /root/app

COPY --from=ghcr.io/theos-dns/auth-api:latest --chmod=777 /root/auth ./auth

ENV allowedIpsFile='/var/nginx/allowed-ips.conf'
ENV nginxConfFile='/etc/nginx/nginx.conf'

CMD ["/bin/sh", "-c", "./nginx-forward-generator -to ${FORWARD_TO:?} -port ${PORTS:?} -nginx-conf-file $nginxConfFile -allowed-ips-file $allowedIpsFile && nginx -g 'daemon off;' | /root/app/auth/auth -db ${DB_PATH:?} -allowed-ips-file $allowedIpsFile"]
