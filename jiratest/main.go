package main

import (
	"fmt"
	"github.com/andygrunwald/go-jira"
)

func main() {
	jiraClient, err := jira.NewClient(nil, "https://issues.redhat.com/")
	if err != nil {
		fmt.Printf("error: %v+\n", err)
		return
	}

	issue, _, err := jiraClient.Issue.Get("OSDEV-244", nil)
	if err != nil {
		fmt.Printf("error: %v+\n", err)
		return
	}


	fmt.Printf("%s: %+v\n", issue.Key, issue.Fields.Summary)

	fmt.Printf("%s: %+v\n", issue.Key, issue.Fields.Summary)
	fmt.Printf("Type: %s\n", issue.Fields.Type.Name)
	fmt.Printf("Priority: %s\n", issue.Fields.Priority.Name)

	// MESOS-3325: Running mesos-slave@0.23 in a container causes slave to be lost after a restart
	// Type: Bug
	// Priority: Critical
}
