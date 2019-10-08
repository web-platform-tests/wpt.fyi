{{ template "_successfully_scraped.md" . }}

There were no regressions detected in the results.

{{ template "_pr_and_master_specs.md" . }}

### Results

Test | `master` | `{{ printf "%.7s" .HeadRun.FullRevisionHash }}`
--- | --- | ---
{{ range $test, $results := .Results -}}
{{ escapeMD $test }} | {{ $results.PassingBefore }} / {{ $results.TotalBefore }} | {{ $results.PassingAfter }} / {{ $results.TotalAfter }}
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
- [Latest results for `{{ printf "%.7s" .HeadRun.FullRevisionHash }}`]({{.HostURL}}?sha={{.HeadRun.Revision}}&label=pr_head)

{{ template "_file_an_issue.md" . }}
