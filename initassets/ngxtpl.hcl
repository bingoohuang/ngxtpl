redis {
  addr = "localhost:6379"
  # password = ""
  # db = 0
  servicesKey = "services"
  # servicesKey = "hashKey element"
}

tpl {
  interval = "10s"
  tplSource = "dataSource:tpl"
  destination = "/etc/nginx/conf.d/load-balancer.conf"
  perms = 0600
  testCommand = "/usr/local/openresty/bin/openresty -t"
  testCommandCheck = "successful"
  command = "/usr/local/openresty/bin/openresty -s reload"
}

# https://learn.hashicorp.com/tutorials/consul/load-balancing-nginx
# https://github.com/hashicorp/hcl
# https://www.convertsimple.com/convert-hcl-to-json/

//nacos {
//  clientConfig {
//    configFile = "/path/to/nacos.yaml"
//  }
//
//  serverConfigs = [
//    {
//      Scheme = "http"
//      IpAddr = "127.0.0.1"
//      Port = 8848
//      ContextPath = "/nacos"
//    },
//    {
//      Scheme = "http"
//      IpAddr = "127.0.0.1"
//      Port = 8849
//      ContextPath = "/nacos"
//    }
//  ]
//
//  serviceParam {
//    ServiceName = "demogo",
//    Clusters = [
//      "clustera"]
//    // default value is DEFAULT
//    GroupName = "groupa"
//    // default value is DEFAULT_GROUP
//  }
//}

//mysql {
//  dataSourceName = "user:pass@tcp(127.0.0.1:3306)/db1?charset=utf8"
//  dataKey = "upstreams"
//  dataSql = "select name,keepalive,ip_hash ipHash,resolver,'{{servers}}' servers from t_upstreams where state='1'"
//  sqls {
//    servers = "select address,port,weight,max_conns maxConns,max_fails maxFails,fail_timeout failTimeout,backup,down,slow_start slowStart from t_servers where upstream_name='{{.name}}' and state='1'"
//  }
//
//  kvSql = "select value from t_config where key = '{{key}}'"
//}
