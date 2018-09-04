package query

import "strings"

func canonicalizeStr(str string) string {
	return strings.ToLower(str)
}
