# theos dns auth nginx
will block all incoming request on defined ports except IPs that are allowed in `/var/nginx/allowed-ips.conf`.
Allowed ips file will be updated by `api` app

forward all incoming requests (which was checked that are authorized by `/var/nginx/allowed-ips.conf`) to defined server on incoming port

Also on port `81` it will return the ip address of client

## image ENVs
- `FORWARD_TO` hostname or ip which requests should be forwarded
- `PORTS` ports listen to, seperated by ',' like: 80,443,1080 also can be range like 8080-8090, or combination of both
- `PROTECT` other services that should be protected. Seperated by ','. Structure: {SERVICE_OR_IP}:{SOURCE_PORT}@{DESTINATION_PORT}
- `STARTUP_SLEEP` seconds to sleep before starting nginx
- `RESOLVER` dns server that resolves protected-services and forward-to hosts

## usage of `nginx-forward-generator`

```
Usage nginx-forward-generator:
  -allowed-ips-file string
        host listen to (default "/var/nginx/allowed-ips.conf")
  -help
        Display help message
  -nginx-conf-file string
        host listen to (default "/etc/nginx/nginx.conf")
  -port string
        port listen to, seperated by ',' like: 80,443,1080 also can be range like 8080-8090, or combination of both (default "443,80")
  -to string
        host address where authorized requests froward to (port will not be changed!)
  -protect string
        other services that should be protected. Seperated by ','. Structure: {SERVICE_OR_IP}:{SOURCE_PORT}@{DESTINATION_PORT} (default "dns-server:53@53,coap:85@5688")
```