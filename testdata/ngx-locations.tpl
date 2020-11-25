{{- /* Refer https://danielparker.me/nginx/consul-template/consul/nginx-consul-template/ */ -}}
{{- range services -}}
    {{- if in .Tags "nginx-route" -}}
        {{- $boxes := service .Name }}
        {{- if gt (len $boxes) 0 -}}
            upstream {{.Name | replaceAll "--" "-"}}-pool {
            least_conn;
            keepalive 32;
            {{- range service .Name }}
                server {{.Address}}:{{.Port}};{{ end }}
            }
        {{- end -}}
    {{- end -}}
{{- end -}}