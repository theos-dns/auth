# theos dns auth api

this file is not for standalone uses. combine it with [nginx](../nginx) docker file


## usage
```
Usage of theos_dns_auth_api:
  -admin-token string
        admin token which will be used to create users and all upstreams
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
  -upstream string
        upstream server witch should get new authorized ip. seperated by ','
```


## api

### Register ip: 
```
/tap-in?ip=192.161.1.5&token=USER_TOKEN
```

### Register user: 
```
/register-user?ip=192.161.1.5&token=USER_TOKEN&adminToken=ADMIN_TOKEN&username=USERNANE&limitation=2
```
register new user or if exist allow its ip(using for upstream call)
### interface to register ip:
```
/
```