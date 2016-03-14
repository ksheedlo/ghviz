package simulate

import (
	"time"

	"github.com/ksheedlo/ghviz/github"
)

type ScoringEventType int

const (
	IssueOpened ScoringEventType = iota
	IssueReviewed
)

type PrState int

const (
	PrStateSubmitted PrState = iota
	PrStateReady
	PrStateReviewed
)

type ScoringEvent struct {
	ActorId   string
	EventType ScoringEventType
	Timestamp time.Time
}

func ScoreIssues(issueEvents []github.DetailedIssueEvent, readyLabel string) []ScoringEvent {
	var scoringEvents []ScoringEvent
	prStates := make(map[int]PrState)
	for _, event := range issueEvents {
		if _, hasIssueState := prStates[event.IssueNumber]; !hasIssueState {
			prStates[event.IssueNumber] = PrStateSubmitted
		}
		switch event.EventType {
		case github.IssueCreated:
			// 1, Creating the issue counts as a submission.
			scoringEvents = append(scoringEvents, ScoringEvent{
				ActorId:   event.ActorId,
				EventType: IssueOpened,
				Timestamp: event.CreatedAt,
			})
		case github.IssueLabeled:
			// 2. The submitter should apply the ready label when the PR is ready
			//    for review.
			labelName := (event.Detail.(map[string]interface{}))["name"].(string)
			if labelName == readyLabel {
				prStates[event.IssueNumber] = PrStateReady
			}
		case github.IssueUnlabeled:
			// 3. When a reviewer removes the ready label from a PR in the ready
			//    state, that constitutes a review.
			labelName := (event.Detail.(map[string]interface{}))["name"].(string)
			if labelName == readyLabel && prStates[event.IssueNumber] == PrStateReady {
				prStates[event.IssueNumber] = PrStateReviewed
				scoringEvents = append(scoringEvents, ScoringEvent{
					ActorId:   event.ActorId,
					EventType: IssueReviewed,
					Timestamp: event.CreatedAt,
				})
			}
		case github.IssueClosed, github.IssueMerged:
			// 4. If a reviewer merges a PR from the ready state, that also
			//    constitutes a review. This is a shorthand version of case 3.
			if prStates[event.IssueNumber] == PrStateReady {
				prStates[event.IssueNumber] = PrStateReviewed
				scoringEvents = append(scoringEvents, ScoringEvent{
					ActorId:   event.ActorId,
					EventType: IssueReviewed,
					Timestamp: event.CreatedAt,
				})
			}
		}
	}
	return scoringEvents
}

func ScoreEvents(scoringEvents []ScoringEvent) map[string]int {
	scores := make(map[string]int)
	for _, event := range scoringEvents {
		switch event.EventType {
		case IssueOpened:
			scores[event.ActorId] = scores[event.ActorId] + 200
		case IssueReviewed:
			scores[event.ActorId] = scores[event.ActorId] + 1000
		}
	}
	return scores
}
