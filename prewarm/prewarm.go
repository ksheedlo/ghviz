package prewarm

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"sort"
	"time"

	"github.com/ksheedlo/ghviz/github"
	"github.com/ksheedlo/ghviz/interfaces"
	"github.com/ksheedlo/ghviz/simulate"
)

func PrewarmHighScores(
	logger *log.Logger,
	gh *github.Client,
	redis interfaces.Rediser,
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
		if jsonBlob, jsonErr := json.Marshal(&event); jsonErr != nil {
			logger.Printf(
				"ERROR: %s; a scoring event will be dropped.",
				jsonErr.Error(),
			)
		} else {
			members = append(members, interfaces.ZZ{
				Score:  timestamp,
				Member: jsonBlob,
			})
		}
	}
	nextEventSetIdInt, err := rand.Int(rand.Reader, big.NewInt(1<<62))
	if err != nil {
		return err
	}
	nextEventSetId := nextEventSetIdInt.Text(36)
	eventSetCacheKey := fmt.Sprintf("gh:repos:%s:%s:issue_events:%s", owner, repo, nextEventSetId)
	if _, err := redis.ZAdd(eventSetCacheKey, members...); err != nil {
		return err
	}
	eventSetIdPtr := fmt.Sprintf("gh:repos:%s:%s:issue_event_setid", owner, repo)
	currentEventSetId, currentEventSetErr := redis.Get(eventSetIdPtr)
	if err := redis.Set(eventSetIdPtr, nextEventSetId, time.Duration(0)); err != nil {
		return err
	}
	if currentEventSetErr != nil && currentEventSetId != "" {
		time.Sleep(5 * time.Second)
		if _, err := redis.Del(
			fmt.Sprintf("gh:repos:%s:%s:issue_events:%s", owner, repo, currentEventSetId),
		); err != nil {
			logger.Printf("ERROR: %s; recovered\n", err.Error())
		}
	}
	return nil
}
