{{/* 生成通用标签 */}}
{{- define "mysql-cluster.labels" -}}
app: {{ include "mysql-cluster.name" . }}
helm.sh/chart: {{ .Chart.Name }}-{{ .Chart.Version }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end -}}

{{/* 生成应用名称 */}}
{{- define "mysql-cluster.name" -}}
{{- default "mysql" .Values.nameOverride | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/* 生成完整名称 */}}
{{- define "mysql-cluster.fullname" -}}
{{- if .Values.fullnameOverride -}}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" -}}
{{- else -}}
{{- $name := default "mysql" .Values.nameOverride -}}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" -}}
{{- end -}}
{{- end -}}

{{/* 生成主节点名称 */}}
{{- define "mysql-cluster.master.fullname" -}}
{{- printf "%s-master" (include "mysql-cluster.fullname" .) -}}
{{- end -}}

{{/* 生成从节点名称 */}}
{{- define "mysql-cluster.slave.fullname" -}}
{{- printf "%s-slave" (include "mysql-cluster.fullname" .) -}}
{{- end -}}

{{/* 生成配置映射名称 */}}
{{- define "mysql-cluster.configmap" -}}
{{- printf "%s-mysql-config" .Release.Name -}}
{{- end -}}
