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

This project provides intelligent analysis of Custom Resource Definition (CRD) changes and their impact on OpenShift documentation. The tool automatically:

1. **Detects CRD Changes**: Monitors git diff for changes in `config/crd/bases/` directory
2. **Downloads Documentation**: Clones the OpenShift docs repository for comprehensive analysis
3. **Intelligent Search**: Uses sophisticated algorithms to find relevant documentation files
4. **AI Analysis**: Leverages Claude Sonnet 4 to provide specific, actionable recommendations
5. **Verified Results**: Returns only real, existing documentation URLs that need updates

The tool bridges the gap between code changes and documentation maintenance, ensuring CRD modifications are properly reflected in user-facing documentation.

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

1. **Navigate to the project directory**:

   ```bash
   cd ai/request-llm
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

```text
request-llm/
├── main.go          # Main application file
├── go.mod          # Go module dependencies
├── README.md       # This file
└── .env.example    # Example environment variables
```

## Code Overview

The project uses:

- **Official Anthropic Go SDK**: Direct integration with Anthropic's API
- **Claude Sonnet 4**: Advanced model for comprehensive CRD analysis
- **Git integration**: Automatically detects CRD changes via git diff
- **OpenShift docs cloning**: Downloads and searches official documentation
- **Environment variables**: For secure API key management

The main application:

1. Reads the API key from the `ANTHROPIC_API_KEY` environment variable
2. Detects CRD changes by analyzing git diff in `config/crd/bases/` directory
3. Clones the OpenShift documentation repository for comprehensive search
4. Extracts relevant search terms from CRD changes
5. Searches documentation files for CRD-specific content (API references, schemas, field descriptions)
6. Uses Claude Sonnet 4 to analyze changes and provide specific documentation update recommendations
7. Returns verified URLs of documentation files that need updates

## Model Information

This project currently uses `claude-sonnet-4-20250514` (Claude Sonnet 4), which is:

- Anthropic's most advanced model for complex analysis tasks
- Excellent for technical documentation analysis and recommendations
- Optimized for understanding code changes and their implications
- Well-suited for CRD analysis and documentation correlation

## Customization

The tool automatically generates detailed prompts for CRD analysis. The main prompt is dynamically constructed in `main.go` based on:

- Detected CRD changes in `config/crd/bases/`
- Found documentation files in OpenShift docs
- Git diff analysis and search terms

You can customize the analysis by modifying:

- Search directories in the `targetDirs` variable
- Search terms extraction logic in `extractSearchTermsFromDiff()`
- Maximum number of documentation files analyzed
- Token limits and truncation thresholds


## Resources

- [Anthropic API Documentation](https://docs.anthropic.com/)
- [Claude Models Guide](https://docs.anthropic.com/claude/docs/models-overview)
