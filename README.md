# ngx-template
golang nginx template 

```bash
 docker run -d --name mysql \
  -e MYSQL_USER=user \
  -e MYSQL_DATABASE=mydb \
  -e MYSQL_PASSWORD=pass \
  -e MYSQL_ROOT_PASSWORD=root \
  -p 3306:3306 \
	mysql:5.7
```

```bash
MYSQL_PWD=root mysql  -h 127.0.0.1 -P 3306 -u root
```

```sql
GRANT ALL PRIVILEGES ON mydb.* TO 'user'@'%';
```

```bash
MYSQL_PWD=pass mysql  -h 127.0.0.1 -P 3306 -u user -D mydb
```

```sql
drop table if exists t_server;
create table t_server(
   id int auto_increment not null comment '自增ID。',
   server varchar(60) not null comment '服务器地址。',
   port varchar(5) not null default "80" comment '端口号: 默认 80',
   state tinyint not null default 1 comment '状态: 0 下线 1 上线',
   primary key (id)
)ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

insert into t_server(server, port) values("192.168.1.1", "8001");
insert into t_server(server, port) values("192.168.1.2", "8001");
update t_server set state = 0 where id = 1;
update t_server set state = 0 ;
```
