package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/liushuangls/go-anthropic/v2"
)

func main() {
	ctx := context.Background()

	// Check if API key is provided
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		log.Fatal("ANTHROPIC_API_KEY environment variable is required. Please set it with your Anthropic API key.")
	}

	// Initialize Anthropic client (using official SDK due to LangChain model issues)
	client := anthropic.NewClient(apiKey)

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

	// Create message request with Claude-3 Haiku model (tested and working)
	request := anthropic.MessagesRequest{
		Model: "claude-3-haiku-20240307",
		Messages: []anthropic.Message{
			anthropic.NewUserTextMessage(prompt),
		},
		MaxTokens: 600,
	}

	// Send request to Anthropic
	response, err := client.CreateMessages(ctx, request)
	if err != nil {
		log.Fatal(err)
	}

	// Print the response
	fmt.Println(response.Content[0].GetText())
}
