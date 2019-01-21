{{ template "_successfully_scraped.md" . }}

There were no regressions detected in the results.

{{ template "_pr_and_master_specs.md" . }}

### Results

Test | `{{ printf "%.7s" .HeadRun.FullRevisionHash }}`
--- | ---
{{ range $test, $results := .Results -}}
{{ $test }} | {{ index $results 0 }} / {{ index $results 1 }}
{{end}}
{{ if gt .More 0 -}}
And {{ .More }} others...
{{ end }}

[Visual comparison of the results]({{ .DiffURL }})

Other links that might be useful:
{{ template "_pr_runs_links.md" . }}
- [`{{ printf "%.7s" .HeadRun.FullRevisionHash }}` vs its merge base]({{ .DiffURL }})
{{- if .MasterDiffURL }}
- [`{{ printf "%.7s" .HeadRun.FullRevisionHash }}` vs latest master]({{ .MasterDiffURL }})
{{- end }}
- [Latest results for `{{ printf "%.7s" .HeadRun.FullRevisionHash }}`]({{.HostURL}}?sha={{.HeadRun.Revision}})

{{ template "_file_an_issue.md" . }}
