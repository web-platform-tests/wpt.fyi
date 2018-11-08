// +build small

package webhook

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestVerifySignature(t *testing.T) {
	message := []byte("test")
	signature := "a053ee211b4693456ef071e336f74ab699250318"
	secret := "95bfab9afa719185ee7c3658356b166b7f45349a"
	assert.True(t, verifySignature(message, signature, secret))

	assert.False(t, verifySignature([]byte("foobar"), signature, secret))
	assert.False(t, verifySignature(
		message, "875a5feef4cde4265d6d5d21c304d755903ccb60", secret))
	assert.False(t, verifySignature(
		message, signature, "875a5feef4cde4265d6d5d21c304d755903ccb60"))
	// Test an ill-formed (odd-length) signature.
	assert.False(t, verifySignature(
		message, "875a5feef4cde4265d6d5d21c304d755903ccb6", secret))
}
