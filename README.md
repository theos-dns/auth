# theos dns auth

## image ENVs
- `FORWARD_TO` hostname or ip which requests should be forwarded
- `PORTS` ports that nginx listen to, seperated by ',' like: 80,443,1080 also can be range like 8080-8090, or combination of both
- `DB_PATH` sqlLite database file path witch should be saved
- `UPSTREAM` upstream server witch should get new authorized ip. seperated by ','
- `PROTECT` other services that should be protected. Seperated by ','. Structure: `{SERVICE_OR_IP}:{SOURCE_PORT}@{DESTINATION_PORT}`
- `STARTUP_SLEEP` seconds to sleep before starting nginx
- `RESOLVER` dns server that resolves protected-services and forward-to hosts


## image Volumes
- `/root/app/auth/db/` which is the default location for `sqllite` db file
- `/var/nginx/allowed-ips.conf` which saves allowed ips

## image
`docker pull ghcr.io/theos-dns/auth:latest`


## todo:
- [x] call upstreams
- [x] protect other services
- [ ] register user
- - [ ] api
- - [ ] interface
- [ ] register user in upstream
- [ ] replace previous allowed ips
