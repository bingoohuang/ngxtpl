{{- /* Refer https://danielparker.me/nginx/consul-template/consul/nginx-consul-template/ */ -}}
{{- range . -}}
location /{{.Name}} {
	proxy_pass https://{{.Name}}-pool;
	proxy_http_version 1.1;
	proxy_set_header Connection "";
}
{{ end }}