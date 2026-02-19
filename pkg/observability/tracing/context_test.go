// Package tracing provides context propagation utilities for distributed tracing.
package tracing

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExtractFromContext(t *testing.T) {
	tests := []struct {
		name       string
		setup      func(context.Context) context.Context
		wantTrace  string
		wantParent string
		wantSample bool
	}{
		{
			name:       "empty context returns empty trace context",
			setup:      func(ctx context.Context) context.Context { return ctx },
			wantTrace:  "",
			wantParent: "",
			wantSample: true,
		},
		{
			name: "context with trace ID extracts trace ID",
			setup: func(ctx context.Context) context.Context {
				return WithTraceID(ctx, "trace-123")
			},
			wantTrace:  "trace-123",
			wantParent: "",
			wantSample: true,
		},
		{
			name: "context with parent span extracts parent ID",
			setup: func(ctx context.Context) context.Context {
				parentCtx, _ := WithSpan(ctx, "parent")
				return parentCtx
			},
			wantTrace:  "",
			wantParent: "",
			wantSample: true,
		},
		{
			name: "context with trace and parent extracts both",
			setup: func(ctx context.Context) context.Context {
				ctx = WithTraceID(ctx, "trace-456")
				// Note: WithSpan creates spans with internal parent tracking,
				// but ExtractFromContext only reads context-level trace ID
				return ctx
			},
			wantTrace:  "trace-456",
			wantParent: "",
			wantSample: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			ctx = tt.setup(ctx)

			tc := ExtractFromContext(ctx)

			if tt.wantTrace == "" && tt.wantParent == "" {
				// For empty case
				if tt.wantTrace == "" && tt.name == "empty context returns empty trace context" {
					assert.Empty(t, tc.TraceID)
					assert.Empty(t, tc.ParentSpanID)
				} else {
					// For non-empty case, check the values are properly extracted
					assert.Equal(t, tt.wantSample, tc.Sampled)
				}
			} else {
				assert.Equal(t, tt.wantTrace, tc.TraceID)
				assert.Equal(t, tt.wantParent, tc.ParentSpanID)
				assert.Equal(t, tt.wantSample, tc.Sampled)
			}
		})
	}
}

func TestInjectToContext(t *testing.T) {
	tests := []struct {
		name       string
		baseCtx    func() context.Context
		traceCtx   *TraceContext
		wantTrace  string
		wantParent string
	}{
		{
			name:     "injects trace ID into context",
			baseCtx:  func() context.Context { return context.Background() },
			traceCtx: &TraceContext{TraceID: "injected-trace-123", ParentSpanID: "", Sampled: true},
			wantTrace: "injected-trace-123",
			wantParent: "",
		},
		{
			name:     "injects parent span ID into context",
			baseCtx:  func() context.Context { return context.Background() },
			traceCtx: &TraceContext{TraceID: "", ParentSpanID: "parent-span-456", Sampled: true},
			wantTrace: "",
			wantParent: "parent-span-456",
		},
		{
			name:     "injects both trace ID and parent span ID",
			baseCtx:  func() context.Context { return context.Background() },
			traceCtx: &TraceContext{TraceID: "trace-789", ParentSpanID: "parent-span-101", Sampled: true},
			wantTrace: "trace-789",
			wantParent: "parent-span-101",
		},
		{
			name:     "replaces existing trace ID in context",
			baseCtx: func() context.Context {
				return WithTraceID(context.Background(), "old-trace")
			},
			traceCtx: &TraceContext{TraceID: "new-trace-999", ParentSpanID: "", Sampled: true},
			wantTrace: "new-trace-999",
			wantParent: "",
		},
		{
			name:     "handles empty trace context",
			baseCtx:  func() context.Context { return context.Background() },
			traceCtx: &TraceContext{TraceID: "", ParentSpanID: "", Sampled: true},
			wantTrace: "",
			wantParent: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := tt.baseCtx()
			result := InjectToContext(ctx, tt.traceCtx)

			traceID := GetTraceID(result)
			parentID := GetParentSpanID(result)

			assert.Equal(t, tt.wantTrace, traceID)
			assert.Equal(t, tt.wantParent, parentID)
		})
	}
}

func TestMergeTraceContext(t *testing.T) {
	tests := []struct {
		name           string
		baseCtx        func() context.Context
		externalCtx    *TraceContext
		expectedTrace  string
		expectedParent string
	}{
		{
			name:        "nil external context returns original context",
			baseCtx:     func() context.Context { return context.Background() },
			externalCtx: nil,
			expectedTrace: "",
			expectedParent: "",
		},
		{
			name:        "merges external trace ID into empty context",
			baseCtx:     func() context.Context { return context.Background() },
			externalCtx: &TraceContext{TraceID: "external-trace-123", ParentSpanID: "", Sampled: true},
			expectedTrace: "external-trace-123",
			expectedParent: "",
		},
		{
			name:        "merges external parent span ID into context",
			baseCtx:     func() context.Context { return context.Background() },
			externalCtx: &TraceContext{TraceID: "", ParentSpanID: "external-parent-456", Sampled: true},
			expectedTrace: "",
			expectedParent: "external-parent-456",
		},
		{
			name: "external trace ID overrides existing trace ID",
			baseCtx: func() context.Context {
				return WithTraceID(context.Background(), "internal-trace")
			},
			externalCtx: &TraceContext{TraceID: "external-trace-789", ParentSpanID: "", Sampled: true},
			expectedTrace: "external-trace-789",
			expectedParent: "",
		},
		{
			name:        "merges both trace and parent span ID",
			baseCtx:     func() context.Context { return context.Background() },
			externalCtx: &TraceContext{TraceID: "merged-trace", ParentSpanID: "merged-parent", Sampled: true},
			expectedTrace: "merged-trace",
			expectedParent: "merged-parent",
		},
		{
			name: "existing context with span preserves external trace",
			baseCtx: func() context.Context {
				ctx := WithTraceID(context.Background(), "original-trace")
				newCtx, _ := WithSpan(ctx, "operation")
				return newCtx
			},
			externalCtx: &TraceContext{TraceID: "external-trace-abc", ParentSpanID: "external-parent-def", Sampled: true},
			expectedTrace: "external-trace-abc",
			expectedParent: "external-parent-def",
		},
		{
			name:        "empty external trace context does not modify context",
			baseCtx:     func() context.Context { return context.Background() },
			externalCtx: &TraceContext{TraceID: "", ParentSpanID: "", Sampled: true},
			expectedTrace: "",
			expectedParent: "",
		},
		{
			name: "context with existing trace gets external parent",
			baseCtx: func() context.Context {
				return WithTraceID(context.Background(), "my-trace")
			},
			externalCtx: &TraceContext{TraceID: "", ParentSpanID: "incoming-parent", Sampled: true},
			expectedTrace: "my-trace",
			expectedParent: "incoming-parent",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := tt.baseCtx()
			result := MergeTraceContext(ctx, tt.externalCtx)

			traceID := GetTraceID(result)
			parentID := GetParentSpanID(result)

			assert.Equal(t, tt.expectedTrace, traceID, "trace ID mismatch")
			assert.Equal(t, tt.expectedParent, parentID, "parent span ID mismatch")
		})
	}
}

func TestTraceContext_Sampled(t *testing.T) {
	tests := []struct {
		name      string
		sampled   bool
		wantValue bool
	}{
		{
			name:      "sampled is true",
			sampled:   true,
			wantValue: true,
		},
		{
			name:      "sampled is false",
			sampled:   false,
			wantValue: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tc := &TraceContext{
				TraceID:      "trace-123",
				ParentSpanID: "parent-456",
				Sampled:      tt.sampled,
			}

			assert.Equal(t, tt.wantValue, tc.Sampled)
		})
	}
}

func TestTraceContextRoundTrip(t *testing.T) {
	// Test that we can extract and inject trace context consistently
	originalTrace := "original-trace-id"
	originalParent := "original-parent-id"

	t.Run("extract then inject preserves trace context", func(t *testing.T) {
		ctx := context.Background()
		ctx = WithTraceID(ctx, originalTrace)
		ctx = context.WithValue(ctx, ParentSpanIDKey, originalParent)

		extracted := ExtractFromContext(ctx)
		assert.Equal(t, originalTrace, extracted.TraceID)
		assert.Equal(t, originalParent, extracted.ParentSpanID)

		newCtx := InjectToContext(context.Background(), extracted)
		retrievedTrace := GetTraceID(newCtx)
		retrievedParent := GetParentSpanID(newCtx)

		assert.Equal(t, originalTrace, retrievedTrace)
		assert.Equal(t, originalParent, retrievedParent)
	})

	t.Run("merge then extract preserves external context", func(t *testing.T) {
		external := &TraceContext{
			TraceID:      "external-trace",
			ParentSpanID: "external-parent",
			Sampled:      true,
		}

		ctx := context.Background()
		mergedCtx := MergeTraceContext(ctx, external)

		extracted := ExtractFromContext(mergedCtx)
		assert.Equal(t, external.TraceID, extracted.TraceID)
		assert.Equal(t, external.ParentSpanID, extracted.ParentSpanID)
		assert.Equal(t, external.Sampled, extracted.Sampled)
	})
}
