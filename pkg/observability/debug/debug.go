// Package debug provides per-component debug mode management for observability.
package debug

import (
	"sync"
)

// Level represents the logging verbosity level
type Level int

const (
	// TraceLevel is the most verbose, tracing every detail
	TraceLevel Level = iota
	// DebugLevel is for detailed debugging information
	DebugLevel
	// InfoLevel is for general informational messages
	InfoLevel
	// WarnLevel is for warning messages
	WarnLevel
	// ErrorLevel is for error messages
	ErrorLevel
)

// String returns the string representation of the level
func (l Level) String() string {
	switch l {
	case TraceLevel:
		return "trace"
	case DebugLevel:
		return "debug"
	case InfoLevel:
		return "info"
	case WarnLevel:
		return "warn"
	case ErrorLevel:
		return "error"
	default:
		return "info"
	}
}

// ParseLevel parses a string into a Level
func ParseLevel(s string) Level {
	switch s {
	case "trace":
		return TraceLevel
	case "debug":
		return DebugLevel
	case "info":
		return InfoLevel
	case "warn", "warning":
		return WarnLevel
	case "error":
		return ErrorLevel
	default:
		return InfoLevel
	}
}

// Manager manages debug levels for different components
type Manager struct {
	mu         sync.RWMutex
	components map[string]Level
	defaultLevel Level
}

// globalManager is the singleton debug manager
var (
	globalManager = &Manager{
		components:  make(map[string]Level),
		defaultLevel: InfoLevel,
	}
	managerMutex sync.RWMutex
)

// NewManager creates a new debug manager
func NewManager() *Manager {
	return &Manager{
		components:  make(map[string]Level),
		defaultLevel: InfoLevel,
	}
}

// SetComponentLevel sets the debug level for a specific component
func (m *Manager) SetComponentLevel(component string, level Level) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.components[component] = level
}

// GetComponentLevel gets the debug level for a component
// If no specific level is set, returns the default level
func (m *Manager) GetComponentLevel(component string) Level {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if level, ok := m.components[component]; ok {
		return level
	}
	return m.defaultLevel
}

// SetDefaultLevel sets the default debug level for all components
func (m *Manager) SetDefaultLevel(level Level) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.defaultLevel = level
}

// GetDefaultLevel returns the default debug level
func (m *Manager) GetDefaultLevel() Level {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.defaultLevel
}

// ClearComponentLevel clears the debug level for a component (reverts to default)
func (m *Manager) ClearComponentLevel(component string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.components, component)
}

// ListComponents returns a list of all components with custom levels
func (m *Manager) ListComponents() map[string]Level {
	m.mu.RLock()
	defer m.mu.RUnlock()
	result := make(map[string]Level, len(m.components))
	for k, v := range m.components {
		result[k] = v
	}
	return result
}

// IsEnabled checks if a given level is enabled for a component
func (m *Manager) IsEnabled(component string, level Level) bool {
	return m.GetComponentLevel(component) <= level
}

// Global functions for convenience

// SetGlobalComponentLevel sets the debug level for a component in the global manager
func SetGlobalComponentLevel(component string, level Level) {
	globalManager.SetComponentLevel(component, level)
}

// GetGlobalComponentLevel gets the debug level for a component from the global manager
func GetGlobalComponentLevel(component string) Level {
	return globalManager.GetComponentLevel(component)
}

// SetGlobalDefaultLevel sets the default debug level in the global manager
func SetGlobalDefaultLevel(level Level) {
	globalManager.SetDefaultLevel(level)
}

// IsGloballyEnabled checks if a level is enabled for a component globally
func IsGloballyEnabled(component string, level Level) bool {
	return globalManager.IsEnabled(component, level)
}

// ShouldTrace returns true if trace level is enabled for the component
func ShouldTrace(component string) bool {
	return globalManager.IsEnabled(component, TraceLevel)
}

// ShouldDebug returns true if debug level is enabled for the component
func ShouldDebug(component string) bool {
	return globalManager.IsEnabled(component, DebugLevel)
}

// SetDebugMode sets multiple components to debug level at once
func SetDebugMode(components []string, level Level) {
	for _, component := range components {
		globalManager.SetComponentLevel(component, level)
	}
}

// EnableTraceFor enables trace level for specified components
func EnableTraceFor(components ...string) {
	SetDebugMode(components, TraceLevel)
}

// EnableDebugFor enables debug level for specified components
func EnableDebugFor(components ...string) {
	SetDebugMode(components, DebugLevel)
}

// DisableDebugFor disables debug (sets to info level) for specified components
func DisableDebugFor(components ...string) {
	for _, component := range components {
		globalManager.ClearComponentLevel(component)
	}
}

// GetDebugConfig returns a map representation of the current debug configuration
func GetDebugConfig() map[string]interface{} {
	config := make(map[string]interface{})
	config["default"] = globalManager.GetDefaultLevel().String()
	config["components"] = make(map[string]string)

	components := globalManager.ListComponents()
	componentMap := make(map[string]string)
	for component, level := range components {
		componentMap[component] = level.String()
	}
	config["components"] = componentMap

	return config
}

// ParseDebugConfig parses a map configuration and applies it
func ParseDebugConfig(config map[string]interface{}) error {
	// Parse default level
	if defaultLevel, ok := config["default"].(string); ok {
		globalManager.SetDefaultLevel(ParseLevel(defaultLevel))
	}

	// Parse component levels
	if components, ok := config["components"].(map[string]interface{}); ok {
		for component, level := range components {
			if levelStr, ok := level.(string); ok {
				globalManager.SetComponentLevel(component, ParseLevel(levelStr))
			}
		}
	}

	return nil
}
