package marketshandlers

import (
	"context"
	"errors"
	"strings"
)

func isRequestCanceled(err error) bool {
	if err == nil {
		return false
	}

	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return true
	}

	errMsg := strings.ToLower(err.Error())
	return strings.Contains(errMsg, "context canceled") ||
		strings.Contains(errMsg, "context already done: context canceled")
}
