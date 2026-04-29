package marketshandlers

import (
	"context"
	"errors"
)

func isRequestCanceled(err error) bool {
	return errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded)
}
