package logging

import "fmt"

// MockLogger is a mock implementation of Logger for testing.
type MockLogger struct {
	CalledFatalf  bool
	MessageFatalf string
	CalledPrintf  bool
	MessagePrintf string
}

func (m *MockLogger) Fatalf(format string, args ...interface{}) {
	m.CalledFatalf = true
	m.MessageFatalf = fmt.Sprintf(format, args...)
	panic(m.MessageFatalf) // Simulate Fatalf by panicking
}

func (m *MockLogger) Printf(format string, args ...interface{}) {
	m.CalledPrintf = true
	m.MessagePrintf = fmt.Sprintf(format, args...)
}
