package ngxtpl_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

type Data struct {
	Upstreams []Upstream `json:"upstreams"`
}

// Upstream defines the data structure for nginx upstream.
type Upstream struct {
	Name      string   `json:"name"`
	State     string   `json:"state"`
	Keepalive string   `json:"keepalive"`
	Servers   []Server `json:"servers"`
}

// Server struct by http://nginx.org/en/docs/http/ngx_http_upstream_module.html
type Server struct {
	Address     string `json:"address"`
	Port        string `json:"port"`
	Weight      string `json:"weight"`      // eg. weight=5， sets the weight of the server.
	MaxConns    string `json:"maxConns"`    // DefaultTo value is zero, meaning there is no limit.
	MaxFails    string `json:"maxFails"`    // By default, the number of unsuccessful attempts is set to 1
	FailTimeout string `json:"failTimeout"` // By default, the parameter is set to 10 seconds.
	Backup      string `json:"backup"`
	SlowStart   string `json:"slowStart"` // DefaultTo value is zero, i.e. slow start is disabled.
}

func TestMapArg(t *testing.T) {
	m := make(map[string]string)

	assert.Equal(t, map[string]string{}, m)

	f := func(a map[string]string) {
		a["address"] = "zuoan"
	}

	f(m)

	assert.Equal(t, map[string]string{"address": "zuoan"}, m)

	mm := []map[string]string{
		{"name": "bingoo"},
		{"name": "huang"},
	}

	assert.Equal(t, []map[string]string{
		{"name": "bingoo"},
		{"name": "huang"},
	}, mm)

	for _, i := range mm {
		f(i)
	}

	assert.Equal(t, []map[string]string{
		{"address": "zuoan", "name": "bingoo"},
		{"address": "zuoan", "name": "huang"},
	}, mm)
}

func TestUpstreamsTemplate(t *testing.T) {
	tmpl := template.New("upstreams")
	tmplText, err := os.ReadFile("testdata/test_upstreams.tpl")
	assert.Nil(t, err)

	tmpl, err = tmpl.Parse(string(tmplText))
	assert.Nil(t, err)

	var out bytes.Buffer
	data := Data{
		Upstreams: []Upstream{{
			Name:      "service1",
			State:     "1",
			Keepalive: "32",
			Servers: []Server{
				{
					Address: "127.0.0.1",
					Port:    "8001",
				}, {
					Address: "127.0.0.1",
					Port:    "8002",
				},
			},
		}},
	}

	err = tmpl.Execute(&out, ToMap(data))
	if err != nil {
		fmt.Printf("err %v\n", err)
	}
	assert.Nil(t, err)
	assert.Equal(t,
		`upstream service1 {
	least_conn;
	keepalive 32;
	server 127.0.0.1:8001;
	server 127.0.0.1:8002;
}

`, out.String())

	s3 := Server{
		Address:     "192.168.1.1",
		Port:        "80",
		Weight:      "10",
		MaxConns:    "10",
		MaxFails:    "10",
		FailTimeout: "10s",
		Backup:      "yes",
		SlowStart:   "30s",
	}

	data = Data{
		Upstreams: []Upstream{
			{
				Name:      "service1",
				State:     "1",
				Keepalive: "32",
				Servers: []Server{
					{
						Address: "127.0.0.1",
						Port:    "8001",
					}, {
						Address: "127.0.0.1",
						Port:    "8002",
					},
				},
			}, {
				Name:      "service2",
				State:     "1",
				Keepalive: "32",
				Servers: []Server{
					{
						Address: "127.0.0.1",
						Port:    "8201",
					},
					{
						Address: "127.0.0.1",
						Port:    "8202",
					},
					s3,
				},
			}, {
				Name:      "service3",
				State:     "0",
				Keepalive: "32",
				Servers: []Server{
					{
						Address: "127.0.0.1",
						Port:    "9201",
					},
					{
						Address: "127.0.0.1",
						Port:    "9202",
					},
					s3,
				},
			},
		},
	}

	out.Reset()
	assert.Nil(t, tmpl.Execute(&out, ToMap(data)))
	assert.Equal(t,
		`upstream service1 {
	least_conn;
	keepalive 32;
	server 127.0.0.1:8001;
	server 127.0.0.1:8002;
}
upstream service2 {
	least_conn;
	keepalive 32;
	server 127.0.0.1:8201;
	server 127.0.0.1:8202;
	server 192.168.1.1:80 weight=10 max_conns=10 max_fails=10 fail_timeout=10s backup slow_start=30s;
}

`, out.String())
}

func TestLocationsTemplate(t *testing.T) {
	tmpl := template.New("locations")
	tmplText, err := os.ReadFile("testdata/locations.tpl")
	assert.Nil(t, err)

	tmpl, err = tmpl.Parse(string(tmplText))
	assert.Nil(t, err)

	var out bytes.Buffer
	data := Data{Upstreams: []Upstream{
		{
			Name: "service1",
		},
	}}
	assert.Nil(t, tmpl.Execute(&out, ToMap(data)))
	assert.Equal(t,
		`location /service1 {
	proxy_pass https://service1-pool;
	proxy_http_version 1.1;
	proxy_set_header Connection "";
}
`, out.String())

	data = Data{Upstreams: []Upstream{
		{
			Name: "service1",
		}, {
			Name: "service2",
		},
	}}

	out.Reset()
	assert.Nil(t, tmpl.Execute(&out, ToMap(data)))
	assert.Equal(t,
		`location /service1 {
	proxy_pass https://service1-pool;
	proxy_http_version 1.1;
	proxy_set_header Connection "";
}
location /service2 {
	proxy_pass https://service2-pool;
	proxy_http_version 1.1;
	proxy_set_header Connection "";
}
`, out.String())
}

func ToMap(s interface{}) (m map[string]interface{}) {
	v, _ := json.Marshal(s)
	_ = json.Unmarshal(v, &m)
	return
}
