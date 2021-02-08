redis {
  addr = "redis:6379"
  password = ""
  db = 0
  servicesKey = "services"
  resultKey = "services.result"
  #servicesKey = "__gateway_redis__ upstreams"
  # 如果是hash的，servicesKey = "hashKey field"
}

tpl {
  # 不设置interval，将只会运行一次
  # 设置了interval，则会进入循环，每次间隔interval时间后，执行一次
  interval = "10s"
  tplSource = "dataSource:tpl"
  destination = "/etc/app/nginx/upstreams.conf"
  perms = 0600
  testCommand = "/usr/local/openresty/bin/openresty -t"
  testCommandCheck = "successful"
  command = "/usr/local/openresty/bin/openresty -s reload"
}

