ARG VERSION_TO_GET

FROM ghcr.io/theos-dns/auth-api:$VERSION_TO_GET AS api-auth

FROM ghcr.io/theos-dns/auth-nginx:$VERSION_TO_GET
LABEL org.opencontainers.image.source="https://github.com/theos-dns/auth"

WORKDIR /root/app

COPY --from=api-auth --chmod=777 /root/auth ./auth

ENV allowedIpsFile='/var/nginx/allowed-ips.conf'
ENV nginxConfFile='/etc/nginx/nginx.conf'

CMD ["/bin/sh", "-c", "./nginx-forward-generator -to ${FORWARD_TO:?} -port ${PORTS:?} -protect ${PROTECT:-''} -nginx-conf-file $nginxConfFile -allowed-ips-file $allowedIpsFile && nginx -g 'daemon off;' | /root/app/auth/auth -db ${DB_PATH:?} -allowed-ips-file $allowedIpsFile -upstream ${UPSTREAM:-''}"]
