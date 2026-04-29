package errors

import (
	"socialpredict/logger"
)

// ErrorLogger logs an error and returns a boolean indicating whether an error occurred.
func ErrorLogger(err error, errMsg string) bool {
	if err != nil {
		logger.LogError("NormalError", errMsg, err)
		return true
	}
	return false
}
