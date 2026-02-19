// Package debug provides per-component debug mode management for observability.
package debug

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLevel_String(t *testing.T) {
	tests := []struct {
		name     string
		level    Level
		expected string
	}{
		{
			name:     "TraceLevel string",
			level:    TraceLevel,
			expected: "trace",
		},
		{
			name:     "DebugLevel string",
			level:    DebugLevel,
			expected: "debug",
		},
		{
			name:     "InfoLevel string",
			level:    InfoLevel,
			expected: "info",
		},
		{
			name:     "WarnLevel string",
			level:    WarnLevel,
			expected: "warn",
		},
		{
			name:     "ErrorLevel string",
			level:    ErrorLevel,
			expected: "error",
		},
		{
			name:     "invalid level defaults to info",
			level:    Level(99),
			expected: "info",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.level.String()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestParseLevel(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected Level
	}{
		{
			name:     "parse trace",
			input:    "trace",
			expected: TraceLevel,
		},
		{
			name:     "parse debug",
			input:    "debug",
			expected: DebugLevel,
		},
		{
			name:     "parse info",
			input:    "info",
			expected: InfoLevel,
		},
		{
			name:     "parse warn",
			input:    "warn",
			expected: WarnLevel,
		},
		{
			name:     "parse warning (alias for warn)",
			input:    "warning",
			expected: WarnLevel,
		},
		{
			name:     "parse error",
			input:    "error",
			expected: ErrorLevel,
		},
		{
			name:     "parse invalid defaults to info",
			input:    "invalid",
			expected: InfoLevel,
		},
		{
			name:     "parse empty string defaults to info",
			input:    "",
			expected: InfoLevel,
		},
		{
			name:     "case sensitive - uppercase not parsed",
			input:    "DEBUG",
			expected: InfoLevel, // Falls back to default
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseLevel(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestNewManager(t *testing.T) {
	t.Run("creates new manager with defaults", func(t *testing.T) {
		mgr := NewManager()
		assert.NotNil(t, mgr)

		assert.Equal(t, InfoLevel, mgr.GetDefaultLevel())
		assert.Empty(t, mgr.ListComponents())
	})
}

func TestManager_SetComponentLevel(t *testing.T) {
	t.Run("sets component level", func(t *testing.T) {
		mgr := NewManager()

		mgr.SetComponentLevel("agent", DebugLevel)

		level := mgr.GetComponentLevel("agent")
		assert.Equal(t, DebugLevel, level)
	})

	t.Run("replaces existing component level", func(t *testing.T) {
		mgr := NewManager()

		mgr.SetComponentLevel("agent", DebugLevel)
		mgr.SetComponentLevel("agent", TraceLevel)

		level := mgr.GetComponentLevel("agent")
		assert.Equal(t, TraceLevel, level)
	})
}

func TestManager_GetComponentLevel(t *testing.T) {
	t.Run("returns default level for unset component", func(t *testing.T) {
		mgr := NewManager()
		mgr.SetDefaultLevel(DebugLevel)

		level := mgr.GetComponentLevel("nonexistent")
		assert.Equal(t, DebugLevel, level)
	})

	t.Run("returns component level when set", func(t *testing.T) {
		mgr := NewManager()

		mgr.SetComponentLevel("mcp", TraceLevel)

		level := mgr.GetComponentLevel("mcp")
		assert.Equal(t, TraceLevel, level)
	})

	t.Run("different components have independent levels", func(t *testing.T) {
		mgr := NewManager()

		mgr.SetComponentLevel("agent", DebugLevel)
		mgr.SetComponentLevel("mcp", TraceLevel)

		assert.Equal(t, DebugLevel, mgr.GetComponentLevel("agent"))
		assert.Equal(t, TraceLevel, mgr.GetComponentLevel("mcp"))
	})
}

func TestManager_SetDefaultLevel(t *testing.T) {
	t.Run("sets default level", func(t *testing.T) {
		mgr := NewManager()

		mgr.SetDefaultLevel(DebugLevel)

		assert.Equal(t, DebugLevel, mgr.GetDefaultLevel())
	})

	t.Run("affects unset components", func(t *testing.T) {
		mgr := NewManager()

		mgr.SetDefaultLevel(TraceLevel)
		mgr.SetComponentLevel("agent", DebugLevel)

		assert.Equal(t, DebugLevel, mgr.GetComponentLevel("agent"))
		assert.Equal(t, TraceLevel, mgr.GetComponentLevel("unset"))
	})

	t.Run("does not affect set components", func(t *testing.T) {
		mgr := NewManager()

		mgr.SetComponentLevel("agent", DebugLevel)
		mgr.SetDefaultLevel(TraceLevel)

		assert.Equal(t, DebugLevel, mgr.GetComponentLevel("agent"))
	})
}

func TestManager_GetDefaultLevel(t *testing.T) {
	t.Run("returns initial default level", func(t *testing.T) {
		mgr := NewManager()
		assert.Equal(t, InfoLevel, mgr.GetDefaultLevel())
	})

	t.Run("returns updated default level", func(t *testing.T) {
		mgr := NewManager()
		mgr.SetDefaultLevel(DebugLevel)

		assert.Equal(t, DebugLevel, mgr.GetDefaultLevel())
	})
}

func TestManager_ClearComponentLevel(t *testing.T) {
	t.Run("clears component level", func(t *testing.T) {
		mgr := NewManager()

		mgr.SetComponentLevel("agent", DebugLevel)
		assert.Equal(t, DebugLevel, mgr.GetComponentLevel("agent"))

		mgr.ClearComponentLevel("agent")
		assert.Equal(t, mgr.GetDefaultLevel(), mgr.GetComponentLevel("agent"))
	})

	t.Run("clearing nonexistent component is safe", func(t *testing.T) {
		mgr := NewManager()

		// Should not panic
		mgr.ClearComponentLevel("nonexistent")
	})

	t.Run("cleared component uses default level", func(t *testing.T) {
		mgr := NewManager()
		mgr.SetDefaultLevel(DebugLevel)

		mgr.SetComponentLevel("agent", TraceLevel)
		mgr.ClearComponentLevel("agent")

		assert.Equal(t, DebugLevel, mgr.GetComponentLevel("agent"))
	})
}

func TestManager_ListComponents(t *testing.T) {
	t.Run("returns empty map initially", func(t *testing.T) {
		mgr := NewManager()
		assert.Empty(t, mgr.ListComponents())
	})

	t.Run("lists components with custom levels", func(t *testing.T) {
		mgr := NewManager()

		mgr.SetComponentLevel("agent", DebugLevel)
		mgr.SetComponentLevel("mcp", TraceLevel)

		components := mgr.ListComponents()
		assert.Len(t, components, 2)
		assert.Equal(t, DebugLevel, components["agent"])
		assert.Equal(t, TraceLevel, components["mcp"])
	})

	t.Run("does not include cleared components", func(t *testing.T) {
		mgr := NewManager()

		mgr.SetComponentLevel("agent", DebugLevel)
		mgr.SetComponentLevel("mcp", TraceLevel)
		mgr.ClearComponentLevel("agent")

		components := mgr.ListComponents()
		assert.Len(t, components, 1)
		assert.Contains(t, components, "mcp")
		assert.NotContains(t, components, "agent")
	})

	t.Run("returns a copy (modifications don't affect manager)", func(t *testing.T) {
		mgr := NewManager()

		mgr.SetComponentLevel("agent", DebugLevel)

		components := mgr.ListComponents()
		components["new"] = TraceLevel

		// Original manager should not be affected
		assert.NotContains(t, mgr.ListComponents(), "new")
	})
}

func TestManager_IsEnabled(t *testing.T) {
	tests := []struct {
		name       string
		setup      func(*Manager)
		component  string
		level      Level
		wantEnabled bool
	}{
		{
			name: "component at debug level enables debug",
			setup: func(m *Manager) {
				m.SetComponentLevel("agent", DebugLevel)
			},
			component:  "agent",
			level:      DebugLevel,
			wantEnabled: true,
		},
		{
			name: "component at debug level enables info",
			setup: func(m *Manager) {
				m.SetComponentLevel("agent", DebugLevel)
			},
			component:  "agent",
			level:      InfoLevel,
			wantEnabled: true,
		},
		{
			name: "component at info level disables debug",
			setup: func(m *Manager) {
				m.SetComponentLevel("agent", InfoLevel)
			},
			component:  "agent",
			level:      DebugLevel,
			wantEnabled: false,
		},
		{
			name: "component at trace level enables all",
			setup: func(m *Manager) {
				m.SetComponentLevel("agent", TraceLevel)
			},
			component:  "agent",
			level:      DebugLevel,
			wantEnabled: true,
		},
		{
			name: "unset component uses default level",
			setup: func(m *Manager) {
				m.SetDefaultLevel(DebugLevel)
			},
			component:  "agent",
			level:      DebugLevel,
			wantEnabled: true,
		},
		{
			name: "error level is always enabled when level <= error",
			setup: func(m *Manager) {
				m.SetComponentLevel("agent", ErrorLevel)
			},
			component:  "agent",
			level:      ErrorLevel,
			wantEnabled: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mgr := NewManager()
			tt.setup(mgr)

			result := mgr.IsEnabled(tt.component, tt.level)
			assert.Equal(t, tt.wantEnabled, result)
		})
	}
}

func TestSetGlobalComponentLevel(t *testing.T) {
	t.Run("sets global component level", func(t *testing.T) {
		// Reset global manager state
		SetGlobalDefaultLevel(InfoLevel)

		SetGlobalComponentLevel("agent", DebugLevel)

		level := GetGlobalComponentLevel("agent")
		assert.Equal(t, DebugLevel, level)
	})

	t.Run("affects global manager", func(t *testing.T) {
		SetGlobalDefaultLevel(InfoLevel)
		SetGlobalComponentLevel("mcp", TraceLevel)

		assert.True(t, IsGloballyEnabled("mcp", TraceLevel))
		assert.False(t, IsGloballyEnabled("other", TraceLevel))
	})
}

func TestGetGlobalComponentLevel(t *testing.T) {
	t.Run("returns default level for unset component", func(t *testing.T) {
		SetGlobalDefaultLevel(DebugLevel)

		level := GetGlobalComponentLevel("nonexistent")
		assert.Equal(t, DebugLevel, level)
	})
}

func TestSetGlobalDefaultLevel(t *testing.T) {
	t.Run("sets global default level", func(t *testing.T) {
		SetGlobalDefaultLevel(TraceLevel)

		assert.Equal(t, TraceLevel, GetGlobalComponentLevel("any"))
	})
}

func TestIsGloballyEnabled(t *testing.T) {
	t.Run("checks global level for component", func(t *testing.T) {
		SetGlobalDefaultLevel(InfoLevel)
		SetGlobalComponentLevel("agent", DebugLevel)

		assert.True(t, IsGloballyEnabled("agent", DebugLevel))
		assert.False(t, IsGloballyEnabled("agent", TraceLevel))
		assert.True(t, IsGloballyEnabled("other", InfoLevel))
		assert.False(t, IsGloballyEnabled("other", DebugLevel))
	})
}

func TestShouldTrace(t *testing.T) {
	t.Run("returns true when trace level enabled", func(t *testing.T) {
		SetGlobalDefaultLevel(InfoLevel)
		SetGlobalComponentLevel("agent", TraceLevel)

		assert.True(t, ShouldTrace("agent"))
		assert.False(t, ShouldTrace("other"))
	})
}

func TestShouldDebug(t *testing.T) {
	t.Run("returns true when debug level enabled", func(t *testing.T) {
		SetGlobalDefaultLevel(InfoLevel)
		SetGlobalComponentLevel("agent", DebugLevel)

		assert.True(t, ShouldDebug("agent"))
		assert.False(t, ShouldDebug("other"))
	})

	t.Run("trace level also enables debug", func(t *testing.T) {
		SetGlobalDefaultLevel(InfoLevel)
		SetGlobalComponentLevel("agent", TraceLevel)

		assert.True(t, ShouldDebug("agent"))
	})
}

func TestSetDebugMode(t *testing.T) {
	t.Run("sets multiple components to same level", func(t *testing.T) {
		SetGlobalDefaultLevel(InfoLevel)

		components := []string{"agent", "mcp", "tools"}
		SetDebugMode(components, DebugLevel)

		for _, comp := range components {
			assert.True(t, IsGloballyEnabled(comp, DebugLevel))
		}
		assert.False(t, IsGloballyEnabled("other", DebugLevel))
	})
}

func TestEnableTraceFor(t *testing.T) {
	t.Run("enables trace for specified components", func(t *testing.T) {
		SetGlobalDefaultLevel(InfoLevel)

		EnableTraceFor("agent", "mcp")

		assert.True(t, ShouldTrace("agent"))
		assert.True(t, ShouldTrace("mcp"))
		assert.False(t, ShouldTrace("other"))
	})
}

func TestEnableDebugFor(t *testing.T) {
	t.Run("enables debug for specified components", func(t *testing.T) {
		SetGlobalDefaultLevel(InfoLevel)

		EnableDebugFor("agent", "mcp")

		assert.True(t, ShouldDebug("agent"))
		assert.True(t, ShouldDebug("mcp"))
		assert.False(t, ShouldDebug("other"))
	})
}

func TestDisableDebugFor(t *testing.T) {
	t.Run("disables debug for specified components", func(t *testing.T) {
		SetGlobalDefaultLevel(InfoLevel)
		EnableDebugFor("agent", "mcp", "tools")

		assert.True(t, ShouldDebug("agent"))

		DisableDebugFor("agent", "mcp")

		assert.False(t, ShouldDebug("agent"))
		assert.False(t, ShouldDebug("mcp"))
		assert.True(t, ShouldDebug("tools"))
	})
}

func TestGetDebugConfig(t *testing.T) {
	t.Run("returns config with default level", func(t *testing.T) {
		SetGlobalDefaultLevel(DebugLevel)

		config := GetDebugConfig()

		assert.Equal(t, "debug", config["default"])
		assert.NotNil(t, config["components"])
	})

	t.Run("returns config with component levels", func(t *testing.T) {
		SetGlobalDefaultLevel(InfoLevel)
		SetGlobalComponentLevel("agent", DebugLevel)
		SetGlobalComponentLevel("mcp", TraceLevel)

		config := GetDebugConfig()

		assert.Equal(t, "info", config["default"])

		components, ok := config["components"].(map[string]string)
		require.True(t, ok)
		assert.Equal(t, "debug", components["agent"])
		assert.Equal(t, "trace", components["mcp"])
	})
}

func TestParseDebugConfig(t *testing.T) {
	t.Run("parses default level", func(t *testing.T) {
		SetGlobalDefaultLevel(InfoLevel)

		config := map[string]interface{}{
			"default": "debug",
		}

		err := ParseDebugConfig(config)
		assert.NoError(t, err)

		assert.Equal(t, DebugLevel, GetGlobalComponentLevel("any"))
	})

	t.Run("parses component levels", func(t *testing.T) {
		SetGlobalDefaultLevel(InfoLevel)

		config := map[string]interface{}{
			"components": map[string]interface{}{
				"agent": "debug",
				"mcp":   "trace",
			},
		}

		err := ParseDebugConfig(config)
		assert.NoError(t, err)

		assert.Equal(t, DebugLevel, GetGlobalComponentLevel("agent"))
		assert.Equal(t, TraceLevel, GetGlobalComponentLevel("mcp"))
	})

	t.Run("parses both default and component levels", func(t *testing.T) {
		SetGlobalDefaultLevel(InfoLevel)

		config := map[string]interface{}{
			"default": "warn",
			"components": map[string]interface{}{
				"agent": "debug",
			},
		}

		err := ParseDebugConfig(config)
		assert.NoError(t, err)

		assert.Equal(t, WarnLevel, GetGlobalComponentLevel("other"))
		assert.Equal(t, DebugLevel, GetGlobalComponentLevel("agent"))
	})

	t.Run("handles missing components field", func(t *testing.T) {
		SetGlobalDefaultLevel(InfoLevel)

		config := map[string]interface{}{
			"default": "debug",
		}

		err := ParseDebugConfig(config)
		assert.NoError(t, err)
	})

	t.Run("handles invalid component level values", func(t *testing.T) {
		SetGlobalDefaultLevel(InfoLevel)

		config := map[string]interface{}{
			"components": map[string]interface{}{
				"test_agent": "debug",
				"test_mcp":   123, // Invalid type - should be ignored
			},
		}

		err := ParseDebugConfig(config)
		assert.NoError(t, err)

		// test_agent should be set, test_mcp should remain at default
		assert.Equal(t, DebugLevel, GetGlobalComponentLevel("test_agent"))
		assert.Equal(t, InfoLevel, GetGlobalComponentLevel("test_mcp"))
	})
}

func TestManager_ConcurrentAccess(t *testing.T) {
	t.Run("concurrent level setting is safe", func(t *testing.T) {
		mgr := NewManager()
		var wg sync.WaitGroup

		// Concurrent writes
		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func(idx int) {
				defer wg.Done()
				mgr.SetComponentLevel("comp", Level(idx%5))
			}(i)
		}

		// Concurrent reads
		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				_ = mgr.GetComponentLevel("comp")
			}()
		}

		wg.Wait()
		// Just verify no race/detector issues
	})

	t.Run("concurrent list and modify is safe", func(t *testing.T) {
		mgr := NewManager()
		var wg sync.WaitGroup

		for i := 0; i < 5; i++ {
			wg.Add(1)
			go func(idx int) {
				defer wg.Done()
				mgr.SetComponentLevel("comp", Level(idx))
			}(i)
		}

		for i := 0; i < 5; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				_ = mgr.ListComponents()
			}()
		}

		wg.Wait()
	})
}

func TestLevelOrdering(t *testing.T) {
	t.Run("levels are in correct order", func(t *testing.T) {
		// TraceLevel < DebugLevel < InfoLevel < WarnLevel < ErrorLevel
		assert.True(t, TraceLevel < DebugLevel)
		assert.True(t, DebugLevel < InfoLevel)
		assert.True(t, InfoLevel < WarnLevel)
		assert.True(t, WarnLevel < ErrorLevel)
	})
}

func TestRoundTripConfig(t *testing.T) {
	t.Run("config round-trip preserves settings", func(t *testing.T) {
		// Setup initial state
		SetGlobalDefaultLevel(WarnLevel)
		SetGlobalComponentLevel("agent", DebugLevel)
		SetGlobalComponentLevel("mcp", TraceLevel)

		// Get config
		config := GetDebugConfig()

		// Reset
		SetGlobalDefaultLevel(InfoLevel)

		// Parse config back
		err := ParseDebugConfig(config)
		assert.NoError(t, err)

		// Verify
		assert.Equal(t, WarnLevel, GetGlobalComponentLevel("other"))
		assert.Equal(t, DebugLevel, GetGlobalComponentLevel("agent"))
		assert.Equal(t, TraceLevel, GetGlobalComponentLevel("mcp"))
	})
}
