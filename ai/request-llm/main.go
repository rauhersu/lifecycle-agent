package main

import (
	"context"
	"fmt"
	"io/fs"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
)

// truncateString truncates a string to a maximum length
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "\n... [TRUNCATED - content too long]"
}

// filterGitDiffByCRD filters git diff to only include changes under config/crd/bases/
func filterGitDiffByCRD(gitDiff string) string {
	lines := strings.Split(gitDiff, "\n")
	var filteredLines []string
	var inRelevantFile bool

	for _, line := range lines {
		if strings.HasPrefix(line, "diff --git") {
			// Check if this file is under config/crd/bases/
			parts := strings.Split(line, " ")
			if len(parts) >= 4 {
				// Handle different git diff prefixes: a/, i/, etc.
				file := parts[2]
				if strings.Contains(file, "/") {
					// Extract filename after the prefix (a/, i/, b/, w/, etc.)
					slashIdx := strings.Index(file, "/")
					if slashIdx >= 0 {
						file = file[slashIdx+1:]
					}
				}
				inRelevantFile = strings.HasPrefix(file, "config/crd/bases/")
			}
		}

		// Include line if we're in a relevant file
		if inRelevantFile {
			filteredLines = append(filteredLines, line)
		}
	}

	return strings.Join(filteredLines, "\n")
}

// summarizeGitDiff creates a concise summary of the git diff
func summarizeGitDiff(gitDiff string) string {
	lines := strings.Split(gitDiff, "\n")
	var summary strings.Builder
	var changedLines []string

	// Extract key information
	for _, line := range lines {
		if strings.HasPrefix(line, "diff --git") {
			summary.WriteString(line + "\n")
		} else if strings.HasPrefix(line, "@@") {
			summary.WriteString(line + "\n")
		} else if strings.HasPrefix(line, "+") && !strings.HasPrefix(line, "+++") {
			changedLines = append(changedLines, line)
		} else if strings.HasPrefix(line, "-") && !strings.HasPrefix(line, "---") {
			changedLines = append(changedLines, line)
		}
	}

	// Add the most relevant changed lines (limit to avoid token overflow)
	summary.WriteString("\nKey changes:\n")
	maxChanges := 20 // Limit the number of change lines
	for i, line := range changedLines {
		if i >= maxChanges {
			summary.WriteString("... [additional changes truncated]\n")
			break
		}
		summary.WriteString(line + "\n")
	}

	return summary.String()
}

// filterRelevantFiles extracts only CRD-related file changes from git diff
func filterRelevantFiles(gitDiff string) []string {
	var files []string
	lines := strings.Split(gitDiff, "\n")

	for _, line := range lines {
		if strings.HasPrefix(line, "diff --git") {
			parts := strings.Split(line, " ")
			if len(parts) >= 4 {
				// Handle different git diff prefixes: a/, i/, etc.
				file := parts[2]
				if strings.Contains(file, "/") {
					// Extract filename after the prefix (a/, i/, b/, w/, etc.)
					slashIdx := strings.Index(file, "/")
					if slashIdx >= 0 {
						file = file[slashIdx+1:]
					}
				}

				// Only include files under config/crd/bases/
				if strings.HasPrefix(file, "config/crd/bases/") {
					files = append(files, file)
				}
			}
		}
	}

	return files
}

// cloneOpenShiftDocs clones the openshift-docs repository to a temporary directory
func cloneOpenShiftDocs() (string, error) {
	// Create a temporary directory
	tempDir, err := os.MkdirTemp("", "openshift-docs-*")
	if err != nil {
		return "", fmt.Errorf("failed to create temp directory: %w", err)
	}

	fmt.Printf("Cloning OpenShift docs to %s...\n", tempDir)

	// Clone the repository with shallow depth for faster download
	cmd := exec.Command("git", "clone", "--depth", "1", "--branch", "enterprise-4.19",
		"https://github.com/openshift/openshift-docs.git", tempDir)
	output, err := cmd.CombinedOutput()
	if err != nil {
		os.RemoveAll(tempDir) // Cleanup on error
		return "", fmt.Errorf("failed to clone openshift-docs: %w\nOutput: %s", err, string(output))
	}

	return tempDir, nil
}

// searchRelevantDocsFiles searches for documentation files related to the changes in specific directories
func searchRelevantDocsFiles(docsDir string, changedFiles []string, gitDiff string) ([]string, error) {
	var relevantFiles []string

	// Extract search terms from the changes
	searchTerms := extractSearchTermsFromDiff(gitDiff, changedFiles)

	// Define the specific directories to search
	targetDirs := []string{
		filepath.Join(docsDir, "edge_computing"),
		filepath.Join(docsDir, "modules"),
	}

	// Search only in the specified directories
	for _, targetDir := range targetDirs {
		if _, err := os.Stat(targetDir); os.IsNotExist(err) {
			fmt.Printf("Warning: Directory %s does not exist, skipping...\n", targetDir)
			continue
		}

		err := filepath.WalkDir(targetDir, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}

			// Only process .adoc files
			if !d.IsDir() && strings.HasSuffix(strings.ToLower(d.Name()), ".adoc") {
				// Read file content to check for relevant terms
				content, readErr := os.ReadFile(path)
				if readErr != nil {
					// Skip files we can't read, don't fail the whole process
					return nil
				}

				contentStr := strings.ToLower(string(content))

				// Check if file contains any of our search terms
				for _, term := range searchTerms {
					if strings.Contains(contentStr, strings.ToLower(term)) {
						// Convert absolute path to relative path from docs root
						relPath, _ := filepath.Rel(docsDir, path)
						relevantFiles = append(relevantFiles, relPath)
						break // Found one match, no need to check other terms for this file
					}
				}
			}

			return nil
		})

		if err != nil {
			return nil, fmt.Errorf("failed to search in directory %s: %w", targetDir, err)
		}
	}

	return relevantFiles, nil
}

// extractSearchTermsFromDiff extracts CRD-focused search terms targeting detailed API documentation
func extractSearchTermsFromDiff(gitDiff string, changedFiles []string) []string {
	terms := make(map[string]bool) // Use map to avoid duplicates

	// Add CRD API documentation specific terms
	terms["custom resource"] = true
	terms["api reference"] = true
	terms["configuration reference"] = true
	terms["crd"] = true
	terms["custom resource definition"] = true
	terms["field descriptions"] = true
	terms["schema"] = true
	terms["yaml"] = true

	// Add CRD-specific terms based on changed file paths
	for _, file := range changedFiles {
		// Focus on the specific CRD types that changed
		if strings.Contains(file, "imagebasedupgrade") {
			terms["imagebasedupgrade"] = true
			terms["image-based-upgrade"] = true
			terms["image based upgrade"] = true
			terms["ibu"] = true
		}
		if strings.Contains(file, "seedgenerator") {
			terms["seedgenerator"] = true
			terms["seed-generator"] = true
			terms["seed generator"] = true
		}
	}

	// Extract specific API field names and values from CRD diff content for precise matching
	lines := strings.Split(gitDiff, "\n")
	for _, line := range lines {
		if strings.Contains(line, "+") || strings.Contains(line, "-") {
			// Look for specific enum values and field names that would appear in documentation
			if strings.Contains(strings.ToLower(line), "stage") {
				terms["stage"] = true
			}
			if strings.Contains(strings.ToLower(line), "idle") {
				terms["idle"] = true
			}
			if strings.Contains(strings.ToLower(line), "nop") {
				terms["nop"] = true
			}
			if strings.Contains(strings.ToLower(line), "rollback") {
				terms["rollback"] = true
			}
			if strings.Contains(strings.ToLower(line), "spec") {
				terms["spec"] = true
			}
			if strings.Contains(strings.ToLower(line), "status") {
				terms["status"] = true
			}
		}
	}

	// Convert map keys to slice, prioritize CRD documentation specific terms
	var result []string
	priorityOrder := []string{
		"custom resource", "api reference", "crd", "configuration reference",
		"imagebasedupgrade", "image-based-upgrade", "ibu",
		"seedgenerator", "seed-generator", "schema", "yaml",
	}

	// Add priority terms first
	for _, term := range priorityOrder {
		if terms[term] {
			result = append(result, term)
			delete(terms, term)
		}
	}

	// Add remaining field-specific terms
	for term := range terms {
		result = append(result, term)
	}

	// Focus on the most relevant terms for CRD documentation
	if len(result) > 10 {
		result = result[:10]
	}

	return result
}

// getGitDiff executes git diff and returns the output
func getGitDiff() (string, error) {
	// Get the repository root directory
	repoRoot := "/home/rauherna/gorepo/lifecycle-agent.fork"

	// Check if we're running in GitHub Actions
	if os.Getenv("GITHUB_ACTIONS") == "true" {
		return getGitDiffForGitHubActions(repoRoot)
	}

	// Original logic for local development
	return getGitDiffLocal(repoRoot)
}

// getGitDiffForGitHubActions handles git diff in GitHub Actions context
func getGitDiffForGitHubActions(repoRoot string) (string, error) {
	// In GitHub Actions, we want to compare the current HEAD against the base branch
	baseBranch := os.Getenv("GITHUB_BASE_REF") // e.g., "main"
	if baseBranch == "" {
		baseBranch = "main" // fallback
	}

	// Ensure we have the base branch reference
	cmd := exec.Command("git", "fetch", "origin", baseBranch)
	cmd.Dir = repoRoot
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("failed to fetch base branch %s: %w", baseBranch, err)
	}

	// Compare current HEAD against the base branch
	baseRef := fmt.Sprintf("origin/%s", baseBranch)
	cmd = exec.Command("git", "diff", baseRef, "HEAD")
	cmd.Dir = repoRoot
	output, err := cmd.Output()
	if err != nil {
		// Fallback: try without origin prefix
		cmd = exec.Command("git", "diff", baseBranch, "HEAD")
		cmd.Dir = repoRoot
		output, err = cmd.Output()
		if err != nil {
			return "", fmt.Errorf("failed to run git diff against %s: %w", baseBranch, err)
		}
	}

	return string(output), nil
}

// getGitDiffLocal handles git diff for local development
func getGitDiffLocal(repoRoot string) (string, error) {
	// Run git diff to get unstaged changes
	cmd := exec.Command("git", "diff")
	cmd.Dir = repoRoot
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to run git diff: %w", err)
	}

	diff := string(output)

	// If no unstaged changes, try staged changes
	if strings.TrimSpace(diff) == "" {
		cmd = exec.Command("git", "diff", "--cached")
		cmd.Dir = repoRoot
		output, err = cmd.Output()
		if err != nil {
			return "", fmt.Errorf("failed to run git diff --cached: %w", err)
		}
		diff = string(output)
	}

	// If still no changes, try comparing with origin/main
	if strings.TrimSpace(diff) == "" {
		cmd = exec.Command("git", "diff", "origin/main", "HEAD")
		cmd.Dir = repoRoot
		output, err = cmd.Output()
		if err != nil {
			// If origin/main doesn't exist, try main
			cmd = exec.Command("git", "diff", "main", "HEAD")
			cmd.Dir = repoRoot
			output, err = cmd.Output()
			if err != nil {
				return "", fmt.Errorf("failed to run git diff against main: %w", err)
			}
		}
		diff = string(output)
	}

	return diff, nil
}

func main() {
	ctx := context.Background()

	// Check if API key is provided
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		log.Fatal("ANTHROPIC_API_KEY environment variable is required. Please set it with your Anthropic API key.")
	}

	// Get git diff to see what has changed
	fmt.Println("Getting git diff...")
	fullGitDiff, err := getGitDiff()
	if err != nil {
		log.Fatalf("Error getting git diff: %v", err)
	}

	if strings.TrimSpace(fullGitDiff) == "" {
		fmt.Println("No changes detected in git diff. Make sure you have uncommitted changes or commits ahead of main.")
		return
	}

	// Filter git diff to only include CRD changes (config/crd/bases/)
	gitDiff := filterGitDiffByCRD(fullGitDiff)
	if strings.TrimSpace(gitDiff) == "" {
		fmt.Println("No changes detected in config/crd/bases/ directory.")
		fmt.Println("This tool only analyzes CRD (Custom Resource Definition) changes that impact API documentation.")
		fmt.Println("Changes in other directories are not analyzed.")
		return
	}

	// Extract CRD-related changed files for context
	changedFiles := filterRelevantFiles(gitDiff)
	fmt.Printf("CRD files changed in config/crd/bases/: %v\n", changedFiles)

	// Clone OpenShift docs repository
	docsDir, err := cloneOpenShiftDocs()
	if err != nil {
		log.Fatalf("Error cloning OpenShift docs: %v", err)
	}

	// Ensure cleanup of temporary directory
	defer func() {
		fmt.Printf("Cleaning up temporary directory: %s\n", docsDir)
		os.RemoveAll(docsDir)
	}()

	// Search for CRD-specific documentation files (API references, field descriptions, schemas)
	searchTerms := extractSearchTermsFromDiff(gitDiff, changedFiles)
	fmt.Printf("Searching for detailed CRD documentation using terms: %v\n", searchTerms)
	relevantDocFiles, err := searchRelevantDocsFiles(docsDir, changedFiles, gitDiff)
	if err != nil {
		log.Fatalf("Error searching docs files: %v", err)
	}

	// Filter and prioritize files that likely contain detailed CRD information
	maxFiles := 15
	if len(relevantDocFiles) > maxFiles {
		fmt.Printf("Found %d CRD documentation files (showing top %d most relevant):\n", len(relevantDocFiles), maxFiles)
		relevantDocFiles = relevantDocFiles[:maxFiles]
	} else {
		fmt.Printf("Found %d CRD documentation files:\n", len(relevantDocFiles))
	}

	for _, file := range relevantDocFiles {
		fmt.Printf("  - %s\n", file)
	}

	// Initialize Anthropic client using official SDK
	client := anthropic.NewClient(
		option.WithAPIKey(apiKey),
	)

	baseURL := "https://github.com/openshift/openshift-docs/blob/enterprise-4.19"

	// Build verified URLs for the found documentation files
	var verifiedURLs []string
	for _, file := range relevantDocFiles {
		url := fmt.Sprintf("%s/%s", baseURL, file)
		verifiedURLs = append(verifiedURLs, url)
	}

	// Create a concise summary of the git diff to avoid token limits
	gitDiffSummary := summarizeGitDiff(gitDiff)

	// Truncate the summary if still too long
	gitDiffSummary = truncateString(gitDiffSummary, 5000) // Limit to ~5000 characters

	prompt := fmt.Sprintf(`Given the following CRD (Custom Resource Definition) changes to the lifecycle-agent project, analyze what specific CRD documentation files need updates.

**CRD files changed in config/crd/bases/:** %v

**Found CRD documentation files that explain detailed API information (VERIFIED TO EXIST):**
%s

**Verified URLs containing CRD details to update:**
%s

Please focus SPECIFICALLY on files that contain detailed CRD information such as:
- API field descriptions and schemas
- Custom Resource configuration references  
- YAML example configurations
- Field validation rules and enum values
- CRD specification documentation

For each file, provide:
1. **CRD Field Updates**: Which specific CRD field descriptions need changes
2. **Schema Changes**: What schema documentation needs updating (enum values, validation rules)
3. **YAML Examples**: Exact YAML configuration examples that need modification
4. **Field Descriptions**: Specific field description text that needs updates
5. **Priority**: Which CRD documentation aspects are most critical to update

**Summary of CRD changes:**
%s

IMPORTANT: Focus ONLY on files that contain detailed CRD schemas, field descriptions, API references, and configuration examples. Ignore general usage guides - target only the detailed technical CRD documentation.`,
		changedFiles,
		strings.Join(relevantDocFiles, "\n"),
		strings.Join(verifiedURLs, "\n"),
		gitDiffSummary)

	// Estimate token count (rough estimation: 1 token ≈ 4 characters)
	estimatedTokens := len(prompt) / 4
	fmt.Printf("Estimated prompt tokens: %d (limit: 200,000)\n", estimatedTokens)

	if estimatedTokens > 180000 { // Leave some buffer
		log.Fatalf("Prompt is likely too long (%d estimated tokens). Consider reducing the number of files or git diff size.", estimatedTokens)
	}

	fmt.Printf("Prompt: %s\n\n", truncateString(prompt, 1000)) // Show truncated version for debugging
	fmt.Println("Claude's response:")

	// Create message request with Claude Sonnet 4 model
	message, err := client.Messages.New(ctx, anthropic.MessageNewParams{
		MaxTokens: 2000, // Increased for detailed analysis
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock(prompt)),
		},
		Model: anthropic.ModelClaudeSonnet4_20250514,
	})
	if err != nil {
		log.Fatal(err)
	}

	// Print the response
	for _, contentBlock := range message.Content {
		if contentBlock.Type == "text" {
			fmt.Println(contentBlock.Text)
		}
	}

	// Print summary of verified URLs
	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Println("SUMMARY - CRD Documentation Files to Update (Detailed API Information):")
	fmt.Println(strings.Repeat("=", 80))

	if len(verifiedURLs) > 0 {
		fmt.Printf("Found %d CRD documentation files containing detailed API information:\n\n", len(verifiedURLs))
		for i, url := range verifiedURLs {
			fmt.Printf("%d. %s\n", i+1, url)
		}

		fmt.Println("\n" + strings.Repeat("-", 80))
		fmt.Println("NEXT STEPS:")
		fmt.Println("1. Review each URL above for detailed CRD field descriptions and schemas")
		fmt.Println("2. Clone the docs repository to make edits:")
		fmt.Println("   git clone https://github.com/openshift/openshift-docs.git")
		fmt.Println("   cd openshift-docs && git checkout enterprise-4.19")
		fmt.Println("3. Update CRD field descriptions, enum values, and YAML examples")
		fmt.Println("4. Verify schema documentation matches the actual CRD changes")
		fmt.Println("5. Test documentation builds after making changes")
		fmt.Println("6. Submit PR with your CRD documentation updates")
		fmt.Printf("\nNote: Focused on files containing detailed CRD schemas and API field descriptions.\n")
		fmt.Printf("Search targeted CRD-specific documentation in edge_computing/ and modules/.\n")
	} else {
		fmt.Println("No detailed CRD documentation files found based on the changes.")
		fmt.Println("This could mean:")
		fmt.Println("- The CRD changes don't have corresponding detailed field documentation")
		fmt.Println("- The CRD documentation uses different terminology or is in other directories")
		fmt.Println("- The API reference documentation might be auto-generated from the CRD")
		fmt.Println("- Check for API reference files in other directories (assemblies/, etc.)")
	}
}
