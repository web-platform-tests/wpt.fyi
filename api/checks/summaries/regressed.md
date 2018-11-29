Results have successfully been scraped and added to [{{ .HostName }}]({{ .HostURL }}).

Uh-oh - it looks like there are some newly-failing results when we compared the affected tests
to the latest run against the `master` branch.

Run | Spec
`master` | [{{ .MasterRun.String }}]
`{{ printf "%.7s" .PRRun.FullRevisionHash }}` | {{ .PRRun.String }}

### Regressions

Test | `master` | `{{ printf "%.7s" .PRRun.FullRevisionHash }}`
--- | --- | ---
{{ range $test, $results := .Regressions -}}
{{ $test }} | {{ $results.PassingBefore }} / {{ $results.TotalBefore }} | {{ $results.PassingAfter }} / {{ $results.TotalAfter }}
{{end}}
{{ if gt .More 0 -}}
And {{ .More }} others...
{{ end }}

You can view a visual comparison of all the results [here]]({{ .DiffURL }}).

Other views that might be useful:
- [`{{ printf "%.7s" .PRRun.FullRevisionHash }}` vs `master`@`{{ printf "%.7s" .MasterRun.FullRevisionHash }}`]({{ .DiffURL }})
- [`{{ printf "%.7s" .PRRun.FullRevisionHash }}` vs latest master]({{ .MasterDiffURL }})
- [Latest results for `{{ printf "%.7s" .PRRun.FullRevisionHash }}`]({{.HostURL}}?sha={{.PRRun.Revision}})