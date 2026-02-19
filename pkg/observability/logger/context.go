// Package logger provides context-aware logging with trace context propagation.
package logger

import (
	"context"

	"github.com/sipeed/picoclaw/pkg/logger"
	"github.com/sipeed/picoclaw/pkg/observability/tracing"
)

// ContextLogger is a logger with bound context and fields
type ContextLogger struct {
	ctx       context.Context
	component string
	fields    map[string]interface{}
}

// NewContextLogger creates a new context logger with the given context
func NewContextLogger(ctx context.Context) *ContextLogger {
	return &ContextLogger{
		ctx:    ctx,
		fields: make(map[string]interface{}),
	}
}

// WithComponent sets the component name for this logger
func (cl *ContextLogger) WithComponent(component string) *ContextLogger {
	cl.component = component
	return cl
}

// WithField adds a single field to this logger
func (cl *ContextLogger) WithField(key string, value interface{}) *ContextLogger {
	cl.fields[key] = value
	return cl
}

// WithFields adds multiple fields to this logger
func (cl *ContextLogger) WithFields(fields map[string]interface{}) *ContextLogger {
	for k, v := range fields {
		cl.fields[k] = v
	}
	return cl
}

// mergeFields combines trace context fields with user-defined fields
func (cl *ContextLogger) mergeFields() map[string]interface{} {
	result := make(map[string]interface{})

	// Add trace context fields first
	traceFields := tracing.GetSpanContext(cl.ctx)
	for k, v := range traceFields {
		result[k] = v
	}

	// Add user-defined fields (can override trace fields)
	for k, v := range cl.fields {
		result[k] = v
	}

	return result
}

// Debug logs a debug message with context and fields
func (cl *ContextLogger) Debug(message string) {
	logger.DebugCF(cl.component, message, cl.mergeFields())
}

// DebugWithFields logs a debug message with additional fields
func (cl *ContextLogger) DebugWithFields(message string, fields map[string]interface{}) {
	merged := cl.mergeFields()
	for k, v := range fields {
		merged[k] = v
	}
	logger.DebugCF(cl.component, message, merged)
}

// Info logs an info message with context and fields
func (cl *ContextLogger) Info(message string) {
	logger.InfoCF(cl.component, message, cl.mergeFields())
}

// InfoWithFields logs an info message with additional fields
func (cl *ContextLogger) InfoWithFields(message string, fields map[string]interface{}) {
	merged := cl.mergeFields()
	for k, v := range fields {
		merged[k] = v
	}
	logger.InfoCF(cl.component, message, merged)
}

// Warn logs a warning message with context and fields
func (cl *ContextLogger) Warn(message string) {
	logger.WarnCF(cl.component, message, cl.mergeFields())
}

// WarnWithFields logs a warning message with additional fields
func (cl *ContextLogger) WarnWithFields(message string, fields map[string]interface{}) {
	merged := cl.mergeFields()
	for k, v := range fields {
		merged[k] = v
	}
	logger.WarnCF(cl.component, message, merged)
}

// Error logs an error message with context and fields
func (cl *ContextLogger) Error(message string) {
	logger.ErrorCF(cl.component, message, cl.mergeFields())
}

// ErrorWithFields logs an error message with additional fields
func (cl *ContextLogger) ErrorWithFields(message string, fields map[string]interface{}) {
	merged := cl.mergeFields()
	for k, v := range fields {
		merged[k] = v
	}
	logger.ErrorCF(cl.component, message, merged)
}

// Fatal logs a fatal message and exits
func (cl *ContextLogger) Fatal(message string) {
	logger.FatalCF(cl.component, message, cl.mergeFields())
}

// FatalWithFields logs a fatal message with fields and exits
func (cl *ContextLogger) FatalWithFields(message string, fields map[string]interface{}) {
	merged := cl.mergeFields()
	for k, v := range fields {
		merged[k] = v
	}
	logger.FatalCF(cl.component, message, merged)
}

// WithContext creates a new logger with the given context
func WithContext(ctx context.Context) *ContextLogger {
	return NewContextLogger(ctx)
}

// LogWithSpan logs a message within a span, automatically adding span context
func LogWithSpan(ctx context.Context, component, message string, fields map[string]interface{}) {
	merged := tracing.GetSpanContext(ctx)
	for k, v := range fields {
		merged[k] = v
	}
	logger.InfoCF(component, message, merged)
}
