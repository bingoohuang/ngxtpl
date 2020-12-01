# https://learn.hashicorp.com/tutorials/consul/load-balancing-nginx
# https://github.com/hashicorp/hcl

mysql {
  dataSourceName = "user:pass@tcp(127.0.0.1:3306)/db1?charset=utf8"
  upstreamsTable = "t_upstreams"
  serversTable = "t_servers"
}

redis {
  addr = "localhost:6379"
  password = ""
  db = 0
  # servicesKey = "services"
  servicesKey = "__gateway_redis__ upstreams"
  # 如果是hash的，servicesKey = "hashKey field"
}

tpl {
  dataSource = "redis"
  #interval = "10s"
  source = "testdata/upstreams.tpl"
  destination = "testdata/upstreams.conf"
  perms = 0600
  command = "echo reloaded"
}

