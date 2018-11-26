Results have successfully been scraped and added to [{{ .HostName }}]({{ .HostURL }}).

Uh-oh - it looks like there are some newly-failing results.

Test | `master` | `{{ printf "%.7s" .HeadSHA }}`
--- | --- | ---
{{ range $test, $results := .Regressions -}}
{{ $test }} | {{ $results.PassingBefore }} / {{ $results.TotalBefore }} | {{ $results.PassingAfter }} / {{ $results.TotalAfter }}
{{end}}
{{ if gt .More 0 -}}
And {{ .More }} others...
{{ end }}
A visual comparison of the results against `master` run:
[This PR vs master]({{ .DiffURL }})
