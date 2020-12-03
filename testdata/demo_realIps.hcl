redis {
  addr = "localhost:6379"
  password = ""
  db = 0
  # servicesKey = "services"
  servicesKey = "realIps"
  # 如果是hash的，servicesKey = "hashKey field"
}

tpl {
  dataSource = "redis"
  # 不设置interval，将只会运行一次
  # 设置了interval，则会进入循环，每次间隔interval时间后，执行一次
  interval = "10s"
  tplSource = "testdata/set_real_ip_from.tpl"
  destination = "testdata/set_real_ip_from.conf"
  perms = 0600
  command = "echo reloaded"
}

