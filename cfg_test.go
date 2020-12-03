package ngxtpl_test

import (
	"io/ioutil"
	"path/filepath"
	"testing"

	"github.com/bingoohuang/ngxtpl"
	"github.com/hashicorp/hcl"
	"github.com/stretchr/testify/assert"
)

func TestCfgParse(t *testing.T) {
	var cfg ngxtpl.Cfg

	assert.Nil(t, hcl.Decode(&cfg, testReadFile(t, "cfg.hcl")))
	assert.Equal(t, ngxtpl.Cfg{
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

func testReadFile(t *testing.T, n string) string {
	d, err := ioutil.ReadFile(filepath.Join("assets", n))
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	return string(d)
}
