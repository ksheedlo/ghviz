package simulate

import (
	"encoding/json"
	"time"

	"github.com/ksheedlo/ghviz/github"
)

type ScoringEventType int

const (
	IssueOpened ScoringEventType = iota
	IssueReviewed
)

var scoringEventTypesByKey map[string]ScoringEventType = map[string]ScoringEventType{
	"opened":   IssueOpened,
	"reviewed": IssueReviewed,
}

var scoringEventKeysByType map[ScoringEventType]string = (func(
	types map[string]ScoringEventType,
) map[ScoringEventType]string {
	invertedMap := make(map[ScoringEventType]string)
	for key, value := range types {
		invertedMap[value] = key
	}
	return invertedMap
})(scoringEventTypesByKey)

type ScoringEvent struct {
	ActorId   string
	EventType ScoringEventType
	Timestamp time.Time
}

func (sev *ScoringEvent) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		"actor_id":   sev.ActorId,
		"event_type": scoringEventKeysByType[sev.EventType],
		"timestamp":  sev.Timestamp,
	})
}

func (sev *ScoringEvent) UnmarshalJSON(bytes []byte) error {
	item := make(map[string]interface{})
	if err := json.Unmarshal(bytes, &item); err != nil {
		return err
	}
	timestamp, err := time.Parse(time.RFC3339, item["timestamp"].(string))
	if err != nil {
		return err
	}
	sev.ActorId = item["actor_id"].(string)
	sev.EventType = scoringEventTypesByKey[item["event_type"].(string)]
	sev.Timestamp = timestamp
	return nil
}

type ByTimestamp []ScoringEvent

func (a ByTimestamp) Len() int           { return len(a) }
func (a ByTimestamp) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByTimestamp) Less(i, j int) bool { return a[i].Timestamp.Before(a[j].Timestamp) }

type ActorScore struct {
	ActorId string
	Score   int
}

type ByScore []ActorScore

func (a ByScore) Len() int           { return len(a) }
func (a ByScore) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByScore) Less(i, j int) bool { return a[i].Score < a[j].Score }

func (acs *ActorScore) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		"actor_id": acs.ActorId,
		"score":    acs.Score,
	})
}

type PrState int

const (
	PrStateSubmitted PrState = iota
	PrStateReady
	PrStateReviewed
)

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

func ScoreEvents(scoringEvents []ScoringEvent) []ActorScore {
	scoreMap := make(map[string]int)
	for _, event := range scoringEvents {
		switch event.EventType {
		case IssueOpened:
			scoreMap[event.ActorId] = scoreMap[event.ActorId] + 200
		case IssueReviewed:
			scoreMap[event.ActorId] = scoreMap[event.ActorId] + 1000
		}
	}
	var scores []ActorScore
	for actorId, score := range scoreMap {
		scores = append(scores, ActorScore{ActorId: actorId, Score: score})
	}
	return scores
}
