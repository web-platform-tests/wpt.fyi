Run | Spec
--- | ---
`master` | {{ .MasterRun.String }}
`{{ printf "%.7s" .PRRun.FullRevisionHash }}` | {{ .PRRun.String }}