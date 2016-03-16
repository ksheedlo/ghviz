package simulate

import (
	"sort"
	"testing"
	"time"

	"github.com/ksheedlo/ghviz/github"
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
