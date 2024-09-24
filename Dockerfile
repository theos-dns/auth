FROM foo-nginx

WORKDIR /root/app

COPY --from=foo-api --chmod=777 /root/auth ./auth

ENV allowedIpsFile='/var/nginx/allowed-ips.conf'
ENV nginxConfFile='/etc/nginx/nginx.conf'

CMD ["/bin/sh", "-c", "./nginx-forward-generator -to ${FORWARD_TO:?} -port ${PORTS:?} -nginx-conf-file $nginxConfFile -allowed-ips-file $allowedIpsFile && nginx -g 'daemon off;' | /root/app/auth/auth -db ${DB_PATH:?} -allowed-ips-file $allowedIpsFile"]
