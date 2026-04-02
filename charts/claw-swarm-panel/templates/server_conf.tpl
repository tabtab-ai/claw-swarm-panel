{{- define "server.conf" -}}
{{ .Values.claw | toYaml }}
{{- end }}

