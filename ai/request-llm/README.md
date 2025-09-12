# CRD Documentation Analysis Tool

An AI-powered tool that analyzes Custom Resource Definition (CRD) changes and automatically identifies OpenShift documentation files that need updates. Built with Go and Anthropic's Claude AI model.

## 🚀 Key Features

- **🎯 CRD Change Detection**: Automatically detects changes in `config/crd/bases/` directory
- **🔍 Smart Documentation Search**: Searches OpenShift documentation for relevant files  
- **🤖 AI-Powered Analysis**: Uses Claude to provide specific recommendations
- **✅ Verified URLs**: Returns only real, existing documentation URLs
- **⚡ GitHub Integration**: Runs automatically on PRs with non-blocking analysis

## 🔄 GitHub Workflow Integration

This tool is integrated with GitHub Actions to provide automatic documentation analysis on pull requests. When you modify CRD files, the workflow:

1. Automatically runs on PRs that change CRD files
2. Analyzes the specific changes using AI
3. Searches OpenShift docs for potentially impacted files  
4. Posts detailed recommendations as PR comments
5. Provides verified URLs for documentation updates

**See [GITHUB_INTEGRATION.md](GITHUB_INTEGRATION.md) for complete setup instructions.**

### 🧪 Quick Test

To test the GitHub workflow locally:

```bash
cd ai/request-llm
export ANTHROPIC_API_KEY="your-api-key"
./test-workflow.sh
```

**Note:** The tool now automatically detects the git repository root, so it works from any directory. However, for GitHub Actions compatibility, it's recommended to run from the project root:

```bash
# From project root (matches GitHub Actions)
export ANTHROPIC_API_KEY="your-api-key"
go run -mod=mod ai/request-llm/main.go
```

## 🛠️ What This Tool Does

This project demonstrates advanced integration between git analysis, AI processing, and documentation management. It was originally intended to use LangChainGo, but we discovered model identifier compatibility issues. The official Anthropic SDK provides more reliable access to Claude models.

## Prerequisites

- Go 1.21 or higher
- Anthropic API key

## Getting an Anthropic API Key

To use this project, you'll need to get an Anthropic API key:

1. **Sign up for Anthropic**: Go to [https://console.anthropic.com/](https://console.anthropic.com/)
2. **Create an account**: Use your email to sign up for an account
3. **Verify your account**: Check your email and verify your account
4. **Add billing information**: You'll need to add payment information to use the API
5. **Generate API key**: 
   - Go to your account settings or API keys section
   - Click "Create Key" or similar
   - Copy the generated key (it will start with `sk-ant-`)
6. **Set up billing**: Make sure you have credits or a payment method set up

⚠️ **Important**: Keep your API key secure and never commit it to version control!

## Setup Instructions

1. **Clone/Initialize the project**:
   ```bash
   cd /home/rauherna/gorepo/request-llm
   ```

2. **Install dependencies**:
   ```bash
   go mod tidy
   ```

3. **Set your API key**:
   ```bash
   export ANTHROPIC_API_KEY="your-api-key-here"
   ```
   
   Or add it to your shell profile (`.bashrc`, `.zshrc`, etc.):
   ```bash
   echo 'export ANTHROPIC_API_KEY="your-api-key-here"' >> ~/.bashrc
   source ~/.bashrc
   ```

4. **Run the application**:
   ```bash
   go run main.go
   ```

## Project Structure

```
request-llm/
├── main.go          # Main application file
├── go.mod          # Go module dependencies
├── README.md       # This file
└── .env.example    # Example environment variables
```

## Code Overview

The project uses:
- **Official Anthropic Go SDK**: Direct integration with Anthropic's API
- **Claude 3 Haiku**: Currently working model (fast and cost-effective)
- **Environment variables**: For secure API key management

The main application:
1. Reads the API key from the `ANTHROPIC_API_KEY` environment variable
2. Initializes the Anthropic client using the official SDK
3. Creates a message request with the specified model
4. Sends a prompt asking for creative company names
5. Displays Claude's AI-generated response

## Model Information

This project currently uses `claude-3-haiku-20240307`, which is:
- A fast and cost-effective Claude model
- Good for most conversational tasks
- Well-established and reliable

You can change the model by modifying the `Model` field in the `MessagesRequest` in `main.go`.

## Available Claude Models

**Currently Working Models:**
- `claude-3-haiku-20240307` - Fast and cost-effective (currently used)

**Other Models to Try:**
- `claude-3-opus-20240229` - Most capable Claude-3 model, highest cost
- `claude-3-sonnet-20240229` - Balanced performance and cost
- `claude-3-5-sonnet-20240620` - Newer Sonnet model (may require higher API tier)

**⚠️ Important Notes:**
- Model availability depends on your Anthropic API tier and account status
- Some models may return 404 errors if not available for your account
- Check [Anthropic's documentation](https://docs.anthropic.com/claude/docs/models-overview) for current model availability
- Claude-4 models may be available but identifiers are still unclear

## Testing Results - What We Tried

During development, we tested both LangChainGo and the official Anthropic SDK with various models:

### LangChain Issues ❌
- **LangChainGo**: Returned 404 errors for ALL models we tested including:
  - `claude-3-sonnet-20240229`
  - `claude-3-opus-20240229`
  - `claude-3-5-sonnet-20240620`
  - `claude-instant-1.2`
- **Conclusion**: LangChainGo appears to have model identifier compatibility issues

### Official Anthropic SDK Results ✅❌
- ✅ **claude-3-haiku-20240307**: ✨ **WORKS PERFECTLY**
- ❌ **claude-3-opus-20240229**: Returns 404 error (likely requires higher API tier)
- ❌ **claude-3-sonnet-20240229**: Returns 404 error (may require higher API tier)
- ❌ **claude-3-5-sonnet-20240620**: Returns 404 error (may require higher API tier)

### Key Findings
1. **Official SDK > LangChain**: The official Anthropic SDK is more reliable than LangChain for Claude integration
2. **Model Tiers**: Higher-performance models (Opus, Sonnet) may require paid API tiers
3. **Haiku Works**: Claude-3 Haiku is accessible with basic API access and works great
4. **Account Dependent**: Model access varies by account type and payment status

## How to Upgrade to Newer Models

To try newer models like Claude 3.5 Sonnet:

1. **Update the model in `main.go`:**
   ```go
   Model: "claude-3-5-sonnet-20240620", // Try this identifier
   ```

2. **Check your API access:** Some models require higher API tiers

3. **Handle errors gracefully:** Add fallback models if newer ones aren't available

## Customization

To modify the prompt, edit the `prompt` variable in `main.go`:

```go
prompt := "Your custom prompt here"
```

## Error Handling

The application includes basic error handling for:
- Missing API key
- LLM initialization failures
- API call failures

## Security Notes

- Never hardcode your API key in the source code
- Use environment variables for sensitive configuration
- Consider using a `.env` file for local development (but don't commit it!)
- Rotate your API keys regularly

## Troubleshooting

**"ANTHROPIC_API_KEY environment variable is required"**
- Make sure you've set the environment variable correctly
- Check that you're using the correct variable name
- Restart your terminal after setting the variable

**API errors**:
- Verify your API key is correct and active
- Check your Anthropic account has sufficient credits
- Ensure you have proper billing set up

**Go module errors**:
- Run `go mod tidy` to ensure dependencies are properly installed
- Make sure you're using Go 1.21 or higher

## Next Steps

This is a basic example. You can extend it by:
- Adding conversation history
- Implementing streaming responses
- Adding different prompt templates
- Creating a web API wrapper
- Adding configuration files
- Implementing retry logic with exponential backoff

## Resources

- [LangChainGo Documentation](https://tmc.github.io/langchaingo/)
- [Anthropic API Documentation](https://docs.anthropic.com/)
- [Claude Models Guide](https://docs.anthropic.com/claude/docs/models-overview)
