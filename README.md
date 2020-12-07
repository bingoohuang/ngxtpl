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

> 实现备注: 暂时还是轮询方案，没有实现订阅模式.

1. 启动服务器
   - 参考 [Nacos 快速开始](https://nacos.io/zh-cn/docs/quick-start.html)
   - git clone https://github.com/alibaba/nacos.git
   - cd nacos/
   - mvn -Prelease-nacos -Dmaven.test.skip=true clean install -U 
   - ls -al distribution/target/
   - cd distribution/target/nacos-server-$version/nacos/bin
   - 启动命令(standalone代表着单机模式运行，非集群模式): `sh startup.sh -m standalone`
1. 服务注册&发现和配置管理
   - 服务注册`curl -X POST "http://127.0.0.1:8848/nacos/v1/ns/instance?&namespaceId=f3c0ab89-31bb-4414-a495-146941316751&serviceName=groupa@@demogo&clusterName=clustera&ip=127.0.0.1&port=8811&ephemeral=false"`
   - 服务注册`curl -X POST "http://127.0.0.1:8848/nacos/v1/ns/instance?&namespaceId=f3c0ab89-31bb-4414-a495-146941316751&serviceName=groupa@@demogo&clusterName=clustera&ip=127.0.0.1&port=8812&ephemeral=false"`
   - 服务发现`curl -X GET  "http://127.0.0.1:8848/nacos/v1/ns/instance/list?namespaceId=f3c0ab89-31bb-4414-a495-146941316751&serviceName=groupa@@demogo"`
   - [json url encode](https://onlinejsontools.com/url-encode-json)  
   - 配置公共参数 `curl -X POST "http://127.0.0.1:8848/nacos/v1/cs/configs?tenant=f3c0ab89-31bb-4414-a495-146941316751&dataId=demogo&group=upstreams&content=%7B%22keepalive%22%3A13%7D&type=JSON"`
   - 查看公共参数 `curl -X GET  "http://127.0.0.1:8848/nacos/v1/cs/configs?tenant=f3c0ab89-31bb-4414-a495-146941316751&dataId=demogo&group=upstreams"`
1. 关闭服务器 
   - `sh shutdown.sh`
1. 问题
   - [code:503,msg:server is DOWN now, please try again later!](https://github.com/alibaba/nacos/issues/4210), 删除`{nacos.home}/nacos/data/protocol`文件夹后重新启动。
1. reference
   - [nacos-sdk-go](https://github.com/nacos-group/nacos-sdk-go)
   - [nacos open-api](https://nacos.io/zh-cn/docs/open-api.html)

## reference

1. [slides HCL: A human-friendly language for developers and operators](https://speakerdeck.com/anubhavmishra/hcl-a-human-friendly-language-for-developers-and-operators)
