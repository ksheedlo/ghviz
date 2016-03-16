package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sort"
	"strconv"
	"time"

	"github.com/ksheedlo/ghviz/github"
	"github.com/ksheedlo/ghviz/interfaces"
	"github.com/ksheedlo/ghviz/simulate"
	"gopkg.in/redis.v3"
)

func withDefaultStr(config, default_ string) string {
	if config == "" {
		return default_
	}
	return config
}

func main() {
	var redisClient interfaces.Rediser
	if redisHost := os.Getenv("GHVIZ_REDIS_HOST"); redisHost != "" {
		redisPort := withDefaultStr(os.Getenv("GHVIZ_REDIS_PORT"), "6379")
		redisClient = interfaces.NewGoRedis(redis.NewClient(&redis.Options{
			Addr:     fmt.Sprintf("%s:%s", redisHost, redisPort),
			Password: os.Getenv("GHVIZ_REDIS_PASSWORD"),
			DB:       0,
		}))
	}

	gh := github.NewClient(&github.Options{
		RedisClient: redisClient,
		Token:       os.Getenv("GITHUB_TOKEN"),
	})
	logger := log.New(os.Stderr, "", log.Ldate|log.Ltime|log.Lshortfile|log.LUTC)
	owner := os.Getenv("GHVIZ_OWNER")
	repo := os.Getenv("GHVIZ_REPO")
	eventListCacheKey := fmt.Sprintf("gh:repos:%s:%s:issue_events", owner, repo)
	allPrEvents, err := gh.ListAllPrEvents(logger, owner, repo)
	if err != nil {
		logger.Fatal(err)
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
	if _, err := redisClient.ZAdd(eventListCacheKey, members...); err != nil {
		logger.Fatal(err)
	}
	startDate := strconv.FormatInt(
		time.Date(2016, time.February, 1, 0, 0, 0, 0, time.UTC).Unix(),
		10,
	)
	endDate := strconv.FormatInt(
		time.Date(2016, time.March, 1, 0, 0, 0, 0, time.UTC).Unix(),
		10,
	)
	scoringEventJsons, redisErr := redisClient.ZRangeByScore(
		eventListCacheKey,
		&interfaces.ZRangeByScoreOpts{Min: startDate, Max: endDate},
	)
	if redisErr != nil {
		logger.Fatal(redisErr)
	}
	var eventsToScore []simulate.ScoringEvent
	for _, scoringEventJson := range scoringEventJsons {
		scoringEvent := simulate.ScoringEvent{}
		if jsonErr := json.Unmarshal([]byte(scoringEventJson), &scoringEvent); jsonErr != nil {
			logger.Fatal(jsonErr)
		}
		eventsToScore = append(eventsToScore, scoringEvent)
	}
	highScores := simulate.ScoreEvents(eventsToScore)
	sort.Sort(sort.Reverse(simulate.ByScore(highScores)))
	for _, highScore := range highScores {
		fmt.Printf("%d %s\n", highScore.Score, highScore.ActorId)
	}
}
