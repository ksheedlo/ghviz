package simulate

import (
	"testing"
	"time"

	"github.com/ksheedlo/ghviz/github"
	"github.com/stretchr/testify/assert"
)

func TestScoreUnlabeledReview(t *testing.T) {
	t.Parallel()

	scoringEvents := ScoreIssue(
		[]github.DetailedIssueEvent{
			github.DetailedIssueEvent{
				ActorId:   "tester1",
				CreatedAt: time.Unix(1, 0),
				Detail:    nil,
				EventType: github.IssueCreated,
			},
			github.DetailedIssueEvent{
				ActorId:   "tester1",
				CreatedAt: time.Unix(2, 0),
				Detail:    map[string]interface{}{"name": "ready label"},
				EventType: github.IssueLabeled,
			},
			github.DetailedIssueEvent{
				ActorId:   "tester2",
				CreatedAt: time.Unix(3, 0),
				Detail:    map[string]interface{}{"name": "ready label"},
				EventType: github.IssueUnlabeled,
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

	scoringEvents := ScoreIssue(
		[]github.DetailedIssueEvent{
			github.DetailedIssueEvent{
				ActorId:   "tester1",
				CreatedAt: time.Unix(1, 0),
				Detail:    nil,
				EventType: github.IssueCreated,
			},
			github.DetailedIssueEvent{
				ActorId:   "tester1",
				CreatedAt: time.Unix(2, 0),
				Detail:    map[string]interface{}{"name": "ready label"},
				EventType: github.IssueLabeled,
			},
			github.DetailedIssueEvent{
				ActorId:   "tester2",
				CreatedAt: time.Unix(3, 0),
				Detail:    nil,
				EventType: github.IssueMerged,
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

	scoringEvents := ScoreIssue(
		[]github.DetailedIssueEvent{
			github.DetailedIssueEvent{
				ActorId:   "tester1",
				CreatedAt: time.Unix(1, 0),
				Detail:    nil,
				EventType: github.IssueCreated,
			},
			github.DetailedIssueEvent{
				ActorId:   "tester2",
				CreatedAt: time.Unix(3, 0),
				Detail:    nil,
				EventType: github.IssueClosed,
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

	scoringEvents := ScoreIssue(
		[]github.DetailedIssueEvent{
			github.DetailedIssueEvent{
				ActorId:   "tester1",
				CreatedAt: time.Unix(1, 0),
				Detail:    nil,
				EventType: github.IssueCreated,
			},
			github.DetailedIssueEvent{
				ActorId:   "tester1",
				CreatedAt: time.Unix(2, 0),
				Detail:    map[string]interface{}{"name": "something else"},
				EventType: github.IssueLabeled,
			},
			github.DetailedIssueEvent{
				ActorId:   "tester2",
				CreatedAt: time.Unix(3, 0),
				Detail:    map[string]interface{}{"name": "something else"},
				EventType: github.IssueUnlabeled,
			},
		},
		"ready label",
	)

	assert.Len(t, scoringEvents, 1)
	assert.Equal(t, scoringEvents[0].ActorId, "tester1")
	assert.Equal(t, scoringEvents[0].EventType, IssueOpened)
}

func TestScoreEvents(t *testing.T) {
	t.Parallel()

	scores := ScoreEvents([]ScoringEvent{
		ScoringEvent{ActorId: "Tester1", EventType: IssueOpened},
		ScoringEvent{ActorId: "Tester2", EventType: IssueReviewed},
	})
	assert.Equal(t, scores["Tester1"], 200)
	assert.Equal(t, scores["Tester2"], 1000)
}
