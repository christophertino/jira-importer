package main

import (
	"fmt"
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
		fmt.Println("Error loading .env file:", err)
		os.Exit(1)
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
	fmt.Println("--- Jira Next-Gen Importer ---")

	if len(os.Args) <= 1 {
		fmt.Println("Please enter path to Jira export CSV")
		os.Exit(1)
	}

	csvPath := os.Args[1]

	// Build the JiraImporter config
	ji := jiraimporter.JiraImporter{
		CSVPath:    csvPath,
		JiraClient: jiraClient,
	}

	ji.MigrateIssues()
}
