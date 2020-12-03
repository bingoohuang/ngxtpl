# https://learn.hashicorp.com/tutorials/consul/load-balancing-nginx
# https://github.com/hashicorp/hcl

mysql {
  dataSourceName = "user:pass@tcp(127.0.0.1:3306)/db1?charset=utf8"
  dataKey = "upstreams"
  dataSql = "select name,keepalive,ip_hash ipHash,resolver,'{{servers}}' servers from t_upstreams where state='1'"
  sqls {
    servers = "select address,port,weight,max_conns maxConns,max_fails maxFails,fail_timeout failTimeout,backup,down,slow_start slowStart from t_servers where upstream_name='{{.name}}' and state='1'"
  }
}

redis {
  addr = "localhost:6379"
  password = ""
  db = 0
  servicesKey = "services"
  # servicesKey = "__gateway_redis__ upstreams"
  # 如果是hash的，servicesKey = "hashKey field"
}

tpl {
  dataSource = "redis"
  interval = "10s"
  tplSource = "/etc/nginx/conf.d/load-balancer.conf.tpl"
  destination = "/etc/nginx/conf.d/load-balancer.conf"
  perms = 0600
  command = "service nginx reload"
}

