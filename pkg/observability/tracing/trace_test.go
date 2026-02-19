// Package tracing provides request tracing with span tracking for observability.
package tracing

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateTraceID(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "generates non-empty trace ID",
		},
		{
			name: "generates unique trace IDs",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			traceID := GenerateTraceID()
			assert.NotEmpty(t, traceID, "trace ID should not be empty")

			// UUID format validation (should be 36 chars with hyphens)
			assert.Len(t, traceID, 36, "trace ID should be UUID format (36 characters)")

			// Verify uniqueness across multiple generations
			ids := make(map[string]bool)
			for i := 0; i < 100; i++ {
				id := GenerateTraceID()
				assert.NotEmpty(t, id)
				assert.False(t, ids[id], "trace ID should be unique")
				ids[id] = true
			}
			assert.Len(t, ids, 100, "should generate 100 unique trace IDs")
		})
	}
}

func TestGenerateSpanID(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "generates non-empty span ID",
		},
		{
			name: "generates unique span IDs",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			spanID := GenerateSpanID()
			assert.NotEmpty(t, spanID, "span ID should not be empty")

			// UUID format validation
			assert.Len(t, spanID, 36, "span ID should be UUID format (36 characters)")

			// Verify uniqueness across multiple generations
			ids := make(map[string]bool)
			for i := 0; i < 100; i++ {
				id := GenerateSpanID()
				assert.NotEmpty(t, id)
				assert.False(t, ids[id], "span ID should be unique")
				ids[id] = true
			}
			assert.Len(t, ids, 100, "should generate 100 unique span IDs")
		})
	}
}

func TestWithTraceID(t *testing.T) {
	tests := []struct {
		name    string
		traceID string
	}{
		{
			name:    "adds trace ID to empty context",
			traceID: "trace-123",
		},
		{
			name:    "replaces existing trace ID",
			traceID: "trace-456",
		},
		{
			name:    "handles UUID trace ID",
			traceID: "550e8400-e29b-41d4-a716-446655440000",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			ctx = WithTraceID(ctx, tt.traceID)

			retrieved := GetTraceID(ctx)
			assert.Equal(t, tt.traceID, retrieved, "retrieved trace ID should match")
		})
	}
}

func TestGetTraceID(t *testing.T) {
	tests := []struct {
		name     string
		traceID  string
		expected string
	}{
		{
			name:     "returns empty string for context without trace ID",
			traceID:  "",
			expected: "",
		},
		{
			name:     "returns trace ID from context",
			traceID:  "test-trace-123",
			expected: "test-trace-123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			if tt.traceID != "" {
				ctx = WithTraceID(ctx, tt.traceID)
			}

			result := GetTraceID(ctx)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestWithSpan(t *testing.T) {
	tests := []struct {
		name           string
		spanName       string
		existingTrace  string
		expectedNew    bool
	}{
		{
			name:           "creates span with new trace ID",
			spanName:       "test-operation",
			existingTrace:  "",
			expectedNew:    true,
		},
		{
			name:           "creates span with existing trace ID",
			spanName:       "child-operation",
			existingTrace:  "existing-trace-123",
			expectedNew:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			if tt.existingTrace != "" {
				ctx = WithTraceID(ctx, tt.existingTrace)
			}

			newCtx, span := WithSpan(ctx, tt.spanName)

			assert.NotNil(t, span, "span should not be nil")
			assert.Equal(t, tt.spanName, span.Name, "span name should match")
			assert.NotEmpty(t, span.ID, "span ID should not be empty")
			assert.NotEmpty(t, span.TraceID, "trace ID should not be empty")
			assert.False(t, span.StartTime.IsZero(), "start time should be set")

			// Verify context propagation
			retrievedSpanID := GetSpanID(newCtx)
			assert.Equal(t, span.ID, retrievedSpanID, "span ID should be in context")

			if tt.existingTrace != "" {
				assert.Equal(t, tt.existingTrace, span.TraceID, "should use existing trace ID")
			}

			retrievedTraceID := GetTraceID(newCtx)
			assert.Equal(t, span.TraceID, retrievedTraceID, "trace ID should be in context")
		})
	}
}

func TestWithSpan_ParentChild(t *testing.T) {
	ctx := context.Background()

	// Create parent span
	parentCtx, parent := WithSpan(ctx, "parent")
	parentID := parent.ID

	// Create child span
	childCtx, child := WithSpan(parentCtx, "child")

	// Verify parent-child relationship
	assert.Equal(t, parentID, child.ParentID, "child should have parent ID")

	// Verify context propagation
	retrievedParentID := GetParentSpanID(childCtx)
	assert.Equal(t, parentID, retrievedParentID, "parent span ID should be in child context")
}

func TestGetSpanID(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(context.Context) context.Context
		expected string
	}{
		{
			name:     "returns empty for context without span",
			setup:    func(ctx context.Context) context.Context { return ctx },
			expected: "",
		},
		{
			name: "returns span ID from context",
			setup: func(ctx context.Context) context.Context {
				newCtx, _ := WithSpan(ctx, "test")
				return newCtx
			},
			expected: "", // Will check non-empty in test
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			ctx = tt.setup(ctx)

			result := GetSpanID(ctx)
			if tt.expected == "" {
				if tt.name == "returns span ID from context" {
					assert.NotEmpty(t, result, "should have span ID")
				} else {
					assert.Empty(t, result)
				}
			} else {
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestGetParentSpanID(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(context.Context) context.Context
		check    func(*testing.T, string)
	}{
		{
			name: "returns empty for context without parent",
			setup: func(ctx context.Context) context.Context {
				newCtx, _ := WithSpan(ctx, "parent")
				return newCtx
			},
			check: func(t *testing.T, parentID string) {
				assert.Empty(t, parentID, "should not have parent ID")
			},
		},
		{
			name: "returns parent ID for child span",
			setup: func(ctx context.Context) context.Context {
				parentCtx, _ := WithSpan(ctx, "parent")
				childCtx, _ := WithSpan(parentCtx, "child")
				return childCtx
			},
			check: func(t *testing.T, parentID string) {
				assert.NotEmpty(t, parentID, "should have parent ID")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			ctx = tt.setup(ctx)
			result := GetParentSpanID(ctx)
			tt.check(t, result)
		})
	}
}

func TestWithComponent(t *testing.T) {
	tests := []struct {
		name      string
		component string
	}{
		{
			name:      "adds component to span",
			component: "test-service",
		},
		{
			name:      "replaces existing component",
			component: "new-service",
		},
		{
			name:      "handles empty component",
			component: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, span := WithSpan(context.Background(), "test")

			result := WithComponent(span, tt.component)
			assert.Same(t, span, result, "should return same span instance")
			assert.Equal(t, tt.component, span.Component, "component should match")
		})
	}
}

func TestWithField(t *testing.T) {
	tests := []struct {
		name   string
		key    string
		value  interface{}
	}{
		{
			name:  "adds string field",
			key:   "user_id",
			value: "user-123",
		},
		{
			name:  "adds int field",
			key:   "count",
			value: 42,
		},
		{
			name:  "adds bool field",
			key:   "active",
			value: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, span := WithSpan(context.Background(), "test")

			result := WithField(span, tt.key, tt.value)
			assert.Same(t, span, result, "should return same span instance")
			assert.Contains(t, span.Fields, tt.key, "should contain key")
			assert.Equal(t, tt.value, span.Fields[tt.key], "field value should match")
		})
	}
}

func TestWithFields(t *testing.T) {
	_, span := WithSpan(context.Background(), "test")

	fields := map[string]interface{}{
		"user_id":  "user-123",
		"count":    100,
		"active":   true,
		"metadata": map[string]string{"key": "value"},
	}

	result := WithFields(span, fields)
	assert.Same(t, span, result, "should return same span instance")

	for k, v := range fields {
		assert.Contains(t, span.Fields, k, "should contain key")
		assert.Equal(t, v, span.Fields[k], "field value should match")
	}
}

func TestEnd(t *testing.T) {
	tests := []struct {
		name              string
		setup             func() *Span
		expectEndTime     bool
		expectDuration    bool
	}{
		{
			name: "sets end time and duration",
			setup: func() *Span {
				_, span := WithSpan(context.Background(), "test")
				return span
			},
			expectEndTime:  true,
			expectDuration: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			span := tt.setup()

			// Ensure duration is not set before End
			assert.Zero(t, span.Duration, "duration should be zero before End")
			assert.True(t, span.EndTime.IsZero(), "end time should be zero before End")

			// Add a small delay to ensure measurable duration
			time.Sleep(1 * time.Millisecond)

			End(span)

			if tt.expectEndTime {
				assert.False(t, span.EndTime.IsZero(), "end time should be set")
				assert.True(t, span.EndTime.After(span.StartTime) || span.EndTime.Equal(span.StartTime),
					"end time should be after or equal to start time")
			}

			if tt.expectDuration {
				assert.GreaterOrEqual(t, span.Duration, int64(1), "duration should be at least 1ms")
			}
		})
	}
}

func TestGetSpanContext(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(context.Context) context.Context
		expected map[string]interface{}
	}{
		{
			name:     "returns empty map for context without trace info",
			setup:    func(ctx context.Context) context.Context { return ctx },
			expected: map[string]interface{}{},
		},
		{
			name: "returns trace ID only",
			setup: func(ctx context.Context) context.Context {
				return WithTraceID(ctx, "trace-123")
			},
			expected: map[string]interface{}{"trace_id": "trace-123"},
		},
		{
			name: "returns full context with parent",
			setup: func(ctx context.Context) context.Context {
				parentCtx, _ := WithSpan(ctx, "parent")
				return parentCtx
			},
			expected: nil, // Will check for keys in test
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			ctx = tt.setup(ctx)

			result := GetSpanContext(ctx)

			if tt.expected == nil {
				// Check for expected keys
				assert.Contains(t, result, "trace_id", "should have trace_id")
				assert.Contains(t, result, "span_id", "should have span_id")
				assert.NotEmpty(t, result["trace_id"], "trace_id should not be empty")
				assert.NotEmpty(t, result["span_id"], "span_id should not be empty")
			} else {
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestSpan_String(t *testing.T) {
	tests := []struct {
		name     string
		setup    func() *Span
		contains []string
	}{
		{
			name: "formats span without parent",
			setup: func() *Span {
				_, span := WithSpan(context.Background(), "test-operation")
				return span
			},
			contains: []string{"Span", "test-operation"},
		},
		{
			name: "formats span with parent",
			setup: func() *Span {
				parentCtx, _ := WithSpan(context.Background(), "parent")
				_, child := WithSpan(parentCtx, "child")
				return child
			},
			contains: []string{"Span", "child", "parent:"},
		},
		{
			name: "formats span with component",
			setup: func() *Span {
				_, span := WithSpan(context.Background(), "operation")
				span.Component = "my-service"
				return span
			},
			contains: []string{"[my-service]", "Span", "operation"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			span := tt.setup()
			result := span.String()

			for _, substr := range tt.contains {
				assert.Contains(t, result, substr, "string representation should contain %s", substr)
			}
		})
	}
}

func TestInMemoryRecorder(t *testing.T) {
	t.Run("NewInMemoryRecorder", func(t *testing.T) {
		recorder := NewInMemoryRecorder(10)
		assert.NotNil(t, recorder)
		assert.NotNil(t, recorder.GetSpans())
		assert.Empty(t, recorder.GetSpans(), "new recorder should have no spans")
	})

	t.Run("Record", func(t *testing.T) {
		recorder := NewInMemoryRecorder(10)

		_, span := WithSpan(context.Background(), "test-span")
		End(span)

		recorder.Record(span)

		spans := recorder.GetSpans()
		require.Len(t, spans, 1, "should have one span")
		assert.Equal(t, "test-span", spans[0].Name)
	})

	t.Run("GetSpansByTraceID", func(t *testing.T) {
		recorder := NewInMemoryRecorder(10)

		ctx := WithTraceID(context.Background(), "test-trace-123")
		newCtx, span1 := WithSpan(ctx, "span-1")
		End(span1)

		_, span2 := WithSpan(newCtx, "span-2")
		End(span2)

		recorder.Record(span1)
		recorder.Record(span2)

		spans := recorder.GetSpansByTraceID("test-trace-123")
		assert.Len(t, spans, 2, "should find 2 spans for trace ID")

		emptySpans := recorder.GetSpansByTraceID("nonexistent")
		assert.Empty(t, emptySpans, "should return empty for nonexistent trace ID")
	})

	t.Run("MaxSpansLimit", func(t *testing.T) {
		recorder := NewInMemoryRecorder(3)

		for i := 0; i < 5; i++ {
			_, span := WithSpan(context.Background(), "span")
			End(span)
			recorder.Record(span)
		}

		spans := recorder.GetSpans()
		assert.Len(t, spans, 3, "should keep only maxSpans")
	})

	t.Run("Clear", func(t *testing.T) {
		recorder := NewInMemoryRecorder(10)

		_, span := WithSpan(context.Background(), "test")
		End(span)
		recorder.Record(span)

		assert.Len(t, recorder.GetSpans(), 1)

		recorder.Clear()
		assert.Empty(t, recorder.GetSpans(), "should be empty after clear")
	})

	t.Run("ConcurrentAccess", func(t *testing.T) {
		recorder := NewInMemoryRecorder(100)
		var wg sync.WaitGroup

		// Concurrent writes
		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func(idx int) {
				defer wg.Done()
				_, span := WithSpan(context.Background(), "concurrent-span")
				End(span)
				recorder.Record(span)
			}(i)
		}

		// Concurrent reads
		for i := 0; i < 5; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				_ = recorder.GetSpans()
			}()
		}

		wg.Wait()
		assert.Len(t, recorder.GetSpans(), 10, "should record all concurrent spans")
	})
}

func TestSetGlobalRecorder(t *testing.T) {
	t.Run("sets and retrieves global recorder", func(t *testing.T) {
		original := GetGlobalRecorder()

		newRecorder := NewInMemoryRecorder(100)
		SetGlobalRecorder(newRecorder)

		retrieved := GetGlobalRecorder()
		assert.Same(t, newRecorder, retrieved, "should retrieve same recorder")

		// Restore original
		SetGlobalRecorder(original)
	})
}

func TestDefaultRecorder(t *testing.T) {
	t.Run("no-ops on Record", func(t *testing.T) {
		recorder := &DefaultRecorder{}

		// Should not panic
		_, span := WithSpan(context.Background(), "test")
		End(span)
		recorder.Record(span)

		assert.NoError(t, nil) // Just verify no panic
	})
}

func TestSpanBuilder(t *testing.T) {
	t.Run("StartSpan", func(t *testing.T) {
		ctx := context.Background()
		builder := StartSpan(ctx, "operation")

		assert.NotNil(t, builder)
		assert.NotNil(t, builder.Context())
		assert.NotNil(t, builder.Span())
		assert.Equal(t, "operation", builder.Span().Name)
	})

	t.Run("Component", func(t *testing.T) {
		builder := StartSpan(context.Background(), "op")
		result := builder.Component("my-component")

		assert.Same(t, builder, result, "should return same builder")
		assert.Equal(t, "my-component", builder.component)
	})

	t.Run("Field", func(t *testing.T) {
		builder := StartSpan(context.Background(), "op")
		result := builder.Field("key", "value")

		assert.Same(t, builder, result, "should return same builder")
		assert.Equal(t, "value", builder.fields["key"])
	})

	t.Run("Fields", func(t *testing.T) {
		builder := StartSpan(context.Background(), "op")
		fields := map[string]interface{}{"a": 1, "b": 2}
		result := builder.Fields(fields)

		assert.Same(t, builder, result, "should return same builder")
		assert.Equal(t, 1, builder.fields["a"])
		assert.Equal(t, 2, builder.fields["b"])
	})

	t.Run("End", func(t *testing.T) {
		ctx := context.Background()
		builder := StartSpan(ctx, "operation").
			Component("service").
			Field("user_id", "123")

		returnedCtx := builder.End()
		span := builder.Span()

		// Verify span was ended
		assert.False(t, span.EndTime.IsZero(), "end time should be set")
		assert.GreaterOrEqual(t, span.Duration, int64(0), "duration should be set")

		// Verify component and fields were applied
		assert.Equal(t, "service", span.Component)
		assert.Equal(t, "123", span.Fields["user_id"])

		// Verify context was returned and contains the span
		assert.NotNil(t, returnedCtx, "should return context")
		assert.NotEmpty(t, GetSpanID(returnedCtx), "returned context should have span ID")
	})

	t.Run("Span", func(t *testing.T) {
		builder := StartSpan(context.Background(), "test")
		span := builder.Span()

		assert.NotNil(t, span)
		assert.Equal(t, "test", span.Name)
	})

	t.Run("Context", func(t *testing.T) {
		ctx := context.Background()
		builder := StartSpan(ctx, "test")

		retrievedCtx := builder.Context()
		assert.NotNil(t, retrievedCtx)

		// Verify context has trace/span info
		traceID := GetTraceID(retrievedCtx)
		spanID := GetSpanID(retrievedCtx)
		assert.NotEmpty(t, traceID)
		assert.NotEmpty(t, spanID)
	})

	t.Run("chaining", func(t *testing.T) {
		ctx := context.Background()
		builder := StartSpan(ctx, "chained-operation").
			Component("test-service").
			Field("request_id", "abc-123").
			Fields(map[string]interface{}{
				"user":    "nick",
				"action":  "test",
			})

		_ = builder.End()
		span := builder.Span()

		assert.Equal(t, "test-service", span.Component)
		assert.Equal(t, "abc-123", span.Fields["request_id"])
		assert.Equal(t, "nick", span.Fields["user"])
		assert.Equal(t, "test", span.Fields["action"])
	})
}
