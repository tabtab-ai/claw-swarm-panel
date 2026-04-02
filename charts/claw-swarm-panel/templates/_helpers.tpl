{{/*
Expand the name of the chart.
*/}}
{{- define "claw-swarm-panel.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
*/}}
{{- define "claw-swarm-panel.fullname" -}}
{{- if .Values.fullnameOverride }}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- $name := default .Chart.Name .Values.nameOverride }}
{{- if contains $name .Release.Name }}
{{- .Release.Name | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" }}
{{- end }}
{{- end }}
{{- end }}

{{/*
Chart label.
*/}}
{{- define "claw-swarm-panel.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels.
*/}}
{{- define "claw-swarm-panel.labels" -}}
helm.sh/chart: {{ include "claw-swarm-panel.chart" . }}
{{ include "claw-swarm-panel.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Base selector labels (no component).
*/}}
{{- define "claw-swarm-panel.selectorLabels" -}}
app.kubernetes.io/name: {{ include "claw-swarm-panel.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Selector labels for the apiserver component.
*/}}
{{- define "claw-swarm-panel.apiserverSelectorLabels" -}}
{{ include "claw-swarm-panel.selectorLabels" . }}
app.kubernetes.io/component: apiserver
{{- end }}

{{/*
Fully-qualified name for the apiserver Deployment.
*/}}
{{- define "claw-swarm-panel.apiserverName" -}}
{{- printf "%s-apiserver" (include "claw-swarm-panel.fullname" .) | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Service account name.
*/}}
{{- define "claw-swarm-panel.serviceAccountName" -}}
{{- if .Values.serviceAccount.create }}
{{- default (include "claw-swarm-panel.fullname" .) .Values.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.serviceAccount.name }}
{{- end }}
{{- end }}
