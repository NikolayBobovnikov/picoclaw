// Package tracing provides request tracing with span tracking for observability.
package tracing

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
)

// Context key type for trace context
type contextKey int

const (
	// TraceIDKey is the context key for the trace ID
	TraceIDKey contextKey = iota
	// SpanIDKey is the context key for the current span ID
	SpanIDKey
	// ParentSpanIDKey is the context key for the parent span ID
	ParentSpanIDKey
)

// Span represents a single unit of work in a distributed trace
type Span struct {
	ID        string    `json:"id"`
	ParentID  string    `json:"parent_id,omitempty"`
	Name      string    `json:"name"`
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time,omitempty"`
	Duration  int64     `json:"duration_ms,omitempty"`
	TraceID   string    `json:"trace_id"`
	Component string    `json:"component,omitempty"`
	Fields    map[string]interface{} `json:"fields,omitempty"`
}

// SpanRecorder records spans for export/analysis
type SpanRecorder interface {
	Record(span *Span)
}

// DefaultRecorder is a no-op recorder
type DefaultRecorder struct{}

func (d *DefaultRecorder) Record(span *Span) {}

var globalRecorder SpanRecorder = &DefaultRecorder{}
var recorderMutex sync.RWMutex

// SetGlobalRecorder sets the global span recorder
func SetGlobalRecorder(recorder SpanRecorder) {
	recorderMutex.Lock()
	defer recorderMutex.Unlock()
	globalRecorder = recorder
}

// GetGlobalRecorder returns the global span recorder
func GetGlobalRecorder() SpanRecorder {
	recorderMutex.RLock()
	defer recorderMutex.RUnlock()
	return globalRecorder
}

// GenerateTraceID generates a new trace ID using UUID
func GenerateTraceID() string {
	return uuid.New().String()
}

// GenerateSpanID generates a new span ID using UUID
func GenerateSpanID() string {
	return uuid.New().String()
}

// WithTraceID adds a trace ID to the context
func WithTraceID(ctx context.Context, traceID string) context.Context {
	return context.WithValue(ctx, TraceIDKey, traceID)
}

// WithSpan adds a new span to the context and records it when ended
func WithSpan(ctx context.Context, name string) (context.Context, *Span) {
	traceID := GetTraceID(ctx)
	if traceID == "" {
		traceID = GenerateTraceID()
		ctx = WithTraceID(ctx, traceID)
	}

	parentID := GetSpanID(ctx)
	spanID := GenerateSpanID()

	span := &Span{
		ID:        spanID,
		ParentID:  parentID,
		Name:      name,
		StartTime: time.Now(),
		TraceID:   traceID,
		Fields:    make(map[string]interface{}),
	}

	ctx = context.WithValue(ctx, SpanIDKey, spanID)
	if parentID != "" {
		ctx = context.WithValue(ctx, ParentSpanIDKey, parentID)
	}

	return ctx, span
}

// WithComponent adds a component name to the span
func WithComponent(span *Span, component string) *Span {
	span.Component = component
	return span
}

// WithField adds a field to the span
func WithField(span *Span, key string, value interface{}) *Span {
	span.Fields[key] = value
	return span
}

// WithFields adds multiple fields to the span
func WithFields(span *Span, fields map[string]interface{}) *Span {
	for k, v := range fields {
		span.Fields[k] = v
	}
	return span
}

// End marks the span as complete and records it
func End(span *Span) {
	span.EndTime = time.Now()
	span.Duration = span.EndTime.Sub(span.StartTime).Milliseconds()
	GetGlobalRecorder().Record(span)
}

// GetTraceID returns the trace ID from the context
func GetTraceID(ctx context.Context) string {
	if traceID, ok := ctx.Value(TraceIDKey).(string); ok {
		return traceID
	}
	return ""
}

// GetSpanID returns the current span ID from the context
func GetSpanID(ctx context.Context) string {
	if spanID, ok := ctx.Value(SpanIDKey).(string); ok {
		return spanID
	}
	return ""
}

// GetParentSpanID returns the parent span ID from the context
func GetParentSpanID(ctx context.Context) string {
	if parentID, ok := ctx.Value(ParentSpanIDKey).(string); ok {
		return parentID
	}
	return ""
}

// GetSpanContext returns all trace context as a map for logging
func GetSpanContext(ctx context.Context) map[string]interface{} {
	fields := make(map[string]interface{})
	if traceID := GetTraceID(ctx); traceID != "" {
		fields["trace_id"] = traceID
	}
	if spanID := GetSpanID(ctx); spanID != "" {
		fields["span_id"] = spanID
	}
	if parentID := GetParentSpanID(ctx); parentID != "" {
		fields["parent_span_id"] = parentID
	}
	return fields
}

// SpanBuilder provides a fluent API for building spans
type SpanBuilder struct {
	ctx       context.Context
	span      *Span
	component string
	fields    map[string]interface{}
}

// StartSpan begins a new span with the given name
func StartSpan(ctx context.Context, name string) *SpanBuilder {
	newCtx, span := WithSpan(ctx, name)
	return &SpanBuilder{
		ctx:    newCtx,
		span:   span,
		fields: make(map[string]interface{}),
	}
}

// Component sets the component name for the span
func (b *SpanBuilder) Component(component string) *SpanBuilder {
	b.component = component
	return b
}

// Field adds a field to the span
func (b *SpanBuilder) Field(key string, value interface{}) *SpanBuilder {
	b.fields[key] = value
	return b
}

// Fields adds multiple fields to the span
func (b *SpanBuilder) Fields(fields map[string]interface{}) *SpanBuilder {
	for k, v := range fields {
		b.fields[k] = v
	}
	return b
}

// End completes the span and returns the context
func (b *SpanBuilder) End() context.Context {
	if b.component != "" {
		b.span.Component = b.component
	}
	if len(b.fields) > 0 {
		b.span.Fields = b.fields
	}
	End(b.span)
	return b.ctx
}

// Context returns the context with the span
func (b *SpanBuilder) Context() context.Context {
	return b.ctx
}

// Span returns the span itself
func (b *SpanBuilder) Span() *Span {
	return b.span
}

// String returns a string representation of the span for logging
func (s *Span) String() string {
	parent := ""
	if s.ParentID != "" {
		parent = fmt.Sprintf(" (parent: %s)", s.ParentID[:8])
	}
	component := ""
	if s.Component != "" {
		component = fmt.Sprintf("[%s] ", s.Component)
	}
	return fmt.Sprintf("%sSpan %s%s %s", component, s.ID[:8], parent, s.Name)
}

// InMemoryRecorder records spans in memory for testing
type InMemoryRecorder struct {
	mu     sync.Mutex
	spans  []*Span
	maxSpans int
}

// NewInMemoryRecorder creates a new in-memory span recorder
func NewInMemoryRecorder(maxSpans int) *InMemoryRecorder {
	return &InMemoryRecorder{
		spans:    make([]*Span, 0, maxSpans),
		maxSpans: maxSpans,
	}
}

// Record records a span
func (r *InMemoryRecorder) Record(span *Span) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.spans = append(r.spans, span)
	// Keep only the most recent spans
	if len(r.spans) > r.maxSpans {
		r.spans = r.spans[len(r.spans)-r.maxSpans:]
	}
}

// GetSpans returns all recorded spans
func (r *InMemoryRecorder) GetSpans() []*Span {
	r.mu.Lock()
	defer r.mu.Unlock()
	copied := make([]*Span, len(r.spans))
	copy(copied, r.spans)
	return copied
}

// GetSpansByTraceID returns all spans for a given trace ID
func (r *InMemoryRecorder) GetSpansByTraceID(traceID string) []*Span {
	r.mu.Lock()
	defer r.mu.Unlock()
	var result []*Span
	for _, span := range r.spans {
		if span.TraceID == traceID {
			result = append(result, span)
		}
	}
	return result
}

// Clear clears all recorded spans
func (r *InMemoryRecorder) Clear() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.spans = make([]*Span, 0, r.maxSpans)
}
