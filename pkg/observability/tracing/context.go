// Package tracing provides context propagation utilities for distributed tracing.
package tracing

import "context"

// ExtractTraceContext extracts trace context from incoming requests
// This is useful for propagating trace IDs from external systems (e.g., API headers)
type TraceContext struct {
	TraceID      string
	ParentSpanID string
	Sampled      bool
}

// ExtractFromContext extracts trace context from a context.Context
func ExtractFromContext(ctx context.Context) *TraceContext {
	return &TraceContext{
		TraceID:      GetTraceID(ctx),
		ParentSpanID: GetParentSpanID(ctx),
		Sampled:      true, // Default to sampled
	}
}

// InjectToContext injects trace context into a context.Context
func InjectToContext(ctx context.Context, tc *TraceContext) context.Context {
	if tc.TraceID != "" {
		ctx = WithTraceID(ctx, tc.TraceID)
	}
	if tc.ParentSpanID != "" {
		ctx = context.WithValue(ctx, ParentSpanIDKey, tc.ParentSpanID)
	}
	return ctx
}

// MergeTraceContext merges trace context from an external source into the current context
// If the external context has a trace ID, use it; otherwise generate a new one
func MergeTraceContext(ctx context.Context, external *TraceContext) context.Context {
	if external == nil {
		return ctx
	}

	// If external has a trace ID, use it
	if external.TraceID != "" {
		ctx = WithTraceID(ctx, external.TraceID)
	}

	// If external has a parent span ID, set it as parent
	if external.ParentSpanID != "" {
		ctx = context.WithValue(ctx, ParentSpanIDKey, external.ParentSpanID)
	}

	return ctx
}
