{{- range .realIps -}}{{- if eq (or .state "1") "1" -}}
    set_real_ip_from {{.ip}};
{{ end }}{{ end }}