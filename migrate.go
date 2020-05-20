package jiraimporter

import (
	"fmt"
	"strconv"
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
		issueType := issue.IssueType

		switch issue.IssueType {
		case "Epic":
			// We can't convert existing types to Epics, so they become Stories instead
			issueType = "Story"
			break
		case "Sub-task":
			// Next-gen removes hyphen in Sub-task
			issueType = "Subtask"
			break
		}

		// Was this previously the a child of an Epic?
		if issue.EpicLink != "" {
			if issueType == "Subtask" {
				// For subtasks, add the issue as the child of the new Story(epic)
				updateData.Fields.Parent.Key = issue.EpicLink
			} else {
				// Only subtasks can be children. Create a 'Blocks' relationship
				issueLink := jira.IssueLink{
					Type: jira.IssueLinkType{
						Name: "Blocks",
					},
					InwardIssue: &jira.Issue{
						Key: issue.IssueKey,
					},
					OutwardIssue: &jira.Issue{
						Key: issue.EpicLink,
					},
				}
				_, err := ji.JiraClient.Issue.AddLink(&issueLink)
				if err != nil {
					fmt.Printf("Could not add 'Blocks' relationship to %s\n: %s", issue.IssueKey, err)
				}
			}
		}

		// Get the IssueType ID by name
		typeID, err := getIssueTypeID(issueType, project)
		if err != nil {
			fmt.Printf("Error getting type for issue %s: %s\n", issue.IssueKey, err)
			continue
		}
		updateData.Fields.IssueType.ID = typeID

		// Add existing subtasks back to their parent issues
		if issue.ParentID != "" {
			if parentKey := findParentIssueKey(issue.ParentID, issues); parentKey != "" {
				updateData.Fields.Parent.Key = parentKey
			}
		}

		// Import story points
		if issue.StoryPoints != "" {
			points, err := strconv.Atoi(issue.StoryPoints)
			if err == nil {
				updateData.Update.StoryPoints = append(updateData.Update.StoryPoints, struct {
					Set int `json:"set,omitempty"`
				}{
					Set: points,
				})
			}
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

		// Update the issue on Jira
		if err = ji.updateIssue(issue.IssueKey, &updateData); err != nil {
			fmt.Printf("Error updating issue %s: %s\n", issue.IssueKey, err)
		}
	}

	// TODO: Components become labels
}
