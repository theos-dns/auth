# theos dns auth api

this file is not for standalone uses. combine it with [nginx](../nginx) docker file


## usage
```
Usage of theos_dns_auth_api:
  -allowed-ips-file string
        nginx allowed ips file path (default "/var/nginx/allowed-ips.conf")
  -db string
        sqlLite database path (default "./db/database-auth.sqlite3")
  -help
        Display help message
  -host string
        web server host running on (default "0.0.0.0")
  -port string
        web server port running on (default "82")
```


## api

### Register ip: 
```
/tap-in?ip=127.0.0.1&token=123456789
```

### interface to register ip:
```
/
```