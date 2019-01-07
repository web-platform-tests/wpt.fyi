{{ template "_successfully_scraped.md" . }}

Uh-oh - it looks like there are some newly-failing results when we compared the affected tests
to the latest run against the `master` branch.

{{ template "_pr_and_master_specs.md" . }}

### Regressions

Test | `master` | `{{ printf "%.7s" .HeadRun.FullRevisionHash }}`
--- | --- | ---
{{ range $test, $results := .Regressions -}}
{{ $test }} | {{ $results.PassingBefore }} / {{ $results.TotalBefore }} | {{ $results.PassingAfter }} / {{ $results.TotalAfter }}
{{end}}
{{ if gt .More 0 -}}
And {{ .More }} others...
{{ end }}

[Visual comparison of the results]({{ .DiffURL }})

Other links that might be useful:
{{- range $pr := .CheckState.PRNumbers }}
- [Latest results for PR #{{ $pr }}]({{ $.HostURL }}results/?pr={{ $pr }})
- [All runs for PR #{{ $pr }}]({{ $.HostURL }}runs/?pr={{ $pr }})
{{- end}}
- [`{{ printf "%.7s" .HeadRun.FullRevisionHash }}` vs its merge base]({{ .DiffURL }})
{{- if .MasterDiffURL }}
- [`{{ printf "%.7s" .HeadRun.FullRevisionHash }}` vs latest master]({{ .MasterDiffURL }})
{{- end }}
- [Latest results for `{{ printf "%.7s" .HeadRun.FullRevisionHash }}`]({{.HostURL}}results/?sha={{.HeadRun.Revision}})

{{ template "_file_an_issue.md" . }}
