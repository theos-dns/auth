# theos dns auth

## image ENVs
`FORWARD_TO` hostname or ip which requests should be forwarded

`PORTS` ports that nginx listen to, seperated by ',' like: 80,443,1080 also can be range like 8080-8090, or combination of both

`DB_PATH` sqlLite database file path witch should be saved


## image Volumes
`/root/app/auth/db/` witch is de default location on sqlLite db file
`/var/nginx/allowed-ips.conf` witch saves allowed ips