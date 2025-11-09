# Theos DNS Auth

This stack manages access to server ports and services. It can protect ports (e.g., port 80) or proxy-pass a service. For example, it can protect port 53 while redirecting its traffic to the corresponding service, such as a DNS server.

## Features
- Protects specific ports and services.
- Redirects traffic to designated services.
- Manages allowed IPs for access control.
- Supports upstream communication for authorized IP updates.
- Provides an API for user registration and management.

## Image ENVs
- `FORWARD_TO`: Hostname or IP to which requests should be forwarded. The port remains unchanged.
- `PORTS`: Ports that Nginx listens to, separated by `,` (e.g., `80,443,1080`) or ranges (e.g., `8080-8090`), or a combination of both.
- `DB_PATH`: SQLite database file path where data should be saved.
- `UPSTREAM`: Upstream server(s) to receive new authorized IPs, separated by `,`.
- `PROTECT`: Other services to protect, separated by `,`. Structure: `{SERVICE_OR_IP}:{SOURCE_PORT}@{DESTINATION_PORT}`.
- `STARTUP_SLEEP`: Number of seconds to sleep before starting Nginx.
- `RESOLVER`: DNS server that resolves protected services and forward-to hosts.
- `ADMIN_TOKEN`: Admin token used for creating users and managing upstreams. Must match the upstream `ADMIN_TOKEN`.

### Note
The combination of `FORWARD_TO` and `PORTS` works similarly to `PROTECT`, allowing traffic redirection and port protection.

## Image Volumes
- `/root/app/auth/db/`: Default location for the SQLite database file.
- `/var/nginx/allowed-ips.conf`: File that stores allowed IPs.


## todo:
- [x] call upstreams
- [x] protect other services
- [x] replace previous allowed ips
- [ ] register user
- - [x] api
- - [ ] interface
- [x] register user in upstream
