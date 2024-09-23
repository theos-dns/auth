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
```