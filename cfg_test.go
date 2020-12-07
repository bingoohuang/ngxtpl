package ngxtpl_test

import (
	"testing"

	"github.com/bingoohuang/ngxtpl"
	"github.com/hashicorp/hcl"
	"github.com/stretchr/testify/assert"
)

func TestCfgParse(t *testing.T) {
	s, err := ngxtpl.ReadFileE("assets/cfg.hcl")
	assert.Nil(t, err)

	var cfg ngxtpl.Cfg

	assert.Nil(t, hcl.Unmarshal(s, &cfg))
	assert.Equal(t, ngxtpl.Cfg{
		Nacos: ngxtpl.Nacos{
			ClientConfig: ngxtpl.ClientConfig{
				TimeoutMs:            5000,
				BeatInterval:         0,
				NamespaceID:          "f3c0ab89-31bb-4414-a495-146941316751",
				Endpoint:             "",
				RegionID:             "",
				AccessKey:            "",
				SecretKey:            "",
				OpenKMS:              false,
				CacheDir:             "/tmp/nacos/cache",
				UpdateThreadNum:      0,
				NotLoadCacheAtStart:  true,
				UpdateCacheWhenEmpty: false,
				Username:             "",
				Password:             "",
				LogDir:               "/tmp/nacos/log",
				RotateTime:           "1h",
				MaxAge:               3,
				LogLevel:             "debug",
			},
			ServerConfigs: []ngxtpl.ServerConfig{
				{
					Scheme:      "http",
					ContextPath: "/nacos",
					IPAddr:      "127.0.0.1",
					Port:        8848,
				}, {
					Scheme:      "http",
					ContextPath: "/nacos",
					IPAddr:      "127.0.0.1",
					Port:        8849,
				},
			},
			ServiceParam: ngxtpl.ServiceParam{
				Clusters:    []string{"clustera"},
				ServiceName: "demogo",
				GroupName:   "groupa",
			},
		},
		Mysql: ngxtpl.Mysql{
			DataSourceName: "user:pass@tcp(127.0.0.1:3306)/db1?charset=utf8",
			DataKey:        "upstreams",
			DataSQL: "select name,keepalive,ip_hash ipHash,resolver,'{{servers}}' servers " +
				"from t_upstreams where state='1'",
			Sqls: map[string]string{
				"servers": "select address,port,weight,max_conns maxConns,max_fails maxFails," +
					"fail_timeout failTimeout,backup,down,slow_start slowStart " +
					"from t_servers where upstream_name='{{.name}}' and state='1'",
			},
			KVSql: "select value from t_config where key = '{{key}}'",
		},
		Redis: ngxtpl.Redis{
			Addr:        "localhost:6379",
			Password:    "",
			Db:          0,
			ServicesKey: "services",
		},
		Tpl: ngxtpl.Tpl{
			DataSource:  "redis",
			Interval:    "10s",
			TplSource:   "/etc/nginx/conf.d/load-balancer.conf.tpl",
			Destination: "/etc/nginx/conf.d/load-balancer.conf",
			Perms:       0600,
			Command:     "service nginx reload",
		},
	}, cfg)
}
