package prewarm

import (
	"encoding/json"
	"fmt"
	"log"
	"sort"
	"time"

	"github.com/ksheedlo/ghviz/github"
	"github.com/ksheedlo/ghviz/interfaces"
	"github.com/ksheedlo/ghviz/simulate"

	"github.com/jonboulle/clockwork"
)

func PrewarmHighScores(
	logger *log.Logger,
	gh github.ListAllPrEventser,
	redis interfaces.Rediser,
	clock clockwork.Clock,
	randomTagger interfaces.RandomTagger,
	owner, repo string,
) error {
	allPrEvents, httpErr := gh.ListAllPrEvents(logger, owner, repo)
	if httpErr != nil {
		return httpErr
	}
	sort.Sort(github.ByCreatedAt(allPrEvents))
	scoringEvents := simulate.ScoreIssues(allPrEvents, "ready for review")

	var members []interfaces.ZZ
	for _, event := range scoringEvents {
		timestamp := float64(event.Timestamp.Unix())
		// Ignore errors from json.Marshal because we control the serializing
		// routine for ScoringEvents. It should not be possible for an error
		// to occur here.
		jsonBlob, _ := json.Marshal(&event)
		members = append(members, interfaces.ZZ{
			Score:  timestamp,
			Member: jsonBlob,
		})
	}
	nextEventSetId, err := randomTagger.RandomTag()
	if err != nil {
		return err
	}
	eventSetCacheKey := fmt.Sprintf("gh:repos:%s:%s:issue_events:%s", owner, repo, nextEventSetId)
	if _, err := redis.ZAdd(eventSetCacheKey, members...); err != nil {
		return err
	}
	eventSetIdPtr := fmt.Sprintf("gh:repos:%s:%s:issue_event_setid", owner, repo)
	currentEventSetId, currentEventSetErr := redis.Get(eventSetIdPtr)
	if err := redis.Set(eventSetIdPtr, nextEventSetId, time.Duration(0)); err != nil {
		return err
	}
	if currentEventSetErr == nil && currentEventSetId != "" {
		clock.Sleep(5 * time.Second)
		if _, err := redis.Del(
			fmt.Sprintf("gh:repos:%s:%s:issue_events:%s", owner, repo, currentEventSetId),
		); err != nil {
			logger.Printf("ERROR: %s; recovered\n", err.Error())
		}
	}
	return nil
}
