package main

import (
	"fmt"
	"log"
	"os"

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
	issues, err := gh.ListIssues(logger, os.Getenv("GHVIZ_OWNER"), os.Getenv("GHVIZ_REPO"))
	if err != nil {
		logger.Fatal(err)
	}
	c := make(chan map[string]int)
	numPrs := 0
	for i, issue := range issues {
		if issue.IsPr {
			numPrs++
			go func(ii int) {
				logger.Printf("Fetching events for PR #%d", issues[ii].Number)
				issueEvents, err := gh.ListIssueEvents(logger, &issues[ii])
				if err != nil {
					logger.Printf(
						"ERROR: %s; issue %d will not contribute to scoring.",
						err.Error(),
						issues[ii].Number,
					)
					c <- make(map[string]int)
					return
				}
				scoringEvents := simulate.ScoreIssues(issueEvents, "ready for review")
				issueScores := simulate.ScoreEvents(scoringEvents)
				c <- issueScores
			}(i)
		}
	}
	totalScores := make(map[string]int)
	for i := 0; i < numPrs; i++ {
		issueScores := <-c
		for userId, score := range issueScores {
			totalScores[userId] = totalScores[userId] + score
		}
	}
	for userId, score := range totalScores {
		fmt.Printf("%d %s\n", score, userId)
	}
}
