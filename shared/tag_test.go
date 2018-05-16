// +build !small,!medium,!large

package shared

func init() {
	panic("Tests were run without -tags=[small|medium|large]")
}
