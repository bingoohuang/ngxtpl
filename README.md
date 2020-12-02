# ngxtpl

golang nginx template.

1. build: `make install`
1. create a demo config file `ngxtpl --demo`.

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

## reference

1. [nacos-sdk-go](https://github.com/nacos-group/nacos-sdk-go)
1. [slides HCL: A human-friendly language for developers and operators](https://speakerdeck.com/anubhavmishra/hcl-a-human-friendly-language-for-developers-and-operators)
