package logging

import (
	"log"
	"reflect"
	"runtime"
)

// LogAnyType logs any type of variable with its type, value, and length (if applicable).
func LogAnyType(variable interface{}, variableName string) {
	// Get the variable type using reflection
	varType := reflect.TypeOf(variable)
	// Get the caller function details for contextual logging
	_, file, line, ok := runtime.Caller(1)
	location := ""
	if ok {
		location = file + ":" + string(rune(line))
	}

	// Convert the variable to a reflect.Value to access its value
	varValue := reflect.ValueOf(variable)

	// Initialize length variable
	length := -1 // Default to -1 to indicate non-applicable

	// Check if the variable is of a type that has a length
	if varValue.Kind() == reflect.Array || varValue.Kind() == reflect.Slice || varValue.Kind() == reflect.Map || varValue.Kind() == reflect.String {
		length = varValue.Len()
	}

	// Log the type, value, and length (if applicable) of the variable
	if length >= 0 {
		log.Printf("[%s] Variable Name: '%s', Type: %s, Length: %d, Value: %v\n", location, variableName, varType, length, varValue)
	} else {
		log.Printf("[%s] Variable Name: '%s', Type: %s, Value: %v\n", location, variableName, varType, varValue)
	}
}
