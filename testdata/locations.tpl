{{- /* Refer https://danielparker.me/nginx/consul-template/consul/nginx-consul-template/ */ -}}
{{- range .upstreams -}}
location /{{.name}} {
	proxy_pass https://{{.name}}-pool;
	proxy_http_version 1.1;
	proxy_set_header Connection "";
}
{{ end }}