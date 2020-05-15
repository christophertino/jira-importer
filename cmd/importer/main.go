package main

import (
	"bufio"
	"fmt"
	"log"
	"os"

	"github.com/andygrunwald/go-jira"
	jiraimporter "github.com/christophertino/jira-importer"
	"github.com/joho/godotenv"
)

var (
	jiraClient *jira.Client
	err        error
)

func init() {
	// Load local env file
	if err = godotenv.Load(); err != nil {
		log.Fatalf("Error loading .env file: %s", err)
	}

	// Create Jira client
	if jiraClient == nil {
		jiraAuth := jira.BasicAuthTransport{
			Username: os.Getenv(("jira_email")),
			Password: os.Getenv(("jira_token")),
		}
		jiraClient, err = jira.NewClient(jiraAuth.Client(), "https://ghostery.atlassian.net/")
		if err != nil {
			fmt.Println("Error creating Jira client:", err)
			os.Exit(1)
		}
	}
}

func main() {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("--- Jira Next-Gen Importer ---")
	fmt.Println("Enter path to Jira issue export CSV:")
	fmt.Print("-> ")

	csvPath, err := reader.ReadString('\n')
	if err != nil {
		fmt.Println("Error reading CSV path:", err)
		os.Exit(1)
	}

	// Build the JiraImporter config
	ji := jiraimporter.JiraImporter{
		CSVPath:    csvPath,
		JiraClient: jiraClient,
	}

	ji.MigrateIssues()
}
