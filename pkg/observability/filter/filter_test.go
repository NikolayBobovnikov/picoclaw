// Package filter provides log filtering capabilities for observability.
package filter

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLogFilter_Component(t *testing.T) {
	tests := []struct {
		name      string
		filter    LogFilter
		entries   []LogEntry
		wantCount int
	}{
		{
			name: "filters by component name",
			filter: LogFilter{
				Component: "agent",
			},
			entries: []LogEntry{
				{Level: "INFO", Component: "agent", Message: "Agent started"},
				{Level: "INFO", Component: "mcp", Message: "MCP connected"},
				{Level: "DEBUG", Component: "agent", Message: "Agent debug info"},
			},
			wantCount: 2,
		},
		{
			name: "returns all entries when component is empty",
			filter: LogFilter{
				Component: "",
			},
			entries: []LogEntry{
				{Level: "INFO", Component: "agent", Message: "Agent started"},
				{Level: "INFO", Component: "mcp", Message: "MCP connected"},
			},
			wantCount: 2,
		},
		{
			name: "returns no entries when component does not match",
			filter: LogFilter{
				Component: "nonexistent",
			},
			entries: []LogEntry{
				{Level: "INFO", Component: "agent", Message: "Agent started"},
			},
			wantCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var count int
			for _, entry := range tt.entries {
				if matchesFilter(entry, tt.filter) {
					count++
				}
			}
			assert.Equal(t, tt.wantCount, count)
		})
	}
}

func TestLogFilter_Level(t *testing.T) {
	tests := []struct {
		name      string
		filter    LogFilter
		entries   []LogEntry
		wantCount int
	}{
		{
			name: "filters by ERROR level and higher",
			filter: LogFilter{
				Level: "ERROR",
			},
			entries: []LogEntry{
				{Level: "DEBUG", Message: "Debug message"},
				{Level: "INFO", Message: "Info message"},
				{Level: "WARN", Message: "Warning message"},
				{Level: "ERROR", Message: "Error message"},
				{Level: "FATAL", Message: "Fatal message"},
			},
			wantCount: 2, // ERROR and FATAL
		},
		{
			name: "filters by WARN level and higher",
			filter: LogFilter{
				Level: "WARN",
			},
			entries: []LogEntry{
				{Level: "DEBUG", Message: "Debug"},
				{Level: "INFO", Message: "Info"},
				{Level: "WARN", Message: "Warning"},
				{Level: "ERROR", Message: "Error"},
			},
			wantCount: 2, // WARN and ERROR
		},
		{
			name: "filters by DEBUG level returns all",
			filter: LogFilter{
				Level: "DEBUG",
			},
			entries: []LogEntry{
				{Level: "DEBUG", Message: "Debug"},
				{Level: "INFO", Message: "Info"},
				{Level: "WARN", Message: "Warning"},
				{Level: "ERROR", Message: "Error"},
				{Level: "FATAL", Message: "Fatal"},
			},
			wantCount: 5,
		},
		{
			name: "invalid level returns no entries",
			filter: LogFilter{
				Level: "INVALID",
			},
			entries: []LogEntry{
				{Level: "INFO", Message: "Info"},
			},
			wantCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var count int
			for _, entry := range tt.entries {
				if matchesFilter(entry, tt.filter) {
					count++
				}
			}
			assert.Equal(t, tt.wantCount, count)
		})
	}
}

func TestLogFilter_TraceID(t *testing.T) {
	tests := []struct {
		name      string
		filter    LogFilter
		entries   []LogEntry
		wantCount int
	}{
		{
			name: "filters by trace ID in fields",
			filter: LogFilter{
				TraceID: "trace-123",
			},
			entries: []LogEntry{
				{
					Level:   "INFO",
					Message: "Request started",
					Fields:  map[string]interface{}{"trace_id": "trace-123"},
				},
				{
					Level:   "INFO",
					Message: "Other request",
					Fields:  map[string]interface{}{"trace_id": "trace-456"},
				},
				{
					Level:   "INFO",
					Message: "No trace ID",
				},
			},
			wantCount: 1,
		},
		{
			name: "returns no entries when trace ID not in fields",
			filter: LogFilter{
				TraceID: "trace-999",
			},
			entries: []LogEntry{
				{
					Level:   "INFO",
					Message: "Request",
					Fields:  map[string]interface{}{"trace_id": "trace-123"},
				},
			},
			wantCount: 0,
		},
		{
			name: "handles entries without fields",
			filter: LogFilter{
				TraceID: "trace-123",
			},
			entries: []LogEntry{
				{Level: "INFO", Message: "No fields"},
			},
			wantCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var count int
			for _, entry := range tt.entries {
				if matchesFilter(entry, tt.filter) {
					count++
				}
			}
			assert.Equal(t, tt.wantCount, count)
		})
	}
}

func TestLogFilter_Session(t *testing.T) {
	tests := []struct {
		name      string
		filter    LogFilter
		entries   []LogEntry
		wantCount int
	}{
		{
			name: "filters by session in fields",
			filter: LogFilter{
				Session: "session-abc",
			},
			entries: []LogEntry{
				{
					Level:   "INFO",
					Message: "User action",
					Fields:  map[string]interface{}{"session": "session-abc"},
				},
				{
					Level:   "INFO",
					Message: "Other user",
					Fields:  map[string]interface{}{"session": "session-def"},
				},
			},
			wantCount: 1,
		},
		{
			name: "handles entries without session field",
			filter: LogFilter{
				Session: "session-xyz",
			},
			entries: []LogEntry{
				{Level: "INFO", Message: "No session"},
			},
			wantCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var count int
			for _, entry := range tt.entries {
				if matchesFilter(entry, tt.filter) {
					count++
				}
			}
			assert.Equal(t, tt.wantCount, count)
		})
	}
}

func TestLogFilter_TimeRange(t *testing.T) {
	now := time.Now()
	past1h := now.Add(-1 * time.Hour)
	past2h := now.Add(-2 * time.Hour)
	future1h := now.Add(1 * time.Hour)

	tests := []struct {
		name      string
		filter    LogFilter
		entryTime time.Time
		wantMatch bool
	}{
		{
			name: "matches entry after since time (RFC3339)",
			filter: LogFilter{
				Since: past1h.Format(time.RFC3339Nano),
			},
			entryTime: now,
			wantMatch: true,
		},
		{
			name: "does not match entry before since time",
			filter: LogFilter{
				Since: past1h.Format(time.RFC3339Nano),
			},
			entryTime: past2h,
			wantMatch: false,
		},
		{
			name: "matches entry before until time",
			filter: LogFilter{
				Until: future1h.Format(time.RFC3339Nano),
			},
			entryTime: now,
			wantMatch: true,
		},
		{
			name: "does not match entry after until time",
			filter: LogFilter{
				Until: past1h.Format(time.RFC3339Nano),
			},
			entryTime: now,
			wantMatch: false,
		},
		{
			name: "matches entry in time range",
			filter: LogFilter{
				Since: past2h.Format(time.RFC3339Nano),
				Until: future1h.Format(time.RFC3339Nano),
			},
			entryTime: now,
			wantMatch: true,
		},
		{
			name: "matches entry with since duration",
			filter: LogFilter{
				Since: "1h",
			},
			entryTime: now.Add(-30 * time.Minute),
			wantMatch: true,
		},
		{
			name: "does not match entry outside since duration",
			filter: LogFilter{
				Since: "1h",
			},
			entryTime: now.Add(-2 * time.Hour),
			wantMatch: false,
		},
		{
			name: "handles entries without valid timestamp",
			filter: LogFilter{
				Since: past1h.Format(time.RFC3339Nano),
			},
			entryTime: time.Time{},
			wantMatch: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entry := LogEntry{
				Level:     "INFO",
				Timestamp: tt.entryTime.Format(time.RFC3339),
				Message:   "Test message",
			}

			result := matchesFilter(entry, tt.filter)
			assert.Equal(t, tt.wantMatch, result)
		})
	}
}

func TestLevelMatches(t *testing.T) {
	tests := []struct {
		name       string
		entryLevel string
		filterLevel string
		wantMatch  bool
	}{
		{
			name:       "ERROR matches ERROR filter",
			entryLevel: "ERROR",
			filterLevel: "ERROR",
			wantMatch:  true,
		},
		{
			name:       "FATAL matches ERROR filter",
			entryLevel: "FATAL",
			filterLevel: "ERROR",
			wantMatch:  true,
		},
		{
			name:       "WARN does not match ERROR filter",
			entryLevel: "WARN",
			filterLevel: "ERROR",
			wantMatch:  false,
		},
		{
			name:       "INFO matches DEBUG filter",
			entryLevel: "INFO",
			filterLevel: "DEBUG",
			wantMatch:  true,
		},
		{
			name:       "invalid entry level does not match",
			entryLevel: "INVALID",
			filterLevel: "INFO",
			wantMatch:  false,
		},
		{
			name:       "invalid filter level does not match",
			entryLevel: "INFO",
			filterLevel: "INVALID",
			wantMatch:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := levelMatches(tt.entryLevel, tt.filterLevel)
			assert.Equal(t, tt.wantMatch, result)
		})
	}
}

func TestFormatLogEntry(t *testing.T) {
	tests := []struct {
		name     string
		entry    LogEntry
		contains []string
	}{
		{
			name: "formats basic log entry",
			entry: LogEntry{
				Level:     "INFO",
				Timestamp: "2024-01-15T10:30:00Z",
				Message:   "Test message",
			},
			contains: []string{"[2024-01-15T10:30:00Z]", "[INFO]", "Test message"},
		},
		{
			name: "formats entry with component",
			entry: LogEntry{
				Level:     "ERROR",
				Timestamp: "2024-01-15T10:30:00Z",
				Component: "agent",
				Message:   "Error occurred",
			},
			contains: []string{"[ERROR]", "agent:", "Error occurred"},
		},
		{
			name: "formats entry with fields",
			entry: LogEntry{
				Level:     "DEBUG",
				Timestamp: "2024-01-15T10:30:00Z",
				Message:   "Debug info",
				Fields: map[string]interface{}{
					"user_id": "123",
					"count":   5,
				},
			},
			contains: []string{"{", "user_id=123", "count=5", "}"},
		},
		{
			name: "formats entry with caller",
			entry: LogEntry{
				Level:     "WARN",
				Timestamp: "2024-01-15T10:30:00Z",
				Message:   "Warning",
				Caller:    "handler.go:42",
			},
			contains: []string{"[WARN]", "Warning"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatLogEntry(tt.entry)
			for _, substr := range tt.contains {
				assert.Contains(t, result, substr)
			}
		})
	}
}

func TestFormatFields(t *testing.T) {
	tests := []struct {
		name     string
		fields   map[string]interface{}
		contains []string
	}{
		{
			name:     "empty fields",
			fields:   map[string]interface{}{},
			contains: []string{"{}"},
		},
		{
			name: "single field",
			fields: map[string]interface{}{
				"key": "value",
			},
			contains: []string{"key=value"},
		},
		{
			name: "multiple fields",
			fields: map[string]interface{}{
				"user": "john",
				"count": 42,
				"active": true,
			},
			contains: []string{"user=john", "count=42", "active=true"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatFields(tt.fields)
			for _, substr := range tt.contains {
				assert.Contains(t, result, substr)
			}
		})
	}
}

func TestFilterLogs(t *testing.T) {
	t.Run("filters log file by component", func(t *testing.T) {
		// Create temporary log file
		tmpDir := t.TempDir()
		logFile := filepath.Join(tmpDir, "test.log")

		entries := []string{
			`{"level":"INFO","timestamp":"2024-01-15T10:00:00Z","component":"agent","message":"Agent started"}`,
			`{"level":"INFO","timestamp":"2024-01-15T10:00:01Z","component":"mcp","message":"MCP connected"}`,
			`{"level":"DEBUG","timestamp":"2024-01-15T10:00:02Z","component":"agent","message":"Debug info"}`,
			`invalid json line`,
		}

		err := os.WriteFile(logFile, []byte(strings.Join(entries, "\n")), 0644)
		require.NoError(t, err)

		// Filter by component
		filter := LogFilter{Component: "agent"}
		results, err := FilterLogs(logFile, filter)
		require.NoError(t, err)
		assert.Len(t, results, 2)
		assert.Equal(t, "agent", results[0].Component)
		assert.Equal(t, "agent", results[1].Component)
	})

	t.Run("handles non-existent file", func(t *testing.T) {
		_, err := FilterLogs("/nonexistent/file.log", LogFilter{})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to open log file")
	})

	t.Run("skips invalid JSON lines", func(t *testing.T) {
		tmpDir := t.TempDir()
		logFile := filepath.Join(tmpDir, "mixed.log")

		content := strings.Join([]string{
			`{"level":"INFO","timestamp":"2024-01-15T10:00:00Z","message":"Valid"}`,
			`{invalid json}`,
			`{"level":"ERROR","timestamp":"2024-01-15T10:00:01Z","message":"Also valid"}`,
			`plain text line`,
		}, "\n")

		err := os.WriteFile(logFile, []byte(content), 0644)
		require.NoError(t, err)

		results, err := FilterLogs(logFile, LogFilter{})
		require.NoError(t, err)
		assert.Len(t, results, 2)
	})
}

func TestFilterLogs_Tail(t *testing.T) {
	t.Run("returns last N entries with tail", func(t *testing.T) {
		tmpDir := t.TempDir()
		logFile := filepath.Join(tmpDir, "tail.log")

		var entries []string
		for i := 0; i < 10; i++ {
			entries = append(entries, fmt.Sprintf(`{"level":"INFO","timestamp":"2024-01-15T10:00:%02dZ","message":"Entry %d"}`, i, i))
		}

		err := os.WriteFile(logFile, []byte(strings.Join(entries, "\n")), 0644)
		require.NoError(t, err)

		filter := LogFilter{Tail: true, Limit: 5}
		results, err := FilterLogs(logFile, filter)
		require.NoError(t, err)
		assert.Len(t, results, 5)
		assert.Contains(t, results[0].Message, "Entry 5")
		assert.Contains(t, results[4].Message, "Entry 9")
	})

	t.Run("uses default limit of 100 for tail without limit", func(t *testing.T) {
		tmpDir := t.TempDir()
		logFile := filepath.Join(tmpDir, "tail-default.log")

		var entries []string
		for i := 0; i < 150; i++ {
			entries = append(entries, fmt.Sprintf(`{"level":"INFO","timestamp":"2024-01-15T10:00:00Z","message":"Entry %d"}`, i))
		}

		err := os.WriteFile(logFile, []byte(strings.Join(entries, "\n")), 0644)
		require.NoError(t, err)

		filter := LogFilter{Tail: true}
		results, err := FilterLogs(logFile, filter)
		require.NoError(t, err)
		assert.Len(t, results, 100)
	})
}

func TestFilterFromReader(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		filter    LogFilter
		wantCount int
	}{
		{
			name: "filters by level from reader",
			input: `{"level":"INFO","timestamp":"2024-01-15T10:00:00Z","message":"Info"}
{"level":"ERROR","timestamp":"2024-01-15T10:00:01Z","message":"Error"}
{"level":"DEBUG","timestamp":"2024-01-15T10:00:02Z","message":"Debug"}`,
			filter:    LogFilter{Level: "ERROR"},
			wantCount: 1,
		},
		{
			name: "handles empty reader",
			input:     "",
			filter:    LogFilter{},
			wantCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := strings.NewReader(tt.input)
			results, err := FilterFromReader(reader, tt.filter)
			require.NoError(t, err)
			assert.Len(t, results, tt.wantCount)
		})
	}
}

func TestGetTraceLogs(t *testing.T) {
	tmpDir := t.TempDir()
	logFile := filepath.Join(tmpDir, "trace.log")

	entries := []string{
		`{"level":"INFO","timestamp":"2024-01-15T10:00:00Z","message":"Request","fields":{"trace_id":"trace-123"}}`,
		`{"level":"INFO","timestamp":"2024-01-15T10:00:01Z","message":"Other request","fields":{"trace_id":"trace-456"}}`,
		`{"level":"INFO","timestamp":"2024-01-15T10:00:02Z","message":"No trace"}`,
	}

	err := os.WriteFile(logFile, []byte(strings.Join(entries, "\n")), 0644)
	require.NoError(t, err)

	results, err := GetTraceLogs(logFile, "trace-123")
	require.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Contains(t, results[0].Fields["trace_id"], "trace-123")
}

func TestGetComponentLogs(t *testing.T) {
	tmpDir := t.TempDir()
	logFile := filepath.Join(tmpDir, "component.log")

	entries := []string{
		`{"level":"INFO","timestamp":"2024-01-15T10:00:00Z","component":"agent","message":"Agent log"}`,
		`{"level":"INFO","timestamp":"2024-01-15T10:00:01Z","component":"mcp","message":"MCP log"}`,
	}

	err := os.WriteFile(logFile, []byte(strings.Join(entries, "\n")), 0644)
	require.NoError(t, err)

	results, err := GetComponentLogs(logFile, "agent")
	require.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, "agent", results[0].Component)
}

func TestGetSessionLogs(t *testing.T) {
	tmpDir := t.TempDir()
	logFile := filepath.Join(tmpDir, "session.log")

	entries := []string{
		`{"level":"INFO","timestamp":"2024-01-15T10:00:00Z","message":"User action","fields":{"session":"session-abc"}}`,
		`{"level":"INFO","timestamp":"2024-01-15T10:00:01Z","message":"Other user","fields":{"session":"session-def"}}`,
	}

	err := os.WriteFile(logFile, []byte(strings.Join(entries, "\n")), 0644)
	require.NoError(t, err)

	results, err := GetSessionLogs(logFile, "session-abc")
	require.NoError(t, err)
	assert.Len(t, results, 1)
}

func TestGetRecentLogs(t *testing.T) {
	tmpDir := t.TempDir()
	logFile := filepath.Join(tmpDir, "recent.log")

	// Create entries with recent timestamps
	now := time.Now()
	var entries []string
	for i := 0; i < 5; i++ {
		timestamp := now.Add(-time.Duration(i) * time.Minute).Format(time.RFC3339)
		entries = append(entries, fmt.Sprintf(`{"level":"INFO","timestamp":"%s","message":"Entry %d"}`, timestamp, i))
	}

	err := os.WriteFile(logFile, []byte(strings.Join(entries, "\n")), 0644)
	require.NoError(t, err)

	results, err := GetRecentLogs(logFile, 30*time.Minute)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(results), 3) // At least 3 of 5 should be within 30 minutes
}

func TestCombinedFilters(t *testing.T) {
	// Test that multiple filters work together
	filter := LogFilter{
		Component: "agent",
		Level:     "ERROR",
		TraceID:   "trace-123",
	}

	entries := []LogEntry{
		{
			Level:     "ERROR",
			Component: "agent",
			Message:   "Agent error with trace",
			Fields:    map[string]interface{}{"trace_id": "trace-123"},
		},
		{
			Level:     "ERROR",
			Component: "agent",
			Message:   "Agent error without trace",
		},
		{
			Level:     "INFO",
			Component: "agent",
			Message:   "Agent info with trace",
			Fields:    map[string]interface{}{"trace_id": "trace-123"},
		},
		{
			Level:     "ERROR",
			Component: "mcp",
			Message:   "MCP error with trace",
			Fields:    map[string]interface{}{"trace_id": "trace-123"},
		},
	}

	var count int
	for _, entry := range entries {
		if matchesFilter(entry, filter) {
			count++
		}
	}

	assert.Equal(t, 1, count, "only first entry should match all filters")
}

