nacos {
  clientConfig {
    NamespaceId = "f3c0ab89-31bb-4414-a495-146941316751"
    TimeoutMs = 5000
    NotLoadCacheAtStart = true
    LogDir = "/tmp/nacos/log"
    CacheDir = "/tmp/nacos/cache"
    RotateTime = "1h"
    MaxAge = 3
    LogLevel = "debug"
  }

  serverConfigs = [
    {
      Scheme = "http"
      IpAddr = "127.0.0.1"
      Port = 8848
      ContextPath = "/nacos"
    }
  ]

  serviceParam {
    ServiceName = "demogo",
    Clusters = ["clustera"]
    // default value is DEFAULT
    GroupName = "groupa"
    // default value is DEFAULT_GROUP
  }
}

tpl {
  dataSource = "nacos"
  # 不设置interval，将只会运行一次
  # 设置了interval，则会进入循环，每次间隔interval时间后，执行一次
  interval = "10s"
  #tplSource = "http://127.0.0.1:5003/tpl"
  tplSource = "testdata/upstreams.tpl"
  destination = "testdata/upstreams_nacos.conf"
//  destination = "http://127.0.0.1:5003/post"
  perms = 0600
  command = "echo reloaded"
}

