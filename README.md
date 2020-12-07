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
