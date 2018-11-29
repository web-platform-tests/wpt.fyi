{{ template "_successfully_scraped.md" . }}

There were no regressions detected in the results.

{{ template "_pr_and_master_specs.md" . }}

### Results

Test | `{{ printf "%.7s" .PRRun.FullRevisionHash }}`
--- | ---
{{ range $test, $results := .Results -}}
{{ $test }} | {{ index $results 0 }} / {{ index $results 1 }}
{{end}}
{{ if gt .More 0 -}}
And {{ .More }} others...
{{ end }}

[Visual comparison of the results]({{ .DiffURL }})

Other views that might be useful:
- [`{{ printf "%.7s" .PRRun.FullRevisionHash }}` vs `master`@`{{ printf "%.7s" .MasterRun.FullRevisionHash }}`]({{ .DiffURL }})
- [`{{ printf "%.7s" .PRRun.FullRevisionHash }}` vs latest master]({{ .MasterDiffURL }})
- [Latest results for `{{ printf "%.7s" .PRRun.FullRevisionHash }}`]({{.HostURL}}?sha={{.PRRun.Revision}})