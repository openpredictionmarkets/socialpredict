package errors

import (
	"log"
)

// ErrorLogger logs an error and returns a boolean indicating whether an error occurred.
func ErrorLogger(err error, errMsg string) bool {
	if err != nil {
		log.Printf("Error: %s - %s\n", errMsg, err) // Combine your custom message with the error's message.
		return true                                 // Indicate that an error was handled.
	}
	return false // No error to handle.
}
