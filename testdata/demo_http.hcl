tpl {
  dataSource = "http://127.0.0.1:5003/data"
  # 不设置interval，将只会运行一次
  # 设置了interval，则会进入循环，每次间隔interval时间后，执行一次
  interval = "20s"
  tplSource = "http://127.0.0.1:5003/tpl"
  destination = "http://127.0.0.1:5003/post"
  perms = 0600
  command = "echo reloaded"
}

