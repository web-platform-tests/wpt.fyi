//go:build !small && !medium && !large && !cloud && !race
// +build !small,!medium,!large,!cloud,!race

package shared

func init() {
	panic("Tests were run without -tags=[small|medium|large|cloud|race]")
}
