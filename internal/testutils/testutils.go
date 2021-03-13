package testutils

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"strconv"
	"time"
)

const (
	HeaderTimestamp = "X-Slack-Request-Timestamp"
	HeaderSignature = "X-Slack-Signature"
)

func AddSignature(h http.Header, key, body []byte, timestamp time.Time) error {
	hash := hmac.New(sha256.New, key)
	strTime := strconv.FormatInt(timestamp.Unix(), 10)
	if _, err := hash.Write([]byte(fmt.Sprintf("v0:%s:", strTime))); err != nil {
		return err
	}
	if _, err := hash.Write(body); err != nil {
		return err
	}
	signature := hex.EncodeToString(hash.Sum(nil))

	h.Set(HeaderTimestamp, strTime)
	h.Set(HeaderSignature, "v0="+signature)
	return nil
}
