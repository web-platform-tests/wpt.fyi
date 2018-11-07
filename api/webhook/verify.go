package webhook

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/web-platform-tests/wpt.fyi/shared"
)

func verifyAndGetPayload(r *http.Request) ([]byte, error) {
	ctx := shared.NewAppEngineContext(r)
	log := shared.GetLogger(ctx)

	payload, err := ioutil.ReadAll(r.Body)
	r.Body.Close()
	if err != nil {
		log.Errorf("Failed to read request body: %s", err.Error())
		return nil, errors.New("Failed to read request body")
	}

	secret, err := shared.GetSecret(ctx, "github-tc-webhook-secret")
	if err != nil {
		log.Errorf("Failed to get verification secret: %s", err.Error())
		return nil, errors.New("Internal error")
	}

	if !verifySignature(payload, r.Header.Get("X-Hub-Signature"), secret) {
		return nil, errors.New("HMAC verification failed")
	}
	return payload, nil
}

func verifySignature(message []byte, signature string, secret string) bool {
	// https://developer.github.com/webhooks/securing/
	signature = strings.TrimPrefix(signature, "sha1=")
	messageMAC, err := hex.DecodeString(signature)
	if err != nil {
		return false
	}
	mac := hmac.New(sha1.New, []byte(secret))
	mac.Write(message)
	expectedMAC := mac.Sum(nil)
	return hmac.Equal(messageMAC, expectedMAC)
}
