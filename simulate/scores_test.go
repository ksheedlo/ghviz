package simulate

import (
	"encoding/json"
	"sort"
	"testing"
	"time"

	"github.com/ksheedlo/ghviz/github"
	"github.com/ksheedlo/ghviz/mocks"
	"github.com/stretchr/testify/assert"
)

func TestScoreUnlabeledReview(t *testing.T) {
	t.Parallel()

	scoringEvents := ScoreIssues(
		[]github.DetailedIssueEvent{
			github.DetailedIssueEvent{
				ActorId:     "tester1",
				CreatedAt:   time.Unix(1, 0),
				Detail:      nil,
				EventType:   github.IssueCreated,
				IssueNumber: 1,
			},
			github.DetailedIssueEvent{
				ActorId:     "tester1",
				CreatedAt:   time.Unix(2, 0),
				Detail:      map[string]interface{}{"name": "ready label"},
				EventType:   github.IssueLabeled,
				IssueNumber: 1,
			},
			github.DetailedIssueEvent{
				ActorId:     "tester2",
				CreatedAt:   time.Unix(3, 0),
				Detail:      map[string]interface{}{"name": "ready label"},
				EventType:   github.IssueUnlabeled,
				IssueNumber: 1,
			},
		},
		"ready label",
	)

	assert.Len(t, scoringEvents, 2)
	assert.Equal(t, scoringEvents[0].ActorId, "tester1")
	assert.Equal(t, scoringEvents[0].EventType, IssueOpened)
	assert.Equal(t, scoringEvents[1].ActorId, "tester2")
	assert.Equal(t, scoringEvents[1].EventType, IssueReviewed)
}

func TestScoreMergedReview(t *testing.T) {
	t.Parallel()

	scoringEvents := ScoreIssues(
		[]github.DetailedIssueEvent{
			github.DetailedIssueEvent{
				ActorId:     "tester1",
				CreatedAt:   time.Unix(1, 0),
				Detail:      nil,
				EventType:   github.IssueCreated,
				IssueNumber: 1,
			},
			github.DetailedIssueEvent{
				ActorId:     "tester1",
				CreatedAt:   time.Unix(2, 0),
				Detail:      map[string]interface{}{"name": "ready label"},
				EventType:   github.IssueLabeled,
				IssueNumber: 1,
			},
			github.DetailedIssueEvent{
				ActorId:     "tester2",
				CreatedAt:   time.Unix(3, 0),
				Detail:      nil,
				EventType:   github.IssueMerged,
				IssueNumber: 1,
			},
		},
		"ready label",
	)

	assert.Len(t, scoringEvents, 2)
	assert.Equal(t, scoringEvents[0].ActorId, "tester1")
	assert.Equal(t, scoringEvents[0].EventType, IssueOpened)
	assert.Equal(t, scoringEvents[1].ActorId, "tester2")
	assert.Equal(t, scoringEvents[1].EventType, IssueReviewed)
}

func TestScoreClosedBeforeReady(t *testing.T) {
	t.Parallel()

	scoringEvents := ScoreIssues(
		[]github.DetailedIssueEvent{
			github.DetailedIssueEvent{
				ActorId:     "tester1",
				CreatedAt:   time.Unix(1, 0),
				Detail:      nil,
				EventType:   github.IssueCreated,
				IssueNumber: 1,
			},
			github.DetailedIssueEvent{
				ActorId:     "tester2",
				CreatedAt:   time.Unix(3, 0),
				Detail:      nil,
				EventType:   github.IssueClosed,
				IssueNumber: 1,
			},
		},
		"ready label",
	)

	assert.Len(t, scoringEvents, 1)
	assert.Equal(t, scoringEvents[0].ActorId, "tester1")
	assert.Equal(t, scoringEvents[0].EventType, IssueOpened)
}

func TestScoreExtraneousLabel(t *testing.T) {
	t.Parallel()

	scoringEvents := ScoreIssues(
		[]github.DetailedIssueEvent{
			github.DetailedIssueEvent{
				ActorId:     "tester1",
				CreatedAt:   time.Unix(1, 0),
				Detail:      nil,
				EventType:   github.IssueCreated,
				IssueNumber: 1,
			},
			github.DetailedIssueEvent{
				ActorId:     "tester1",
				CreatedAt:   time.Unix(2, 0),
				Detail:      map[string]interface{}{"name": "something else"},
				EventType:   github.IssueLabeled,
				IssueNumber: 1,
			},
			github.DetailedIssueEvent{
				ActorId:     "tester2",
				CreatedAt:   time.Unix(3, 0),
				Detail:      map[string]interface{}{"name": "something else"},
				EventType:   github.IssueUnlabeled,
				IssueNumber: 1,
			},
		},
		"ready label",
	)

	assert.Len(t, scoringEvents, 1)
	assert.Equal(t, scoringEvents[0].ActorId, "tester1")
	assert.Equal(t, scoringEvents[0].EventType, IssueOpened)
}

func TestScoreMultipleIssues(t *testing.T) {
	t.Parallel()

	scoringEvents := ScoreIssues(
		[]github.DetailedIssueEvent{
			github.DetailedIssueEvent{
				ActorId:     "tester1",
				CreatedAt:   time.Unix(1, 0),
				Detail:      nil,
				EventType:   github.IssueCreated,
				IssueNumber: 1,
			},
			github.DetailedIssueEvent{
				ActorId:     "tester1",
				CreatedAt:   time.Unix(2, 0),
				Detail:      nil,
				EventType:   github.IssueCreated,
				IssueNumber: 2,
			},
			github.DetailedIssueEvent{
				ActorId:     "tester1",
				CreatedAt:   time.Unix(3, 0),
				Detail:      map[string]interface{}{"name": "ready label"},
				EventType:   github.IssueLabeled,
				IssueNumber: 1,
			},
			github.DetailedIssueEvent{
				ActorId:     "tester1",
				CreatedAt:   time.Unix(4, 0),
				Detail:      map[string]interface{}{"name": "ready label"},
				EventType:   github.IssueLabeled,
				IssueNumber: 2,
			},
			github.DetailedIssueEvent{
				ActorId:     "tester2",
				CreatedAt:   time.Unix(5, 0),
				Detail:      map[string]interface{}{"name": "ready label"},
				EventType:   github.IssueUnlabeled,
				IssueNumber: 1,
			},
			github.DetailedIssueEvent{
				ActorId:     "tester2",
				CreatedAt:   time.Unix(6, 0),
				Detail:      nil,
				EventType:   github.IssueClosed,
				IssueNumber: 2,
			},
		},
		"ready label",
	)

	assert.Len(t, scoringEvents, 4)
	assert.Equal(t, scoringEvents[0].ActorId, "tester1")
	assert.Equal(t, scoringEvents[0].EventType, IssueOpened)
	assert.Equal(t, scoringEvents[1].ActorId, "tester1")
	assert.Equal(t, scoringEvents[1].EventType, IssueOpened)
	assert.Equal(t, scoringEvents[2].ActorId, "tester2")
	assert.Equal(t, scoringEvents[2].EventType, IssueReviewed)
	assert.Equal(t, scoringEvents[3].ActorId, "tester2")
	assert.Equal(t, scoringEvents[3].EventType, IssueReviewed)
}

func TestScoreEvents(t *testing.T) {
	t.Parallel()

	scores := ScoreEvents([]ScoringEvent{
		ScoringEvent{ActorId: "Tester1", EventType: IssueOpened},
		ScoringEvent{ActorId: "Tester2", EventType: IssueReviewed},
	})
	sort.Sort(ByScore(scores))
	assert.Equal(t, scores[0].ActorId, "Tester1")
	assert.Equal(t, scores[0].Score, 200)
	assert.Equal(t, scores[1].ActorId, "Tester2")
	assert.Equal(t, scores[1].Score, 1000)
}

func TestMarshalScoringEvent(t *testing.T) {
	t.Parallel()

	sev := ScoringEvent{
		ActorId:   "tester1",
		EventType: IssueReviewed,
		Timestamp: time.Unix(1458966366, 892000000).UTC(),
	}
	jsonBytes := mocks.MarshalJSON(t, &sev)
	var sevMap map[string]interface{}
	assert.NoError(t, json.Unmarshal(jsonBytes, &sevMap))
	assert.Equal(t, "tester1", sevMap["actor_id"].(string))
	assert.Equal(t, "2016-03-26T04:26:06.892Z", sevMap["timestamp"].(string))
	var copySev ScoringEvent
	assert.NoError(t, json.Unmarshal(jsonBytes, &copySev))
	assert.Equal(t, "tester1", copySev.ActorId)
	assert.Equal(t, IssueReviewed, copySev.EventType)
	assert.Equal(t, sev.Timestamp, copySev.Timestamp)
}

func TestUnmarshalBadJSON(t *testing.T) {
	t.Parallel()

	var sev ScoringEvent
	assert.Error(t, json.Unmarshal([]byte(`{"actor_id":"FooBarso`), &sev))
}

const scoringEventBadTimestamp string = `{
	"actor_id": "FooBarson",
	"timestamp": "fish",
	"event_type": 0
}`

func TestUnmarshalBadTimestamp(t *testing.T) {
	t.Parallel()

	var sev ScoringEvent
	assert.Error(t, json.Unmarshal([]byte(scoringEventBadTimestamp), &sev))
}

func TestMarshalActorScore(t *testing.T) {
	t.Parallel()

	jsonBytes, err := json.Marshal(&ActorScore{
		ActorId: "panda99",
		Score:   1337,
	})
	assert.NoError(t, err)
	var actorScore map[string]interface{}
	assert.NoError(t, json.Unmarshal(jsonBytes, &actorScore))
	assert.Equal(t, "panda99", actorScore["actor_id"].(string))
	assert.Equal(t, 1337.0, actorScore["score"].(float64))
}
