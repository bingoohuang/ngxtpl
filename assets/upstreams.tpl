{{- /* Refer https://danielparker.me/nginx/consul-template/consul/nginx-consul-template/ */ -}}
{{- /* Refer http://nginx.org/en/docs/http/ngx_http_upstream_module.html#upstream  */ -}}
{{- range . -}}
upstream {{.Name}}-pool {
	least_conn;
	keepalive 32;
	{{- range .Servers }}
	server {{.Address}}:{{.Port}}{{if .Weight}} weight={{.Weight}}{{end}}{{if .MaxConns}} max_conns={{.MaxConns}}{{end}}{{if .MaxFails}} max_fails={{.MaxFails}}{{end}}{{if .FailTimeout}} fail_timeout={{.FailTimeout}}{{end}}{{if .Backup}} backup{{end}}{{if .SlowStart}} slow_start={{.SlowStart}}{{end}};{{- end }}
}
{{ end }}