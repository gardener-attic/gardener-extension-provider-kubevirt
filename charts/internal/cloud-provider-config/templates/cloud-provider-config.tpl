{{- define "cloud-provider-config" -}}
kubeconfig: |
{{ .Values.kubeconfig | indent 2 }}
loadbalancer:
  enabled: {{ .Values.loadBalancer.enabled }}
  creationPollInterval: {{ .Values.loadBalancer.creationPollInterval }}
instances:
  enabled: {{ .Values.instances.enabled }}
  enableInstanceTypes: {{ .Values.instances.enableInstanceTypes }}
zones:
  enabled: {{ .Values.zones.enabled }}
{{- end -}}

