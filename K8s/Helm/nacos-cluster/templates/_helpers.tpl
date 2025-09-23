{{/* 生成通用标签 */}}
{{- define "nacos-cluster.labels" -}}
app: {{ include "nacos-cluster.name" . }}
helm.sh/chart: {{ .Chart.Name }}-{{ .Chart.Version }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end -}}

{{/* 生成应用名称 */}}
{{- define "nacos-cluster.name" -}}
{{- default "nacos" .Values.nameOverride | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/* 生成完整名称 */}}
{{- define "nacos-cluster.fullname" -}}
{{- if .Values.fullnameOverride -}}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" -}}
{{- else -}}
{{- $name := default "nacos" .Values.nameOverride -}}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" -}}
{{- end -}}
{{- end -}}