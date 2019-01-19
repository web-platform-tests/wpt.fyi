{{- range $pr := .CheckState.PRNumbers }}
- [Latest results for PR #{{ $pr }}]({{ $.HostURL }}results/?pr={{ $pr }}&label=pr_head)
- [All runs for PR #{{ $pr }}]({{ $.HostURL }}runs/?pr={{ $pr }})
{{- end}}