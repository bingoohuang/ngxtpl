{{- /* Refer https://danielparker.me/nginx/consul-template/consul/nginx-consul-template/ */ -}}
{{- /* Refer http://nginx.org/en/docs/http/ngx_http_upstream_module.html#upstream  */ -}}
{{- range .upstreams }}{{ if eq (or .state "1") "1" -}}
upstream {{.name}} {
	least_conn;
	{{if .keepalive }}keepalive {{.keepalive}};{{end}}
	{{- range .servers }}{{ if eq (or .state "1") "1" }}
	server {{.address}}:{{.port}}
	{{- if .weight }} weight={{.weight}}{{end}}
	{{- if .maxConns }} max_conns={{.maxConns}}{{end}}
	{{- if .maxFails }} max_fails={{.maxFails}}{{end}}
	{{- if .failTimeout }} fail_timeout={{.failTimeout}}{{end}}
	{{- if .backup}}{{ if eq .backup "yes" }} backup{{end}}{{end}}
	{{- if .slowStart}} slow_start={{.slowStart}}{{end}};
	{{- end }}
	{{- end }}
}

# https://github.com/openresty/lua-resty-upstream-healthcheck
init_worker_by_lua_block {
    local hc = require "resty.upstream.healthcheck"
    local ok, err = hc.spawn_checker{
        shm = "healthcheck",  -- defined by "lua_shared_dict"
        upstream = "{{.name}}", -- defined by "upstream"
        type = "http",
        http_req = "GET /status HTTP/1.0\r\nHost: foo.com\r\n\r\n",
                -- raw HTTP request for checking
        interval = 2000,  -- run the check cycle every 2 sec
        timeout = 1000,   -- 1 sec is the timeout for network operations
        fall = 3,  -- # of successive failures before turning a peer down
        rise = 2,  -- # of successive successes before turning a peer up
        valid_statuses = {200, 302},  -- a list valid HTTP status code
        concurrency = 10,  -- concurrency level for test requests
    }
    if not ok then
        ngx.log(ngx.ERR, "failed to spawn health checker: ", err)
        return
    end
}
{{ end }}{{ end }}
