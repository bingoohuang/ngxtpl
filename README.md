# nginx-template
golang nginx template 

```bash
docker run -d --name mysql \
    -e MYSQL_USER=user \
    -e MYSQL_DATABASE=mydb \
    -e MYSQL_PASSWORD=pass \
    -e MYSQL_ROOT_PASSWORD=root \
    -p 33306:3306 \
    mysql:5.7
```

```bash
MYSQL_PWD=root mysql  -h 127.0.0.1 -P 33306 -u root
```

```sql
GRANT ALL PRIVILEGES ON mydb.* TO 'user'@'%';
```

```bash
MYSQL_PWD=pass mysql  -h 127.0.0.1 -P 33306 -u user -D mydb
```

```sql
drop table if exists t_server;
create table t_server(
   id int auto_increment not null comment '自增ID。',
   address varchar(60) not null comment '服务器地址。',
   port int not null default 80 comment '端口号。',
   weight int not null default 1 comment '权重。',
   max_conns int not null default 0 comment '并发最大连接数。',
   max_fails int not null default 1 comment '并发失败尝试次数。',
   fail_timeout varchar(20) not null default '10s' comment '失败超时。',
   backup tinyint not null default 0 comment '是否备份服务。',
   down tinyint not null default 0 comment '状态: 0 上线 1 下线。',
   slow_start varchar(20) not null default '0' comment '慢启动时间。',
   primary key (id)
)ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

insert into t_server(address, port) values("192.168.1.1", "8001");
insert into t_server(address, port) values("192.168.1.2", "8001");
update t_server set down = 1 where id = 1;
update t_server set down = 1 ;
```

查看系统端口监听

```bash
➜  lsof -i -n -P | grep LISTEN | grep 33306
com.docke 31232 bingoobjca   21u  IPv4 0x2fc7a8888dea524d      0t0  TCP *:33306 (LISTEN)
com.docke 31232 bingoobjca   24u  IPv6 0x2fc7a88898c9320d      0t0  TCP [::1]:33306 (LISTEN)
```

`worker_shutdown_timeout 10s`

1. [worker_shutdown_timeout](http://nginx.org/en/docs/ngx_core_module.html#worker_shutdown_timeout)
1. 旧 worker 进程不能及时退出，就会一直占用着系统资源（CPU、内存和文件描述符等），这对系统资源是一种浪费，因此 nginx/1.11.11 加入了一个新的指令（即 worker_shutdown_timeout，见 Core functionality），允许用户自定义 shutdown 超时时间，如果一个 worker 在接收到退出的指令后经过 worker_shutdown_timeout 时长后还不能退出，就会被强制退出。
1. If the shutdown timer was set, after the worker_shutdown_timeout interval, all connections are closed.
1. [NGINX’s New ‘worker_shutdown_timeout’ Directive](https://medium.com/statuscode/nginxs-new-worker-shutdown-timeout-directive-d60f9c1142f8)
1. [谈谈 nginx 信号集](http://io.upyun.com/2017/08/19/nginx-signals/)
1. [是什么让你的 nginx 服务退出这么慢？](https://zhuanlan.zhihu.com/p/34792840)


Nginx 各种timeout：

1. proxy_connect_timeout 500s; # 后端服务器连接的超时时间_发起握手等候响应超时时间
1. proxy_read_timeout 500s; # 连接成功后_等候后端服务器响应时间_其实已经进入后端的排队之中等候处理（也可以说是后端服务器处理请求的时间）
1. proxy_send_timeout 500s; # 后端服务器数据回传时间_就是在规定时间之内后端服务器必须传完所有的数据
1. keepalive_timeout 60s; # 指定每个 TCP 连接最多可以保持多长时间。Nginx 的默认值是 75 秒，有些浏览器最多只保持 60 秒，所以可以设定为 60 秒。若将它设置为 0，就禁止了 keepalive 连接。
1. client_body_timeout 20s; # 指定客户端与服务端建立连接后发送 request body 的超时时间。如果客户端在指定时间内没有发送任何内容，Nginx 返回 HTTP 408（Request Timed Out）。
1. client_header_timeout 10s; # 客户端向服务端发送一个完整的 request header 的超时时间。如果客户端在指定时间内没有发送一个完整的 request header，Nginx 返回 HTTP 408（Request Timed Out）。
1. send_timeout 60s; # 发送数据至客户端超时, 默认60s, 如果连续的60s内客户端没有收到1个字节, 连接关闭。
1. client_header_timeout 60s; # 接收客户端header超时, 默认60s, 如果60s内没有收到完整的http包头, 返回408。
1. client_body_timeout 60s; # 接收客户端body超时, 默认60s, 如果连续的60s内没有收到客户端的1个字节, 返回408。
1. lingering_timeout 5s; # TCP连接关闭时的SO_LINGER延时设置，默认5s。
1. resolver_timeout 5s; # 域名解析超时，默认30s。
1. [Nginx的超时timeout配置详解](https://juejin.im/post/5b696a24e51d45191e0d3e56)

                      
