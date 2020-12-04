redis {
  addr = "localhost:6379"
  password = ""
  db = 0
  # servicesKey = "services"
  servicesKey = "__gateway_redis__ upstreams"
  # 如果是hash的，servicesKey = "hashKey field"
}

tpl {
  dataSource = "http://127.0.0.1:5003/data"
  #dataSource = "redis"
  # 不设置interval，将只会运行一次
  # 设置了interval，则会进入循环，每次间隔interval时间后，执行一次
  interval = "10s"
  tplSource = "http://127.0.0.1:5003/tpl"
  #tplSource = "testdata/upstreams.tpl"
  #destination = "testdata/upstreams.conf"
  destination = "http://127.0.0.1:5003/post"
  perms = 0600
  command = "echo reloaded"
}

