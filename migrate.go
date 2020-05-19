package jiraimporter

import (
	"fmt"
	"time"

	"github.com/andygrunwald/go-jira"
)

const rateLimit = time.Second / 100 // 100 calls per second

// MigrateIssues is the entry point for the Jira import migration
func (ji *JiraImporter) MigrateIssues() {
	// Load the CSV export
	issues, err := parseCSV(ji.CSVPath)
	if err != nil {
		fmt.Println("Error parsing CSV file:", err)
		return
	}

	// Get Project info
	project, _, err := ji.JiraClient.Project.Get(issues[0].ProjectKey)
	if err != nil {
		fmt.Println("Error getting Project info:", err)
		return
	}

	// Handle rate limiting
	throttle := time.Tick(rateLimit)

	// Convert issues to correct types
	for _, issue := range issues {
		<-throttle
		updateData := issueUpdateData{}
		issuetype := issue.IssueType

		switch issue.IssueType {
		case "Epic":
			// We can't convert existing types to Epics, so they become Stories instead
			issuetype = "Story"
			break
		case "Sub-task":
			// Next-gen removes hyphen in Sub-task
			issuetype = "Subtask"
			break
		}

		// If this was previously the a child of an Epic, make it a Subtask
		// so we can make it a child of the new Story type
		if issue.EpicLink != "" {
			issuetype = "Subtask"
		}

		// Get the IssueType ID by name
		typeID, err := getIssueTypeID(issuetype, project)
		if err != nil {
			fmt.Printf("Error getting type for issue %s: %s\n", issue.IssueKey, err)
			continue
		}
		updateData.Fields.Issuetype.ID = typeID

		// Import story points
		if issue.StoryPoints != "" {
			updateData.Update.StoryPoints[0].Set = issue.StoryPoints
		}

		// Update the issue on Jira
		if err = ji.updateIssue(issue.IssueKey, &updateData); err != nil {
			fmt.Printf("Error updating issue %s: %s\n", issue.IssueKey, err)
			continue
		}
	}

	// Now that Issue types are correct, we can set relationships
	for _, issue := range issues {
		<-throttle
		updateData := issueUpdateData{}

		// Add children back to Epics
		if issue.EpicLink != "" {

		}

		// Add subtasks back to their parent issues
		if issue.ParentID != "" {

		}

		// Fix missing Issue Split relationships
		if issue.IssueSplit != "" {
			issueLink := jira.IssueLink{
				Type: jira.IssueLinkType{
					Name: "Issue split",
				},
				InwardIssue: &jira.Issue{
					Key: issue.IssueKey,
				},
				OutwardIssue: &jira.Issue{
					Key: issue.IssueSplit,
				},
			}
			_, err := ji.JiraClient.Issue.AddLink(&issueLink)
			if err != nil {
				fmt.Printf("Could not add 'Splits' relationship to %s\n: %s", issue.IssueKey, err)
			}
		}

	}

	// Epics become stories and subtasks are relationships

	// Handle subtasks by adding them to correct parents

	// Components become labels
}
