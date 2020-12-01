{{- /* Refer https://danielparker.me/nginx/consul-template/consul/nginx-consul-template/ */ -}}
{{- /* Refer http://nginx.org/en/docs/http/ngx_http_upstream_module.html#upstream  */ -}}
{{- range .upstreams -}}
upstream {{.name}}-pool {
	least_conn;
	keepalive 32;
	{{- range .servers }}
	server {{.address}}:{{.port}}{{if .weight}} weight={{.weight}}{{end}}{{if .maxConns}} max_conns={{.maxConns}}{{end}}{{if .maxFails}} max_fails={{.maxFails}}{{end}}{{if .failTimeout}} fail_timeout={{.failTimeout}}{{end}}{{if .backup}} backup{{end}}{{if .slowStart}} slow_start={{.slowStart}}{{end}};{{- end }}
}
{{ end }}