# CRD Documentation Check - GitHub Integration

This document explains how the CRD documentation analysis tool is integrated with GitHub Actions to provide automatic documentation impact analysis on pull requests.

## How It Works

The GitHub Action workflow (`.github/workflows/crd-doc-check.yaml`) automatically runs on PRs that modify:

- `config/crd/bases/**` - CRD definition files
- `api/**/v1/**` - API type definitions  
- `docs/**` - Documentation files
- `ai/request-llm/**` - The analysis tool itself

When triggered, the workflow:

1. **Sets up git context** - Ensures both PR branch and base branch are available
2. **Detects CRD Changes** - Compares PR branch against base branch for CRD changes
3. **Runs AI Analysis** - Uses the `main.go` tool to analyze documentation impact
4. **Posts PR Comment** - Adds analysis results as a PR comment
5. **Non-blocking** - Doesn't prevent PR merging

### ⚠️ **Important: Git Context Handling**

The tool automatically detects whether it's running in GitHub Actions vs local development:

- **Local Development**: Compares unstaged/staged changes or commits ahead of main
- **GitHub Actions**: Compares the PR branch against the PR's base branch using `GITHUB_BASE_REF`

This ensures the analysis works correctly in both environments.

## Setup Instructions

### 1. Add Anthropic API Key Secret

The workflow requires an Anthropic API key to function. Add it as a GitHub repository secret:

1. Go to your repository settings
2. Navigate to **Secrets and variables** → **Actions**
3. Click **New repository secret**
4. Name: `ANTHROPIC_API_KEY`
5. Value: Your Anthropic API key (starts with `sk-ant-`)

### 2. Enable Workflow Permissions

Ensure the workflow has the necessary permissions:

1. Go to repository **Settings** → **Actions** → **General**
2. Under **Workflow permissions**, select:
   - ✅ **Read and write permissions**
   - ✅ **Allow GitHub Actions to create and approve pull requests**

### 3. Test the Workflow

To test the integration:

1. Create a test branch:
   ```bash
   git checkout -b test-crd-doc-workflow
   ```

2. Make a small change to a CRD file:
   ```bash
   # Add a comment or modify a field description
   vim config/crd/bases/lca.openshift.io_imagebasedupgrades.yaml
   ```

3. Commit and push:
   ```bash
   git add config/crd/bases/lca.openshift.io_imagebasedupgrades.yaml
   git commit -m "test: trigger CRD documentation check"
   git push origin test-crd-doc-workflow
   ```

4. Create a PR and observe the workflow execution

## Workflow Output

The workflow generates a PR comment with:

### 🎯 Documentation Files to Review
- List of OpenShift documentation URLs that may need updates
- Specific files containing CRD field descriptions and schemas

### 🤖 AI Analysis Results
- Claude's detailed analysis of required changes
- Specific field descriptions and YAML examples to update
- Priority recommendations

### Example Output

```markdown
📋 **CRD Documentation Impact Analysis**

Based on the CRD changes in this PR, here are the documentation files that may need updates:

🎯 **Documentation Files to Review:**

1. https://github.com/openshift/openshift-docs/blob/enterprise-4.19/modules/ibu-imagebasedupgrade-custom-resource.adoc
2. https://github.com/openshift/openshift-docs/blob/enterprise-4.19/modules/ibu-api-reference.adoc

🤖 **AI Analysis Results:**
```
Based on the CRD changes showing "Rollback" to "RollbackTransaction" enum value change:

**CRD Field Updates Required:**
1. modules/ibu-imagebasedupgrade-custom-resource.adoc
   - Update stage field enum description: Change "Rollback" to "RollbackTransaction"
   - Line 45: Update field validation table
```
```

## Local Testing

You can run the analysis locally before creating a PR:

```bash
cd ai/request-llm
export ANTHROPIC_API_KEY="your-api-key"
go run main.go
```

This helps verify the analysis works correctly before triggering the workflow.

## Workflow Configuration

### Triggers

The workflow runs on:
- **Pull request events**: `opened`, `synchronize`, `reopened`
- **Path filters**: Only when relevant files change
- **Draft exclusion**: Skips draft PRs

### Customization

To modify the workflow behavior:

1. **Change trigger paths**: Edit the `paths:` section in the workflow
2. **Adjust comment format**: Modify the `Process Analysis Results` step
3. **Add more file types**: Update the path filters

### Rate Limiting

The workflow:
- ✅ Only runs on relevant file changes (not every PR)
- ✅ Updates existing comments instead of creating new ones
- ✅ Uploads logs as artifacts for debugging
- ✅ Has proper error handling

## Benefits

### For Developers
- 🎯 **Automatic detection** of documentation impact
- 📝 **Specific recommendations** for updates needed
- 🔍 **AI-powered analysis** of complex CRD changes
- ⚡ **Non-blocking** - doesn't slow down development

### For Documentation Maintainers
- 📋 **Clear tracking** of changes requiring doc updates
- 🎯 **Specific URLs** for files needing attention
- 🤖 **Detailed analysis** of required changes
- 📊 **Audit trail** through PR comments

### For Project Quality
- 🛡️ **Consistent documentation** updates
- 🎯 **Reduced missed updates** when CRDs change
- 📝 **Better API documentation** maintenance
- 🔄 **Automated workflow** integration

## Troubleshooting

### Workflow Doesn't Run
- Check that the trigger paths match your changes
- Verify the workflow file is in `.github/workflows/`
- Ensure the PR is not a draft

### Analysis Fails
- Check that `ANTHROPIC_API_KEY` secret is set correctly
- Verify the API key has sufficient credits
- Check workflow logs for specific error messages

### No Results Posted (Most Common Issue)
- **Root Cause**: The tool says "No changes detected in config/crd/bases/ directory" even though you modified CRD files
- **Solution**: This was a git context issue that has been fixed. The tool now properly compares PR branches against their base branch
- **Debugging**: Check the "Setup git for comparison" step in workflow logs to verify branch setup

### Other Issues
- Ensure the tool finds CRD changes in `config/crd/bases/`
- Check that the OpenShift docs repository clone succeeds
- Verify network connectivity to external repositories

### Comment Not Posted
- Check workflow permissions (need `pull-requests: write`)
- Verify the GitHub token has necessary permissions
- Check for rate limiting issues

## Future Enhancements

Potential improvements:
- 📊 **Metrics tracking** for documentation coverage
- 🔄 **Auto-update PRs** for simple documentation fixes
- 📝 **Template generation** for new documentation sections
- 🎯 **Integration with documentation builds**
- 📋 **Dashboard** for tracking documentation debt
