Run | Spec
--- | ---
`{{ printf "%.7s" .BaseRun.FullRevisionHash }}` | {{ .BaseRun.String }}
`{{ printf "%.7s" .HeadRun.FullRevisionHash }}` | {{ .HeadRun.String }}
