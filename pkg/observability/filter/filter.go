// Package filter provides log filtering capabilities for observability.
package filter

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"time"
)

// LogFilter defines criteria for filtering log entries
type LogFilter struct {
	Component string
	Level     string
	TraceID   string
	Session   string
	Since     string
	Until     string
	Tail      bool
	Limit     int
}

// LogEntry represents a parsed log entry from JSON logs
type LogEntry struct {
	Level     string                 `json:"level"`
	Timestamp string                 `json:"timestamp"`
	Component string                 `json:"component,omitempty"`
	Message   string                 `json:"message"`
	Fields    map[string]interface{} `json:"fields,omitempty"`
	Caller    string                 `json:"caller,omitempty"`
}

// FilterLogs reads log entries from a file and applies the filter
func FilterLogs(filePath string, filter LogFilter) ([]LogEntry, error) {
	var entries []LogEntry

	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		var entry LogEntry
		if err := json.Unmarshal([]byte(line), &entry); err != nil {
			// Skip invalid JSON lines
			continue
		}

		if matchesFilter(entry, filter) {
			entries = append(entries, entry)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading log file: %w", err)
	}

	// Handle tail/limit
	if filter.Tail || filter.Limit > 0 {
		limit := filter.Limit
		if limit == 0 {
			limit = 100 // Default tail limit
		}
		if len(entries) > limit {
			entries = entries[len(entries)-limit:]
		}
	}

	return entries, nil
}

// matchesFilter checks if a log entry matches the filter criteria
func matchesFilter(entry LogEntry, filter LogFilter) bool {
	// Filter by component
	if filter.Component != "" && entry.Component != filter.Component {
		return false
	}

	// Filter by level
	if filter.Level != "" && !levelMatches(entry.Level, filter.Level) {
		return false
	}

	// Filter by trace ID (in fields)
	if filter.TraceID != "" {
		if traceID, ok := entry.Fields["trace_id"].(string); !ok || traceID != filter.TraceID {
			return false
		}
	}

	// Filter by session key (in fields)
	if filter.Session != "" {
		if session, ok := entry.Fields["session"].(string); !ok || session != filter.Session {
			return false
		}
	}

	// Filter by time range
	if filter.Since != "" || filter.Until != "" {
		entryTime, err := time.Parse(time.RFC3339, entry.Timestamp)
		if err != nil {
			return false
		}

		if filter.Since != "" {
			sinceTime, err := time.Parse(time.RFC3339Nano, filter.Since)
			if err != nil {
				// Try parsing as duration (e.g., "1h")
				if duration, err := time.ParseDuration(filter.Since); err == nil {
					sinceTime = time.Now().Add(-duration)
				} else {
					return false
				}
			}
			if entryTime.Before(sinceTime) {
				return false
			}
		}

		if filter.Until != "" {
			untilTime, err := time.Parse(time.RFC3339Nano, filter.Until)
			if err != nil {
				return false
			}
			if entryTime.After(untilTime) {
				return false
			}
		}
	}

	return true
}

// levelMatches checks if the entry level matches or is higher severity than the filter level
func levelMatches(entryLevel, filterLevel string) bool {
	levels := map[string]int{
		"DEBUG": 0,
		"INFO":  1,
		"WARN":  2,
		"ERROR": 3,
		"FATAL": 4,
	}

	entrySeverity, ok := levels[entryLevel]
	if !ok {
		return false
	}

	filterSeverity, ok := levels[filterLevel]
	if !ok {
		return false
	}

	return entrySeverity >= filterSeverity
}

// FormatLogEntry formats a log entry for human-readable output
func FormatLogEntry(entry LogEntry) string {
	var parts []string

	// Timestamp
	parts = append(parts, fmt.Sprintf("[%s]", entry.Timestamp))

	// Level
	parts = append(parts, fmt.Sprintf("[%s]", entry.Level))

	// Component
	if entry.Component != "" {
		parts = append(parts, fmt.Sprintf("%s:", entry.Component))
	}

	// Message
	parts = append(parts, entry.Message)

	// Fields
	if len(entry.Fields) > 0 {
		fieldStr := formatFields(entry.Fields)
		parts = append(parts, fieldStr)
	}

	return strings.Join(parts, " ")
}

// formatFields formats fields map as key=value pairs
func formatFields(fields map[string]interface{}) string {
	var parts []string
	for k, v := range fields {
		parts = append(parts, fmt.Sprintf("%s=%v", k, v))
	}
	return fmt.Sprintf("{%s}", strings.Join(parts, ", "))
}

// FilterFromReader reads log entries from an io.Reader and applies the filter
func FilterFromReader(r io.Reader, filter LogFilter) ([]LogEntry, error) {
	var entries []LogEntry

	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := scanner.Text()
		var entry LogEntry
		if err := json.Unmarshal([]byte(line), &entry); err != nil {
			continue
		}

		if matchesFilter(entry, filter) {
			entries = append(entries, entry)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading logs: %w", err)
	}

	return entries, nil
}

// GetTraceLogs returns all log entries for a specific trace ID
func GetTraceLogs(filePath, traceID string) ([]LogEntry, error) {
	filter := LogFilter{
		TraceID: traceID,
	}
	return FilterLogs(filePath, filter)
}

// GetComponentLogs returns all log entries for a specific component
func GetComponentLogs(filePath, component string) ([]LogEntry, error) {
	filter := LogFilter{
		Component: component,
	}
	return FilterLogs(filePath, filter)
}

// GetSessionLogs returns all log entries for a specific session
func GetSessionLogs(filePath, session string) ([]LogEntry, error) {
	filter := LogFilter{
		Session: session,
	}
	return FilterLogs(filePath, filter)
}

// GetRecentLogs returns log entries from a specified duration ago
func GetRecentLogs(filePath string, since time.Duration) ([]LogEntry, error) {
	filter := LogFilter{
		Since: since.String(),
	}
	return FilterLogs(filePath, filter)
}
