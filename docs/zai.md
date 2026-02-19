# Z.ai Provider Configuration

Z.ai is a high-performance LLM provider that offers access to advanced GLM models with excellent performance for both coding and general AI tasks.

## Quick Start

### 1. Get Your API Key

1. Visit [Z.ai](https://z.ai) to sign up for an account
2. Navigate to the API keys section in your dashboard
3. Generate a new API key
4. Set the environment variable:

```bash
export ZAI_API_KEY=your-zai-api-key-here
```

### 2. Configure PicoClaw

Edit your `~/.picoclaw/config.json`:

```json
{
  "agents": {
    "defaults": {
      "model": "glm-4-plus",
      "max_tokens": 8192,
      "temperature": 0.7,
      "max_tool_iterations": 20
    }
  },
  "providers": {
    "zai": {
      "api_key": "YOUR_ZAI_API_KEY",
      "api_base": "https://api.z.ai/api/coding/paas/v4/chat/completions"
    }
  }
}
```

### 3. Start Using PicoClaw

```bash
picoclaw agent -m "Hello! Can you help me write a Go function?"
```

## Available Models

Z.ai provides access to several GLM (General Language Model) variants:

| Model | Use Case | Context Window | Best For |
|-------|----------|----------------|----------|
| **GLM-4-Plus** | General purpose | 128K+ | Most tasks, balanced performance |
| **GLM-4-Flash** | Fast responses | 128K+ | Quick responses, simple tasks |
| **GLM-4-Air** | Lightweight | 128K+ | Resource-constrained environments |
| **GLM-4V** | Vision capable | 128K+ | Image analysis, multimodal tasks |
| **GLM-4-Long** | Long context | 1M+ | Large document analysis |

### Model Selection Guide

- **Default Choice**: `glm-4-plus` - Best overall performance and capabilities
- **Speed**: `glm-4-flash` - Fastest response times, ideal for simple queries
- **Low Resource**: `glm-4-air` - Optimized for minimal memory/CPU usage
- **Vision**: `glm-4v` - When you need image understanding capabilities
- **Long Documents**: `glm-4-long` - For analyzing very large texts

## Configuration Examples

### Basic Configuration

```json
{
  "agents": {
    "defaults": {
      "model": "glm-4-plus",
      "max_tokens": 8192,
      "temperature": 0.7
    }
  },
  "providers": {
    "zai": {
      "api_key": "YOUR_ZAI_API_KEY",
      "api_base": "https://api.z.ai/api/coding/paas/v4/chat/completions"
    }
  }
}
```

### Fast Response Configuration

```json
{
  "agents": {
    "defaults": {
      "model": "glm-4-flash",
      "max_tokens": 4096,
      "temperature": 0.5
    }
  },
  "providers": {
    "zai": {
      "api_key": "YOUR_ZAI_API_KEY"
    }
  }
}
```

### Environment Variable Configuration

For better security, use environment variables instead of hardcoding API keys:

```bash
# Set environment variable
export ZAI_API_KEY=your-zai-api-key-here
```

```json
{
  "providers": {
    "zai": {
      "api_key": "${ZAI_API_KEY}",
      "api_base": "https://api.z.ai/api/coding/paas/v4/chat/completions"
    }
  }
}
```

### Custom Endpoint

If you're using a custom endpoint or proxy:

```json
{
  "providers": {
    "zai": {
      "api_key": "${ZAI_API_KEY}",
      "api_base": "https://your-custom-endpoint.com/api/coding/paas/v4/chat/completions"
    }
  }
}
```

## Advanced Configuration

### Temperature Settings

The temperature parameter controls the randomness of responses:

- `0.0 - 0.3`: More focused, deterministic outputs (good for code generation)
- `0.4 - 0.7`: Balanced creativity and focus (good for general tasks)
- `0.8 - 1.0`: More creative, diverse outputs (good for brainstorming)

```json
{
  "agents": {
    "defaults": {
      "temperature": 0.3
    }
  }
}
```

### Token Limits

Adjust `max_tokens` based on your needs:

```json
{
  "agents": {
    "defaults": {
      "max_tokens": 16384
    }
  }
}
```

### Multiple Provider Setup

You can configure Z.ai alongside other providers:

```json
{
  "providers": {
    "zai": {
      "api_key": "${ZAI_API_KEY}",
      "api_base": "https://api.z.ai/api/coding/paas/v4/chat/completions"
    },
    "zhipu": {
      "api_key": "YOUR_ZHIPU_API_KEY",
      "api_base": "https://open.bigmodel.cn/api/paas/v4"
    },
    "openrouter": {
      "api_key": "YOUR_OPENROUTER_API_KEY"
    }
  }
}
```

## Environment Variables

PicoClaw supports environment variable substitution in the config file:

```bash
# Required
ZAI_API_KEY=your-zai-api-key-here

# Optional: Override API base
ZAI_API_BASE=https://api.z.ai/api/coding/paas/v4/chat/completions
```

Usage in config:

```json
{
  "providers": {
    "zai": {
      "api_key": "${ZAI_API_KEY}",
      "api_base": "${ZAI_API_BASE:-https://api.z.ai/api/coding/paas/v4/chat/completions}"
    }
  }
}
```

## Use Cases

### 1. Code Development

```bash
picoclaw agent -m "Write a REST API in Go that handles user authentication"
```

Recommended model: `glm-4-plus`

### 2. Quick Questions

```bash
picoclaw agent -m "What's the difference between TCP and UDP?"
```

Recommended model: `glm-4-flash`

### 3. Long Document Analysis

```bash
picoclaw agent -m "Summarize the key points from this document: /path/to/large-document.txt"
```

Recommended model: `glm-4-long`

### 4. Image Analysis

```bash
picoclaw agent -m "Analyze this screenshot and describe the UI layout"
```

Recommended model: `glm-4v`

## Performance Optimization

### For Low-Resource Devices

If running on hardware with limited resources:

```json
{
  "agents": {
    "defaults": {
      "model": "glm-4-air",
      "max_tokens": 4096,
      "temperature": 0.5,
      "max_tool_iterations": 10
    }
  }
}
```

### For Maximum Quality

For best quality outputs on capable hardware:

```json
{
  "agents": {
    "defaults": {
      "model": "glm-4-plus",
      "max_tokens": 16384,
      "temperature": 0.7,
      "max_tool_iterations": 30
    }
  }
}
```

## Troubleshooting

### API Key Issues

**Error**: "Authentication failed" or "401 Unauthorized"

**Solution**:
1. Verify your API key is correct
2. Check that the environment variable is set: `echo $ZAI_API_KEY`
3. Ensure the API key has not expired

### Connection Issues

**Error**: "Connection refused" or "timeout"

**Solution**:
1. Check your internet connection
2. Verify the API base URL is correct
3. Check if a proxy is required

### Model Not Available

**Error**: "Model not found"

**Solution**:
1. Verify the model name is correct
2. Check if the model is available in your account tier
3. Try a different model (e.g., `glm-4-plus` instead of `glm-4-long`)

### Rate Limiting

**Error**: "Rate limit exceeded"

**Solution**:
1. Reduce request frequency
2. Upgrade your Z.ai plan for higher limits
3. Consider using `glm-4-flash` for faster, lower-cost requests

## Comparison with Other Providers

| Provider | Speed | Quality | Cost | Best For |
|----------|-------|---------|------|----------|
| **Z.ai** | Fast | High | Competitive | Chinese language, coding |
| **Zhipu** | Fast | High | Competitive | Chinese language |
| **OpenRouter** | Variable | Variable | Variable | Access to multiple models |
| **OpenAI** | Fast | High | Higher | English language, GPT-4 |

## API Reference

### Base URL

```
https://api.z.ai/api/coding/paas/v4/chat/completions
```

### Authentication

```
Authorization: Bearer YOUR_ZAI_API_KEY
```

### Request Format

Z.ai uses an OpenAI-compatible API format:

```json
{
  "model": "glm-4-plus",
  "messages": [
    {"role": "system", "content": "You are a helpful assistant"},
    {"role": "user", "content": "Your question here"}
  ],
  "max_tokens": 8192,
  "temperature": 0.7,
  "top_p": 0.95,
  "stream": false
}
```

### Response Format

```json
{
  "choices": [
    {
      "message": {
        "content": "Response here",
        "reasoning_content": "Thinking process (if applicable)"
      }
    }
  ],
  "usage": {
    "prompt_tokens": 100,
    "completion_tokens": 500
  }
}
```

## Additional Resources

- **Z.ai Website**: [https://z.ai](https://z.ai)
- **API Documentation**: Available in your Z.ai dashboard
- **Model Pricing**: Check your Z.ai account for current pricing
- **Community**: Join the [PicoClaw Discord](https://discord.gg/V4sAZ9XWpN) for community support

## Integration with Chat Apps

Z.ai works seamlessly with all PicoClaw chat integrations:

### Telegram

```json
{
  "channels": {
    "telegram": {
      "enabled": true,
      "token": "YOUR_BOT_TOKEN",
      "allow_from": ["YOUR_USER_ID"]
    }
  },
  "providers": {
    "zai": {
      "api_key": "${ZAI_API_KEY}"
    }
  }
}
```

### Discord

```json
{
  "channels": {
    "discord": {
      "enabled": true,
      "token": "YOUR_BOT_TOKEN",
      "allow_from": ["YOUR_USER_ID"]
    }
  },
  "providers": {
    "zai": {
      "api_key": "${ZAI_API_KEY}"
    }
  }
}
```

## Best Practices

1. **Use Environment Variables**: Never hardcode API keys in config files
2. **Choose the Right Model**: Match model to task complexity
3. **Monitor Usage**: Track token usage to manage costs
4. **Handle Errors**: Implement proper error handling in your workflows
5. **Cache Responses**: For repeated queries, consider caching results

## Security Considerations

1. **API Key Storage**: Always use environment variables or secure secret management
2. **Access Control**: Use `allow_from` in channel configs to restrict access
3. **Workspace Restrictions**: Enable `restrict_to_workspace` for file operations
4. **HTTPS**: Ensure all API communications use HTTPS

## Updates and Changelog

### Version 1.0 (2026-02-20)
- Initial Z.ai provider integration
- Support for GLM-4 model family
- OpenAI-compatible API interface
- Environment variable support
- Full documentation

## Support

For issues or questions about Z.ai integration:

1. Check the [main PicoClaw README](../README.md)
2. Review [troubleshooting section](#troubleshooting)
3. Open an issue on [GitHub](https://github.com/sipeed/picoclaw/issues)
4. Join the [Discord community](https://discord.gg/V4sAZ9XWpN)
