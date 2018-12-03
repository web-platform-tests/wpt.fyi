{{ template "_successfully_scraped.md" . }}

Uh-oh - it looks like there are some newly-failing results when we compared the affected tests
to the latest run against the `master` branch.

{{ template "_pr_and_master_specs.md" . }}

### Regressions

Test | `master` | `{{ printf "%.7s" .PRRun.FullRevisionHash }}`
--- | --- | ---
{{ range $test, $results := .Regressions -}}
{{ $test }} | {{ $results.PassingBefore }} / {{ $results.TotalBefore }} | {{ $results.PassingAfter }} / {{ $results.TotalAfter }}
{{end}}
{{ if gt .More 0 -}}
And {{ .More }} others...
{{ end }}

[Visual comparison of the results]({{ .DiffURL }})

Other views that might be useful:
- [`{{ printf "%.7s" .PRRun.FullRevisionHash }}` vs `master`@`{{ printf "%.7s" .MasterRun.FullRevisionHash }}`]({{ .DiffURL }})
- [`{{ printf "%.7s" .PRRun.FullRevisionHash }}` vs latest master]({{ .MasterDiffURL }})
- [Latest results for `{{ printf "%.7s" .PRRun.FullRevisionHash }}`]({{.HostURL}}?sha={{.PRRun.Revision}})

{{ template "_file_an_issue.md" . }}
