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

// filterRelevantFiles extracts relevant file patterns from git diff
func filterRelevantFiles(gitDiff string) []string {
	var files []string
	lines := strings.Split(gitDiff, "\n")

	for _, line := range lines {
		if strings.HasPrefix(line, "diff --git") {
			parts := strings.Split(line, " ")
			if len(parts) >= 4 {
				// Extract the file path (remove a/ and b/ prefixes)
				file := strings.TrimPrefix(parts[2], "a/")
				files = append(files, file)
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

// searchRelevantDocsFiles searches for documentation files related to the changes
func searchRelevantDocsFiles(docsDir string, changedFiles []string, gitDiff string) ([]string, error) {
	var relevantFiles []string

	// Extract search terms from the changes
	searchTerms := extractSearchTermsFromDiff(gitDiff, changedFiles)

	// Walk through all .adoc files in the docs directory
	err := filepath.WalkDir(docsDir, func(path string, d fs.DirEntry, err error) error {
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
		return nil, fmt.Errorf("failed to search docs files: %w", err)
	}

	return relevantFiles, nil
}

// extractSearchTermsFromDiff extracts relevant search terms from git diff and file paths
func extractSearchTermsFromDiff(gitDiff string, changedFiles []string) []string {
	terms := make(map[string]bool) // Use map to avoid duplicates

	// Add terms based on changed file paths
	for _, file := range changedFiles {
		if strings.Contains(file, "imagebasedupgrade") {
			terms["imagebasedupgrade"] = true
			terms["image-based-upgrade"] = true
			terms["image based upgrade"] = true
			terms["ibu"] = true
		}
		if strings.Contains(file, "lifecycle") {
			terms["lifecycle-agent"] = true
			terms["lifecycle agent"] = true
		}
		if strings.Contains(file, "seedgenerator") {
			terms["seedgenerator"] = true
			terms["seed-generator"] = true
			terms["seed generator"] = true
		}
	}

	// Extract terms from diff content (look for struct names, constants, etc.)
	lines := strings.Split(gitDiff, "\n")
	for _, line := range lines {
		if strings.Contains(line, "+") || strings.Contains(line, "-") {
			// Look for Go identifiers, constants, etc.
			if strings.Contains(line, "Idle") {
				terms["idle"] = true
			}
			if strings.Contains(line, "Nop") {
				terms["nop"] = true
			}
			if strings.Contains(line, "Stage") {
				terms["stage"] = true
			}
		}
	}

	// Convert map keys to slice
	var result []string
	for term := range terms {
		result = append(result, term)
	}

	return result
}

// getGitDiff executes git diff and returns the output
func getGitDiff() (string, error) {
	// Get the repository root directory
	repoRoot := "/home/rauherna/gorepo/lifecycle-agent.fork"

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
	gitDiff, err := getGitDiff()
	if err != nil {
		log.Fatalf("Error getting git diff: %v", err)
	}

	if strings.TrimSpace(gitDiff) == "" {
		fmt.Println("No changes detected in git diff. Make sure you have uncommitted changes or commits ahead of main.")
		return
	}

	// Extract changed files for better context
	changedFiles := filterRelevantFiles(gitDiff)
	fmt.Printf("Changed files detected: %v\n", changedFiles)

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

	// Search for relevant documentation files
	fmt.Println("Searching for relevant documentation files...")
	relevantDocFiles, err := searchRelevantDocsFiles(docsDir, changedFiles, gitDiff)
	if err != nil {
		log.Fatalf("Error searching docs files: %v", err)
	}

	fmt.Printf("Found %d relevant documentation files:\n", len(relevantDocFiles))
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

	prompt := fmt.Sprintf(`Given the following git diff showing changes to the lifecycle-agent project, analyze what specific OpenShift documentation files would be impacted.

**Changed files in lifecycle-agent:** %v

**Found relevant documentation files (VERIFIED TO EXIST):**
%s

**Verified URLs to update:**
%s

Please provide:
1. **Specific File Analysis**: For each relevant documentation file above, explain what changes would be needed
2. **Content Updates**: What specific content changes are required based on the git diff
3. **Priority**: Which files are most critical to update first
4. **Change Details**: Exact text/examples that need to be updated

**Git diff showing the actual changes:**
%s

Focus on specific, actionable recommendations for updating each of the verified documentation files listed above.`,
		changedFiles,
		strings.Join(relevantDocFiles, "\n"),
		strings.Join(verifiedURLs, "\n"),
		gitDiff)

	fmt.Printf("Prompt: %s\n\n", prompt)
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
	fmt.Println("SUMMARY - Verified Documentation Files to Update:")
	fmt.Println(strings.Repeat("=", 80))

	if len(verifiedURLs) > 0 {
		fmt.Printf("Found %d relevant documentation files that need updates:\n\n", len(verifiedURLs))
		for i, url := range verifiedURLs {
			fmt.Printf("%d. %s\n", i+1, url)
		}

		fmt.Println("\n" + strings.Repeat("-", 80))
		fmt.Println("NEXT STEPS:")
		fmt.Println("1. Review each URL above and make the recommended changes")
		fmt.Println("2. Clone the docs repository to make edits:")
		fmt.Println("   git clone https://github.com/openshift/openshift-docs.git")
		fmt.Println("   cd openshift-docs && git checkout enterprise-4.19")
		fmt.Println("3. Test documentation builds after making changes")
		fmt.Println("4. Submit PR with your documentation updates")
	} else {
		fmt.Println("No relevant documentation files found.")
		fmt.Println("This could mean:")
		fmt.Println("- The changes don't impact user-facing documentation")
		fmt.Println("- The search terms need refinement")
		fmt.Println("- The documentation uses different terminology")
	}
}
