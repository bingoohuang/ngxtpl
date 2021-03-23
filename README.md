# ngxtpl

golang nginx template.

1. build: `make install`
1. create a demo config file `ngxtpl --demo`.
1. for http get/post mocking `go get github.com/bingoohuang/httplive/...`, `httplive -d testdata/httplive.db`

![image](https://user-images.githubusercontent.com/1940588/100842805-57feb500-34b4-11eb-9edf-7b3fd6091178.png)

![image](https://user-images.githubusercontent.com/1940588/100977065-68c62e00-357b-11eb-8183-6a65335740a3.png)

## redis

1. edit the created `demo.hcl` config file, see [demo redis confi](testdata/demo.hcl), [demo redis config for set_real_ip_from](testdata/demo_realIps.hcl)
1. put some data into redis.
    - `redis-cli -x hset __gateway_redis__ upstreams  < testdata/upstreams.json`
    - `redis-cli -x set upstreams < testdata/upstreams.json` 
    - `redis-cli -x set realIps  < testdata/set_real_ip_from.json`
1. run `ngxtpl -c demo.hcl`, or `ngxtpl -c demo.hcl -c demo_realIps.hcl`

## mysql

1. start MySQL `docker run -d --name mysql -e MYSQL_USER=user -e MYSQL_DATABASE=mydb -e MYSQL_PASSWORD=pass -e MYSQL_ROOT_PASSWORD=root -p 33306:3306 mysql:5.7`
1. prepare tables  `MYSQL_PWD=root mysql -h hostname -P 33306 -u user mydb < testdata/mysql.sql`
1. prepare config file, see [demo mysql config](testdata/demo_mysql.hcl)

## nacos

[详见](testdata/nacos.md)

## reference

1. [slides HCL: A human-friendly language for developers and operators](https://speakerdeck.com/anubhavmishra/hcl-a-human-friendly-language-for-developers-and-operators)

## docker-compose testing env

### redis install

1. `docker pull redis:6.0.10-alpine3.13`
1. `https://phoenixnap.com/kb/docker-redis` `https://hub.docker.com/_/redis`
1. `docker run --name redis -p 6379:6379 -d redis:6.0.10-alpine3.13` `docker exec -it redis redis-cli`
1. [Another Redis DeskTop Manager](https://github.com/qishibo/AnotherRedisDesktopManager)

## openresty install

1. `https://hub.docker.com/r/openresty/openresty`
1. `docker pull openresty/openresty:1.19.3.1-2-alpine`
1. `docker run --name openresty -v $PWD/testdata/nginx:/etc/nginx/conf.d  -p 8080:8080 -p 8081:8081 -p 8082:8082 -p 8083:8083 -d openresty/openresty:1.19.3.1-2-alpine`

```sh
$ docker-compose exec openresty sh
/ # ps -ef|grep nginx
    1 root      0:00 {openresty} nginx: master process /usr/local/openresty/bin/openresty -g daemon off;
    7 nobody    0:00 {openresty} nginx: worker process
   15 root      0:00 grep nginx
/ # /usr/local/openresty/bin/openresty -t
nginx: the configuration file /usr/local/openresty/nginx/conf/nginx.conf syntax is ok
nginx: configuration file /usr/local/openresty/nginx/conf/nginx.conf test is successful
/ # /usr/local/openresty/bin/openresty -s reload
2021/02/07 09:47:47 [notice] 17#17: signal process started
/ # echo $?
0
/ # 
```

## ngxtpl install

1. `docker build -f Dockerfile -t openresty/ngxtpl .`

## docker compose

1. `docker-compose up`
1. set `tpl` ([demo](testdata/upstreams.tpl)) and `services`([demo](testdata/upstreams.json)) in redis by docker.hcl config.
1. change `services` data in redis and run curl to check results:

```sh
🕙[ 00:13:06 ] ❯ for n in {1..8}; do curl 127.0.0.1:8090/service1/abc;print; done
8081 welcome you!
{"status":"success","result":"8082 nginx json"}
8083 welcome you!
8084 welcome you!
8081 welcome you!
{"status":"success","result":"8082 nginx json"}
8083 welcome you!
8084 welcome you!
```

## daemon service on linux

`/app/ngxtpl install -c /etc/app/docker.hcl; /app/ngxtpl start`
