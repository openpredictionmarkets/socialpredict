package logger

import (
	"io"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"time"
)

const (
	FieldAddress    = "address"
	FieldComponent  = "component"
	FieldDurationMS = "duration_ms"
	FieldError      = "error"
	FieldMethod     = "method"
	FieldOperation  = "operation"
	FieldPath       = "path"
	FieldRequestID  = "request_id"
	FieldSource     = "source"
	FieldSpanID     = "span_id"
	FieldStatusCode = "status_code"
	FieldTraceFlags = "trace_flags"
	FieldTraceID    = "trace_id"

	redactedValue = "[REDACTED]"
)

var (
	keyValueSecretPattern = regexp.MustCompile(`(?i)("?(?:password|passwd|pwd|token|secret|api[_-]?key|apikey|authorization|cookie|set-cookie)"?\s*[:=]\s*)("[^"]*"|'[^']*'|[^\s,;]+)`)
	bearerSecretPattern   = regexp.MustCompile(`(?i)\bBearer\s+[^\s",;]+`)
)

// Field is a single stable runtime log field.
type Field struct {
	Key   string
	Value string
}

// RuntimeLogger is the concrete backend runtime log adapter.
type RuntimeLogger struct {
	sink *log.Logger
	exit func(int)
}

var standard = newRuntimeLogger(os.Stdout, os.Exit)

// New creates a logger that writes runtime logs to output.
func New(output io.Writer) *RuntimeLogger {
	return newRuntimeLogger(output, os.Exit)
}

func newRuntimeLogger(output io.Writer, exit func(int)) *RuntimeLogger {
	if output == nil {
		output = os.Stdout
	}
	if exit == nil {
		exit = os.Exit
	}

	return &RuntimeLogger{
		sink: log.New(output, "", log.LstdFlags),
		exit: exit,
	}
}

// Standard returns the shared process logger used by package-level helpers.
func Standard() *RuntimeLogger {
	return standard
}

// String creates a normalized string field.
func String(key, value string) Field {
	return Field{
		Key:   normalizeKey(key),
		Value: value,
	}
}

// Address records a listening or remote address using the stable runtime key.
func Address(value string) Field {
	return String(FieldAddress, value)
}

// DurationMS records a duration in whole milliseconds.
func DurationMS(value time.Duration) Field {
	return String(FieldDurationMS, strconv.FormatInt(value.Milliseconds(), 10))
}

// Method records an HTTP method for request-boundary logs.
func Method(value string) Field {
	return String(FieldMethod, value)
}

// Operation records the runtime operation name for startup, server, or middleware logs.
func Operation(value string) Field {
	return String(FieldOperation, value)
}

// Path records an HTTP path for request-boundary logs.
func Path(value string) Field {
	return String(FieldPath, value)
}

// RequestID records the boundary-generated or propagated request identifier.
func RequestID(value string) Field {
	return String(FieldRequestID, value)
}

// Err records an error value using the stable runtime key.
func Err(err error) Field {
	if err == nil {
		return Field{}
	}

	return String(FieldError, err.Error())
}

// TraceContext adds explicit OpenTelemetry-aligned correlation fields when available.
// It intentionally stops at log-field correlation; provider and exporter wiring stays
// outside this package until a later runtime observability rollout.
func TraceContext(traceID, spanID, traceFlags string) []Field {
	fields := make([]Field, 0, 3)

	traceID = strings.ToLower(strings.TrimSpace(traceID))
	if validTraceID(traceID) {
		fields = append(fields, String(FieldTraceID, traceID))
	}

	spanID = strings.ToLower(strings.TrimSpace(spanID))
	if validSpanID(spanID) {
		fields = append(fields, String(FieldSpanID, spanID))
	}

	traceFlags = strings.ToLower(strings.TrimSpace(traceFlags))
	if validHex(traceFlags, 2) {
		fields = append(fields, String(FieldTraceFlags, traceFlags))
	}

	return fields
}

// TraceContextFromTraceparent extracts OTel-aligned trace fields from a W3C traceparent header.
// It preserves request-boundary correlation without turning stdout logging into the
// tracing or exporter contract.
func TraceContextFromTraceparent(traceparent string) []Field {
	traceparent = strings.TrimSpace(traceparent)
	if traceparent == "" {
		return nil
	}

	parts := strings.Split(traceparent, "-")
	if len(parts) != 4 || !validHex(strings.ToLower(parts[0]), 2) {
		return nil
	}
	if !validTraceID(strings.ToLower(parts[1])) || !validSpanID(strings.ToLower(parts[2])) || !validHex(strings.ToLower(parts[3]), 2) {
		return nil
	}

	return []Field{
		String(FieldTraceID, strings.ToLower(parts[1])),
		String(FieldSpanID, strings.ToLower(parts[2])),
		String(FieldTraceFlags, strings.ToLower(parts[3])),
	}
}

// StatusCode records an HTTP status code for request-boundary logs.
func StatusCode(value int) Field {
	return String(FieldStatusCode, strconv.Itoa(value))
}

// Info writes an informational runtime log entry.
func Info(component, message string, fields ...Field) {
	standard.log(3, "INFO", component, message, fields...)
}

// Warn writes a warning runtime log entry.
func Warn(component, message string, fields ...Field) {
	standard.log(3, "WARN", component, message, fields...)
}

// Error writes an error runtime log entry.
func Error(component, message string, err error, fields ...Field) {
	fields = append(fields, Err(err))
	standard.log(3, "ERROR", component, message, fields...)
}

// Fatal writes a fatal runtime log entry and exits the process.
func Fatal(component, message string, err error, fields ...Field) {
	fields = append(fields, Err(err))
	standard.log(3, "FATAL", component, message, fields...)
	standard.exit(1)
}

// Info writes an informational runtime log entry.
func (l *RuntimeLogger) Info(component, message string, fields ...Field) {
	l.log(3, "INFO", component, message, fields...)
}

// Warn writes a warning runtime log entry.
func (l *RuntimeLogger) Warn(component, message string, fields ...Field) {
	l.log(3, "WARN", component, message, fields...)
}

// Error writes an error runtime log entry.
func (l *RuntimeLogger) Error(component, message string, err error, fields ...Field) {
	fields = append(fields, Err(err))
	l.log(3, "ERROR", component, message, fields...)
}

// Fatal writes a fatal runtime log entry and exits the process.
func (l *RuntimeLogger) Fatal(component, message string, err error, fields ...Field) {
	fields = append(fields, Err(err))
	l.log(3, "FATAL", component, message, fields...)
	l.exit(1)
}

// LogInfo retains the older call shape while emitting the stable runtime vocabulary.
func LogInfo(context, function, message string) {
	standard.log(3, "INFO", context, message, Operation(function))
}

// LogWarn retains the older call shape while emitting the stable runtime vocabulary.
func LogWarn(context, function, message string) {
	standard.log(3, "WARN", context, message, Operation(function))
}

// LogError retains the older call shape while emitting the stable runtime vocabulary.
func LogError(context, function string, err error) {
	fields := []Field{Operation(function)}
	if err != nil {
		fields = append(fields, Err(err))
	}

	standard.log(3, "ERROR", context, failureMessage(function), fields...)
}

func (l *RuntimeLogger) log(skip int, level, component, message string, fields ...Field) {
	if l == nil || l.sink == nil {
		return
	}

	values := map[string]string{}
	if component = redactSensitiveText(strings.TrimSpace(component)); component != "" {
		values[FieldComponent] = component
	}

	for _, field := range fields {
		key := normalizeKey(field.Key)
		if key == "" {
			continue
		}
		values[key] = sanitizeFieldValue(key, field.Value)
	}

	if source, ok := callerSource(skip); ok {
		values[FieldSource] = source
	}

	parts := []string{
		"level=" + level,
		"msg=" + strconv.Quote(redactSensitiveText(strings.TrimSpace(message))),
	}

	orderedKeys := []string{
		FieldComponent,
		FieldOperation,
		FieldRequestID,
		FieldTraceID,
		FieldSpanID,
		FieldTraceFlags,
		FieldMethod,
		FieldPath,
		FieldStatusCode,
		FieldDurationMS,
		FieldAddress,
		FieldError,
		FieldSource,
	}

	used := make(map[string]struct{}, len(orderedKeys))
	for _, key := range orderedKeys {
		value, ok := values[key]
		if !ok || value == "" {
			continue
		}
		parts = append(parts, formatField(key, value))
		used[key] = struct{}{}
	}

	extraKeys := make([]string, 0, len(values))
	for key, value := range values {
		if value == "" {
			continue
		}
		if _, ok := used[key]; ok {
			continue
		}
		extraKeys = append(extraKeys, key)
	}
	sort.Strings(extraKeys)

	for _, key := range extraKeys {
		parts = append(parts, formatField(key, values[key]))
	}

	l.sink.Println(strings.Join(parts, " "))
}

func callerSource(skip int) (string, bool) {
	_, file, line, ok := runtime.Caller(skip)
	if !ok {
		return "", false
	}

	return filepath.Base(file) + ":" + strconv.Itoa(line), true
}

func formatField(key, value string) string {
	return key + "=" + strconv.Quote(value)
}

func failureMessage(operation string) string {
	operation = strings.TrimSpace(operation)
	if operation == "" {
		return "operation failed"
	}

	return operation + " failed"
}

func sanitizeFieldValue(key, value string) string {
	if sensitiveFieldKey(key) {
		return redactedValue
	}

	return redactSensitiveText(strings.TrimSpace(value))
}

func redactSensitiveText(value string) string {
	if value == "" {
		return ""
	}

	redacted := bearerSecretPattern.ReplaceAllString(value, "Bearer "+redactedValue)
	redacted = keyValueSecretPattern.ReplaceAllString(redacted, "${1}"+redactedValue)
	return redacted
}

func sensitiveFieldKey(key string) bool {
	normalized := strings.ReplaceAll(normalizeKey(key), "_", "")
	for _, fragment := range []string{
		"password",
		"passwd",
		"pwd",
		"token",
		"secret",
		"apikey",
		"authorization",
		"cookie",
		"setcookie",
	} {
		if strings.Contains(normalized, fragment) {
			return true
		}
	}

	return false
}

func normalizeKey(key string) string {
	key = strings.TrimSpace(strings.ToLower(key))
	if key == "" {
		return ""
	}

	var builder strings.Builder
	builder.Grow(len(key))

	lastUnderscore := false
	for _, r := range key {
		switch {
		case r >= 'a' && r <= 'z':
			builder.WriteRune(r)
			lastUnderscore = false
		case r >= '0' && r <= '9':
			builder.WriteRune(r)
			lastUnderscore = false
		case r == '_' || r == '-' || r == ' ':
			if lastUnderscore {
				continue
			}
			builder.WriteByte('_')
			lastUnderscore = true
		}
	}

	return strings.Trim(builder.String(), "_")
}

func validTraceID(value string) bool {
	return validHex(value, 32) && !allZero(value)
}

func validSpanID(value string) bool {
	return validHex(value, 16) && !allZero(value)
}

func validHex(value string, length int) bool {
	if len(value) != length {
		return false
	}

	for _, r := range value {
		switch {
		case r >= '0' && r <= '9':
		case r >= 'a' && r <= 'f':
		case r >= 'A' && r <= 'F':
		default:
			return false
		}
	}

	return true
}

func allZero(value string) bool {
	for _, r := range value {
		if r != '0' {
			return false
		}
	}

	return true
}
