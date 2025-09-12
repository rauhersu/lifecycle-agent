package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
)

func main() {
	ctx := context.Background()

	// Check if API key is provided
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		log.Fatal("ANTHROPIC_API_KEY environment variable is required. Please set it with your Anthropic API key.")
	}

	// Initialize Anthropic client using official SDK
	client := anthropic.NewClient(
		option.WithAPIKey(apiKey),
	)

	modifiedFile := "/home/rauherna/gorepo/lifecycle-agent.fork/api/imagebasedupgrade/v1/types.go"
	docURL := "https://github.com/openshift/openshift-docs/tree/enterprise-4.19"
	// docURL := "https://docs.redhat.com/documentation/openshift_container_platform/4.19"
	//docURL := "https://docs.openshift.com/container-platform/4.19"

	prompt := fmt.Sprintf(
		"Given a local file change to %s, explain what specific parts of the documentation at %s would be impacted. Be specific for the sections and provide exact URLs which must correspond to already existing in the previous github repo",
		modifiedFile,
		docURL,
	)

	//	prompt := "What would be a good company name for a company that makes colorful socks?"

	fmt.Printf("Prompt: %s\n\n", prompt)
	fmt.Println("Claude's response:")

	// Create message request with Claude Sonnet 4 model
	message, err := client.Messages.New(ctx, anthropic.MessageNewParams{
		MaxTokens: 600,
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
}
