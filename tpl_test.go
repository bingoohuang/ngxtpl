package ngxtpl_test

import (
	"bytes"
	"html/template"
	"io/ioutil"
	"testing"

	"github.com/bingoohuang/ngxtpl"
	"github.com/stretchr/testify/assert"
)

func TestUpstreamsTemplate(t *testing.T) {
	tmpl := template.New("upstreams")
	tmplText, err := ioutil.ReadFile("assets/upstreams.tpl")
	assert.Nil(t, err)

	tmpl, err = tmpl.Parse(string(tmplText))
	assert.Nil(t, err)

	var out bytes.Buffer
	upstream1 := ngxtpl.Upstream{
		Name: "service1",
		Servers: []ngxtpl.Server{
			{
				Address: "127.0.0.1",
				Port:    8001,
			}, {
				Address: "127.0.0.1",
				Port:    8002,
			},
		},
	}
	assert.Nil(t, tmpl.Execute(&out, []ngxtpl.Upstream{upstream1}))
	assert.Equal(t,
		`upstream service1-pool {
	least_conn;
	keepalive 32;
	server 127.0.0.1:8001;
	server 127.0.0.1:8002;
}
`, out.String())

	s3 := ngxtpl.Server{
		Address:     "192.168.1.1",
		Port:        80,
		Weight:      10,
		MaxConns:    10,
		MaxFails:    10,
		FailTimeout: "10s",
		Backup:      true,
		SlowStart:   "30s",
	}

	upstream2 := ngxtpl.Upstream{
		Name: "service2",
		Servers: []ngxtpl.Server{
			{
				Address: "127.0.0.1",
				Port:    8201,
			},
			{
				Address: "127.0.0.1",
				Port:    8202,
			},
			s3,
		},
	}
	out.Reset()
	assert.Nil(t, tmpl.Execute(&out, []ngxtpl.Upstream{upstream1, upstream2}))
	assert.Equal(t,
		`upstream service1-pool {
	least_conn;
	keepalive 32;
	server 127.0.0.1:8001;
	server 127.0.0.1:8002;
}
upstream service2-pool {
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
	tmplText, err := ioutil.ReadFile("assets/locations.tpl")
	assert.Nil(t, err)

	tmpl, err = tmpl.Parse(string(tmplText))
	assert.Nil(t, err)

	var out bytes.Buffer
	upstream1 := ngxtpl.Upstream{
		Name: "service1",
	}
	assert.Nil(t, tmpl.Execute(&out, []ngxtpl.Upstream{upstream1}))
	assert.Equal(t,
		`location /service1 {
	proxy_pass https://service1-pool;
	proxy_http_version 1.1;
	proxy_set_header Connection "";
}
`, out.String())

	upstream2 := ngxtpl.Upstream{
		Name: "service2",
	}
	out.Reset()
	assert.Nil(t, tmpl.Execute(&out, []ngxtpl.Upstream{upstream1, upstream2}))
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
