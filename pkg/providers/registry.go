// PicoClaw - Ultra-lightweight personal AI agent
// Inspired by and based on nanobot: https://github.com/HKUDS/nanobot
// License: MIT
//
// Copyright (c) 2026 PicoClaw contributors

package providers

import (
	"fmt"
	"strings"

	"github.com/sipeed/picoclaw/pkg/config"
)

// ResolvedLLMConfig represents a fully resolved LLM configuration from the registry.
// It merges provider settings (API key, endpoint) with model-specific settings.
type ResolvedLLMConfig struct {
	Provider    string
	APIKey      string
	APIBase     string
	Proxy       string
	Model       string
	AuthMethod  string
	Workspace   string
	ConnectMode string
	WebSearch   bool
}

// providerResolver is a function that creates a provider from resolved config.
type providerResolver func(cfg *ResolvedLLMConfig) (LLMProvider, error)

// registry maps provider names to their resolver functions.
var registry = map[string]providerResolver{
	"anthropic":     resolveAnthropicProvider,
	"openai":        resolveOpenAIProvider,
	"zai":           resolveZAIProvider,
	"zhipu":         resolveHTTPCompatProvider,
	"groq":          resolveHTTPCompatProvider,
	"gemini":        resolveHTTPCompatProvider,
	"openrouter":    resolveHTTPCompatProvider,
	"vllm":          resolveHTTPCompatProvider,
	"deepseek":      resolveHTTPCompatProvider,
	"ollama":        resolveHTTPCompatProvider,
	"nvidia":        resolveHTTPCompatProvider,
	"moonshot":      resolveHTTPCompatProvider,
	"shengsuanyun":  resolveHTTPCompatProvider,
	"claude-cli":    resolveClaudeCLIProvider,
	"claudecode":    resolveClaudeCLIProvider,
	"codex-cli":     resolveCodexCLIProvider,
	"codex-code":    resolveCodexCLIProvider,
	"github_copilot": resolveGitHubCopilotProvider,
	"copilot":       resolveGitHubCopilotProvider,
}

// ResolveLLM resolves a model reference to a full LLM configuration.
// It uses the provider name from the model reference or defaults to look up
// the provider based on available API keys and model patterns.
func ResolveLLM(cfg *config.Config, modelRef string) (*ResolvedLLMConfig, error) {
	if modelRef == "" {
		modelRef = cfg.Agents.Defaults.Model
	}
	if modelRef == "" {
		return nil, fmt.Errorf("no model specified and no default model in config")
	}

	ref := ParseModelRef(modelRef, cfg.Agents.Defaults.Provider)
	if ref == nil {
		return nil, fmt.Errorf("invalid model reference: %s", modelRef)
	}

	// Get provider name - either from ref or from config default
	providerName := ref.Provider
	if providerName == "" {
		providerName = strings.ToLower(cfg.Agents.Defaults.Provider)
	}

	// Normalize provider name
	providerName = NormalizeProvider(providerName)

	// Resolve the config based on provider
	resolved, err := resolveProviderConfig(cfg, providerName, ref.Model)
	if err != nil {
		return nil, fmt.Errorf("resolving provider %s for model %s: %w", providerName, ref.Model, err)
	}

	// Override model if specified in ref
	if ref.Model != "" {
		resolved.Model = ref.Model
	}

	return resolved, nil
}

// resolveProviderConfig resolves the configuration for a specific provider.
func resolveProviderConfig(cfg *config.Config, providerName string, model string) (*ResolvedLLMConfig, error) {
	resolved := &ResolvedLLMConfig{
		Provider: providerName,
		Model:    model,
	}

	switch providerName {
	case "anthropic", "claude":
		resolved.APIKey = cfg.Providers.Anthropic.APIKey
		resolved.APIBase = cfg.Providers.Anthropic.APIBase
		resolved.Proxy = cfg.Providers.Anthropic.Proxy
		resolved.AuthMethod = cfg.Providers.Anthropic.AuthMethod
		if resolved.APIBase == "" {
			resolved.APIBase = "https://api.anthropic.com/v1"
		}

	case "openai", "gpt":
		resolved.APIKey = cfg.Providers.OpenAI.APIKey
		resolved.APIBase = cfg.Providers.OpenAI.APIBase
		resolved.Proxy = cfg.Providers.OpenAI.Proxy
		resolved.AuthMethod = cfg.Providers.OpenAI.AuthMethod
		resolved.WebSearch = cfg.Providers.OpenAI.WebSearch
		if resolved.APIBase == "" {
			resolved.APIBase = "https://api.openai.com/v1"
		}

	case "zai":
		resolved.APIKey = cfg.Providers.ZAI.APIKey
		resolved.APIBase = cfg.Providers.ZAI.APIBase
		resolved.Proxy = cfg.Providers.ZAI.Proxy
		if resolved.APIBase == "" {
			resolved.APIBase = "https://api.z.ai/api/coding/paas/v4/chat/completions"
		}

	case "zhipu", "glm":
		resolved.APIKey = cfg.Providers.Zhipu.APIKey
		resolved.APIBase = cfg.Providers.Zhipu.APIBase
		resolved.Proxy = cfg.Providers.Zhipu.Proxy
		if resolved.APIBase == "" {
			resolved.APIBase = "https://open.bigmodel.cn/api/paas/v4"
		}

	case "groq":
		resolved.APIKey = cfg.Providers.Groq.APIKey
		resolved.APIBase = cfg.Providers.Groq.APIBase
		resolved.Proxy = cfg.Providers.Groq.Proxy
		if resolved.APIBase == "" {
			resolved.APIBase = "https://api.groq.com/openai/v1"
		}

	case "gemini", "google":
		resolved.APIKey = cfg.Providers.Gemini.APIKey
		resolved.APIBase = cfg.Providers.Gemini.APIBase
		resolved.Proxy = cfg.Providers.Gemini.Proxy
		if resolved.APIBase == "" {
			resolved.APIBase = "https://generativelanguage.googleapis.com/v1beta"
		}

	case "openrouter":
		resolved.APIKey = cfg.Providers.OpenRouter.APIKey
		resolved.APIBase = cfg.Providers.OpenRouter.APIBase
		resolved.Proxy = cfg.Providers.OpenRouter.Proxy
		if resolved.APIBase == "" {
			resolved.APIBase = "https://openrouter.ai/api/v1"
		}

	case "vllm":
		resolved.APIKey = cfg.Providers.VLLM.APIKey
		resolved.APIBase = cfg.Providers.VLLM.APIBase
		resolved.Proxy = cfg.Providers.VLLM.Proxy

	case "deepseek":
		resolved.APIKey = cfg.Providers.DeepSeek.APIKey
		resolved.APIBase = cfg.Providers.DeepSeek.APIBase
		resolved.Proxy = cfg.Providers.DeepSeek.Proxy
		if resolved.APIBase == "" {
			resolved.APIBase = "https://api.deepseek.com/v1"
		}
		if model != "deepseek-chat" && model != "deepseek-reasoner" {
			resolved.Model = "deepseek-chat"
		}

	case "ollama":
		resolved.APIKey = cfg.Providers.Ollama.APIKey
		resolved.APIBase = cfg.Providers.Ollama.APIBase
		resolved.Proxy = cfg.Providers.Ollama.Proxy
		if resolved.APIBase == "" {
			resolved.APIBase = "http://localhost:11434/v1"
		}

	case "nvidia":
		resolved.APIKey = cfg.Providers.Nvidia.APIKey
		resolved.APIBase = cfg.Providers.Nvidia.APIBase
		resolved.Proxy = cfg.Providers.Nvidia.Proxy
		if resolved.APIBase == "" {
			resolved.APIBase = "https://integrate.api.nvidia.com/v1"
		}

	case "moonshot", "kimi":
		resolved.APIKey = cfg.Providers.Moonshot.APIKey
		resolved.APIBase = cfg.Providers.Moonshot.APIBase
		resolved.Proxy = cfg.Providers.Moonshot.Proxy
		if resolved.APIBase == "" {
			resolved.APIBase = "https://api.moonshot.cn/v1"
		}

	case "shengsuanyun":
		resolved.APIKey = cfg.Providers.ShengSuanYun.APIKey
		resolved.APIBase = cfg.Providers.ShengSuanYun.APIBase
		resolved.Proxy = cfg.Providers.ShengSuanYun.Proxy
		if resolved.APIBase == "" {
			resolved.APIBase = "https://router.shengsuanyun.com/api/v1"
		}

	case "claude-cli", "claude-code", "claudecode":
		resolved.Workspace = cfg.WorkspacePath()
		if resolved.Workspace == "" {
			resolved.Workspace = "."
		}

	case "codex-cli", "codex-code":
		resolved.Workspace = cfg.WorkspacePath()
		if resolved.Workspace == "" {
			resolved.Workspace = "."
		}

	case "github_copilot", "copilot":
		resolved.APIBase = cfg.Providers.GitHubCopilot.APIBase
		resolved.ConnectMode = cfg.Providers.GitHubCopilot.ConnectMode
		if resolved.APIBase == "" {
			resolved.APIBase = "localhost:4321"
		}

	default:
		return nil, fmt.Errorf("unsupported provider: %s", providerName)
	}

	return resolved, nil
}

// CreateProviderFromRegistry creates an LLM provider using the registry resolution pattern.
// This is the main entry point for dynamic provider creation based on model references.
func CreateProviderFromRegistry(cfg *config.Config, modelRef string) (LLMProvider, error) {
	resolved, err := ResolveLLM(cfg, modelRef)
	if err != nil {
		return nil, err
	}

	resolver, exists := registry[resolved.Provider]
	if !exists {
		return nil, fmt.Errorf("no resolver registered for provider: %s", resolved.Provider)
	}

	return resolver(resolved)
}

// Provider resolver functions

func resolveAnthropicProvider(cfg *ResolvedLLMConfig) (LLMProvider, error) {
	if cfg.AuthMethod == "oauth" || cfg.AuthMethod == "token" {
		cred, err := getCredential("anthropic")
		if err != nil {
			return nil, fmt.Errorf("loading auth credentials: %w", err)
		}
		if cred == nil {
			return nil, fmt.Errorf("no credentials for anthropic. Run: picoclaw auth login --provider anthropic")
		}
		return NewClaudeProviderWithTokenSourceAndBaseURL(cred.AccessToken, createClaudeTokenSource(), cfg.APIBase), nil
	}

	if cfg.APIKey == "" {
		return nil, fmt.Errorf("no API key for anthropic provider")
	}
	return NewClaudeProviderWithBaseURL(cfg.APIKey, cfg.APIBase), nil
}

func resolveOpenAIProvider(cfg *ResolvedLLMConfig) (LLMProvider, error) {
	if cfg.AuthMethod == "codex-cli" {
		p := NewCodexProviderWithTokenSource("", "", CreateCodexCliTokenSource())
		p.enableWebSearch = cfg.WebSearch
		return p, nil
	}

	if cfg.AuthMethod == "oauth" || cfg.AuthMethod == "token" {
		cred, err := getCredential("openai")
		if err != nil {
			return nil, fmt.Errorf("loading auth credentials: %w", err)
		}
		if cred == nil {
			return nil, fmt.Errorf("no credentials for openai. Run: picoclaw auth login --provider openai")
		}
		p := NewCodexProviderWithTokenSource(cred.AccessToken, cred.AccountID, createCodexTokenSource())
		p.enableWebSearch = cfg.WebSearch
		return p, nil
	}

	if cfg.APIKey == "" {
		return nil, fmt.Errorf("no API key for openai provider")
	}
	return NewHTTPProvider(cfg.APIKey, cfg.APIBase, cfg.Proxy), nil
}

func resolveZAIProvider(cfg *ResolvedLLMConfig) (LLMProvider, error) {
	if cfg.APIKey == "" {
		return nil, fmt.Errorf("no API key for zai provider")
	}
	return NewZAIProvider(cfg.APIKey, cfg.APIBase, cfg.Proxy), nil
}

func resolveHTTPCompatProvider(cfg *ResolvedLLMConfig) (LLMProvider, error) {
	if cfg.APIKey == "" && !strings.HasPrefix(cfg.Model, "bedrock/") {
		return nil, fmt.Errorf("no API key for provider %s", cfg.Provider)
	}
	if cfg.APIBase == "" {
		return nil, fmt.Errorf("no API base configured for provider %s", cfg.Provider)
	}
	return NewHTTPProvider(cfg.APIKey, cfg.APIBase, cfg.Proxy), nil
}

func resolveClaudeCLIProvider(cfg *ResolvedLLMConfig) (LLMProvider, error) {
	return NewClaudeCliProvider(cfg.Workspace), nil
}

func resolveCodexCLIProvider(cfg *ResolvedLLMConfig) (LLMProvider, error) {
	return NewCodexCliProvider(cfg.Workspace), nil
}

func resolveGitHubCopilotProvider(cfg *ResolvedLLMConfig) (LLMProvider, error) {
	return NewGitHubCopilotProvider(cfg.APIBase, cfg.ConnectMode, cfg.Model)
}
