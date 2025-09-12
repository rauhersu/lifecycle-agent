#!/bin/bash
set -e

# Test script for CRD Documentation Check workflow
# This script simulates what the GitHub Action does locally

echo "🔍 Testing CRD Documentation Check workflow locally..."

# Check prerequisites
if [[ -z "$ANTHROPIC_API_KEY" ]]; then
    echo "❌ Error: ANTHROPIC_API_KEY environment variable is required"
    echo "   Set it with: export ANTHROPIC_API_KEY='sk-ant-...'"
    exit 1
fi

if ! command -v go &> /dev/null; then
    echo "❌ Error: Go is not installed or not in PATH"
    exit 1
fi

# Check if we're in the right directory structure
if [[ ! -f "main.go" ]]; then
    echo "❌ Error: main.go not found. Run this script from the ai/request-llm directory"
    echo "   cd ai/request-llm && ./test-workflow.sh"
    exit 1
fi

# Check if we can find the project structure  
if [[ ! -d "../../config/crd/bases" ]]; then
    echo "❌ Error: Cannot find config/crd/bases directory from current location"
    echo "   Make sure you're in the correct project structure"
    exit 1
fi

project_root=$(git rev-parse --show-toplevel 2>/dev/null || echo "")
if [[ -z "$project_root" ]]; then
    echo "❌ Error: Cannot find git repository root"
    exit 1
fi

echo "📂 Project structure check:"
echo "   Current directory: $(pwd)"
echo "   Project root: $project_root"
echo "   Tool location: $project_root/ai/request-llm/main.go"
echo "   CRD directory: $project_root/config/crd/bases/"

# Check if we're in a git repository
if ! git rev-parse --git-dir > /dev/null 2>&1; then
    echo "❌ Error: Not in a git repository"
    exit 1
fi

# Test both local and GitHub Actions scenarios
echo "🔍 Testing in local development mode..."
current_branch=$(git branch --show-current)
echo "   Current branch: $current_branch"

# Check for local changes
if git diff --quiet HEAD && git diff --cached --quiet; then
    echo "   No uncommitted changes detected"
    
    # Check for commits ahead of main
    if git rev-list --count main..HEAD > /dev/null 2>&1; then
        commits_ahead=$(git rev-list --count main..HEAD)
        if [ "$commits_ahead" -gt 0 ]; then
            echo "   ✅ Found $commits_ahead commit(s) ahead of main - this will be analyzed"
        else
            echo "   ⚠️  No commits ahead of main"
        fi
    else
        echo "   ⚠️  Cannot compare with main branch"
    fi
    
    echo ""
    echo "   💡 To test with uncommitted changes:"
    echo "      vim ../../config/crd/bases/lca.openshift.io_imagebasedupgrades.yaml"
    echo ""
else
    echo "   ✅ Uncommitted changes detected - these will be analyzed"
fi

echo ""
echo "🔍 Simulating GitHub Actions environment..."
export GITHUB_ACTIONS=true
export GITHUB_BASE_REF="main"
echo "   Set GITHUB_ACTIONS=true"
echo "   Set GITHUB_BASE_REF=main"

echo "📋 Checking dependencies..."
go mod tidy

echo ""
echo "🚀 Running CRD Documentation Analysis..."
echo "   (This simulates exactly what the GitHub Action executes)"
echo ""

# Test from current directory first (should work now with dynamic repo root detection)
echo "🔍 Testing: go run main.go (from ai/request-llm/)"
set +e
output_local=$(go run main.go 2>&1)
exit_code_local=$?
set -e

echo "   Exit code: $exit_code_local"
echo ""

# Test from project root (GitHub Actions way)  
echo "🔍 Testing: go run -mod=mod ai/request-llm/main.go (from project root)"
cd "$project_root"
set +e
output=$(go run -mod=mod ai/request-llm/main.go 2>&1)
exit_code=$?
set -e

echo "   Exit code: $exit_code"
echo ""

# Use the project root execution results for analysis (matches GitHub Actions)
if [[ $exit_code -eq 0 ]]; then
    echo "✅ Both execution methods work! Using project root results (matches GitHub Actions):"
elif [[ $exit_code_local -eq 0 ]]; then
    echo "⚠️  Local execution works but project root failed. Using local results:"
    output="$output_local"
    exit_code=$exit_code_local
else
    echo "❌ Both execution methods failed. Showing both outputs:"
    echo ""
    echo "=== From ai/request-llm/ directory ==="
    echo "$output_local"
    echo ""
    echo "=== From project root ==="
    echo "$output"
fi

echo "Exit code: $exit_code"
echo "Output length: ${#output} characters"
echo ""

if [[ $exit_code -eq 0 && ${#output} -gt 100 ]]; then
    echo "✅ Analysis completed successfully!"
    echo ""
    echo "📄 Full output:"
    echo "================================================================================"
    echo "$output"
    echo "================================================================================"
    echo ""
    
    # Extract key sections (like the workflow does)
    if echo "$output" | grep -q "SUMMARY -"; then
        echo "🎯 Summary section found - this would be posted as a PR comment"
        echo ""
        
        # Show what would be in the PR comment
        echo "📋 PR Comment Preview:"
        echo "================================================================================"
        echo "📋 **CRD Documentation Impact Analysis**"
        echo ""
        echo "Based on the CRD changes in this PR, here are the documentation files that may need updates:"
        echo ""
        
        if echo "$output" | grep -A 20 "SUMMARY -" | grep -E "https://github.com/openshift/openshift-docs" > /tmp/urls.txt 2>/dev/null; then
            echo "🎯 **Documentation Files to Review:**"
            echo ""
            cat /tmp/urls.txt 2>/dev/null || echo "(No URLs found)"
            echo ""
            rm -f /tmp/urls.txt
        fi
        
        echo "🤖 **AI Analysis Results:**"
        echo '```'
        if echo "$output" | grep -A 50 "Claude's response:" | grep -B 50 "SUMMARY -" | head -n -1 > /tmp/recommendations.txt 2>/dev/null; then
            head -n 10 /tmp/recommendations.txt 2>/dev/null || echo "(No recommendations found)"
            if [[ -f /tmp/recommendations.txt && $(wc -l < /tmp/recommendations.txt 2>/dev/null || echo 0) -gt 10 ]]; then
                echo "... (truncated for readability)"
            fi
            rm -f /tmp/recommendations.txt
        fi
        echo '```'
        echo ""
        echo "---"
        echo "⚠️ **Note:** This is a non-blocking check. The analysis is powered by AI and may suggest files that don't require changes."
        echo "📝 Please review the suggested files and update documentation as needed."
        echo "================================================================================"
    else
        echo "⚠️  No summary section found in output"
    fi
    
elif [[ $exit_code -eq 0 ]]; then
    echo "ℹ️  Analysis completed but with minimal output"
    echo "   This usually means no CRD changes were detected"
    echo ""
    echo "📄 Output:"
    echo "$output"
    echo ""
    echo "💡 Try making changes to files in config/crd/bases/ directory"
    
else
    echo "❌ Analysis failed (exit code: $exit_code)"
    echo ""
    echo "📄 Error output:"
    echo "$output"
    echo ""
    echo "💡 Common issues:"
    echo "   - ANTHROPIC_API_KEY not set or invalid"
    echo "   - Network connectivity issues"
    echo "   - Git repository issues"
    echo "   - Go compilation errors"
fi

echo ""
echo "🔍 Testing complete!"
echo ""
echo "Next steps:"
echo "1. If this works locally, the GitHub Action should work too"
echo "2. Make sure ANTHROPIC_API_KEY is set as a repository secret"
echo "3. Create a PR with CRD changes to test the full workflow"
