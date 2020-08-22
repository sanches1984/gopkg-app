package transport

import (
	"context"
	"regexp"
	"strings"

	"github.com/severgroup-tt/gopkg-errors"
)

const (
	msgBadRequest       = "Invalid JSON payload"
	msgInvalidEnum      = "Unknown enum value"
	msgInvalidTimestamp = "Invalid timestamp value"
)

var unknownEnumValueRe = regexp.MustCompile(`unknown value (.+) for enum`)
var badTimestampRe = regexp.MustCompile(`parsing time ("[^"]+")`)

// TransformUnmarshalerError ...
func TransformUnmarshalerError(err error) error {
	msg := msgBadRequest
	errMsg := err.Error()

	if strings.Contains(errMsg, "unknown value") {
		matches := unknownEnumValueRe.FindStringSubmatch(errMsg)
		if len(matches) < 2 {
			msg = msgInvalidEnum
		} else {
			msg = msgInvalidEnum + " " + matches[1]
		}
	} else if strings.Contains(errMsg, "bad Timestamp") {
		matches := badTimestampRe.FindStringSubmatch(errMsg)
		if len(matches) < 2 {
			msg = msgInvalidTimestamp
		} else {
			msg = msgInvalidTimestamp + " " + matches[1]
		}
	}

	return errors.BadRequest.ErrWrap(context.Background(), msg, err)
}
