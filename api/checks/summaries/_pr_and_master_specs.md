Run | Spec
--- | ---
`master` | {{ .BaseRun.String }}
`{{ printf "%.7s" .HeadRun.FullRevisionHash }}` | {{ .HeadRun.String }}
