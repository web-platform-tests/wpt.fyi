// +build !small,!medium,!large,!cloud

package shared

func init() {
	panic("Tests were run without -tags=[small|medium|large|cloud]")
}
