package models

import (
	"sort"
	"time"

	"github.com/ksheedlo/ghviz/github"
)

type IssueEventType int

const (
	IssueOpened IssueEventType = iota
	IssueClosed
)

type IssueEvent struct {
	EventType IssueEventType
	IsPr      bool
	Timestamp time.Time
}

type byTimestamp []IssueEvent

func (a byTimestamp) Len() int           { return len(a) }
func (a byTimestamp) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a byTimestamp) Less(i, j int) bool { return a[i].Timestamp.Before(a[j].Timestamp) }

func IssueEventsFromApi(issues []github.Issue) []IssueEvent {
	var issueEvents []IssueEvent

	for _, issue := range issues {
		issueEvents = append(issueEvents, IssueEvent{
			EventType: IssueOpened,
			IsPr:      issue.IsPr,
			Timestamp: issue.CreatedAt,
		})

		if issue.IsClosed {
			issueEvents = append(issueEvents, IssueEvent{
				EventType: IssueClosed,
				IsPr:      issue.IsPr,
				Timestamp: issue.ClosedAt,
			})
		}
	}
	sort.Sort(byTimestamp(issueEvents))
	return issueEvents
}
