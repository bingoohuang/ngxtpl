package ngx_template

// Server struct by http://nginx.org/en/docs/http/ngx_http_upstream_module.html
type Server struct {
	Host   string `json:"host"` // eg. backend1.example.com, backend2.example.com:8080, 192.0.2.1
	Port   int    `json:"port"`
	Weight int    `json:"weight"` // eg. weight=5， sets the weight of the server, by default, 1.
	// max_conns=number
	// limits the maximum number of simultaneous active connections to the proxied server (1.11.5).
	// Default value is zero, meaning there is no limit.
	MaxConns int `json:"max_conns"`
}

type Upstream struct {
	Name    string   `json:"name"`
	Servers []Server `json:"servers"`
}
